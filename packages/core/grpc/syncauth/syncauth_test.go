package syncauth

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const secret = "0123456789abcdef0123456789abcdef-test"

func mint(t *testing.T, c Claims, ttl time.Duration, now time.Time) string {
	t.Helper()
	tok, err := IssueToken(secret, c, ttl, now)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func TestVerifyAcceptsValidToken(t *testing.T) {
	v, _ := NewVerifier(secret, nil)
	c, err := v.Verify(mint(t, Claims{TenantSlug: "acme", StoreID: 7, DeviceID: "pos-1"}, time.Hour, time.Now()))
	if err != nil {
		t.Fatal(err)
	}
	if c.TenantSlug != "acme" || c.StoreID != 7 || c.DeviceID != "pos-1" {
		t.Fatalf("unexpected claims %+v", c)
	}
}

func TestVerifyRejects(t *testing.T) {
	v, _ := NewVerifier(secret, []string{"pos-revoked"})
	now := time.Now()
	userJWT, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"tenant_slug": "acme", "sub": "pos-1", "exp": now.Add(time.Hour).Unix(),
	}).SignedString([]byte(secret))
	other, _ := IssueToken(strings.Repeat("x", 40), Claims{TenantSlug: "acme", DeviceID: "d"}, time.Hour, now)
	cases := map[string]string{
		"empty":          "",
		"garbage":        "not-a-jwt",
		"expired":        mint(t, Claims{TenantSlug: "acme", DeviceID: "d"}, time.Minute, now.Add(-time.Hour)),
		"wrong secret":   other,
		"no audience":    userJWT,
		"revoked device": mint(t, Claims{TenantSlug: "acme", DeviceID: "pos-revoked"}, time.Hour, now),
	}
	for name, tok := range cases {
		if _, err := v.Verify(tok); err == nil {
			t.Errorf("%s: expected rejection", name)
		}
	}
}

func TestVerifyRejectsAlgNone(t *testing.T) {
	v, _ := NewVerifier(secret, nil)
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodNone, tokenClaims{
		TenantSlug: "acme",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: Issuer, Subject: "d", Audience: jwt.ClaimStrings{Audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, err := v.Verify(tok); err == nil {
		t.Fatal("alg=none must be rejected")
	}
}

func TestShortSecretRejected(t *testing.T) {
	if _, err := NewVerifier("short", nil); err == nil {
		t.Fatal("short secret must fail")
	}
	if _, err := IssueToken("short", Claims{TenantSlug: "a", DeviceID: "d"}, time.Hour, time.Now()); err == nil {
		t.Fatal("short secret must fail to issue")
	}
}

func TestAuthorize(t *testing.T) {
	c := &Claims{TenantSlug: "acme", StoreID: 7}
	if err := Authorize(c, "acme", 7); err != nil {
		t.Fatal(err)
	}
	if status.Code(Authorize(c, "other", 7)) != codes.PermissionDenied {
		t.Fatal("cross-tenant must be denied")
	}
	if status.Code(Authorize(c, "acme", 8)) != codes.PermissionDenied {
		t.Fatal("cross-store must be denied")
	}
	if status.Code(Authorize(nil, "acme", 7)) != codes.Unauthenticated {
		t.Fatal("missing claims must be unauthenticated")
	}
	if err := Authorize(&Claims{TenantSlug: "acme"}, "acme", 99); err != nil {
		t.Fatal("tenant-wide token may use any store of the tenant")
	}
}

type probeServer struct {
	syncpb.UnimplementedSyncServiceServer
	gotClaims *Claims
}

func (p *probeServer) StreamPull(_ *syncpb.PullRequest, s syncpb.SyncService_StreamPullServer) error {
	p.gotClaims, _ = FromContext(s.Context())
	return nil
}

func TestInterceptorEndToEnd(t *testing.T) {
	v, _ := NewVerifier(secret, nil)
	lis := bufconn.Listen(1 << 16)
	srv := grpc.NewServer(grpc.ChainStreamInterceptor(StreamInterceptor(v)))
	probe := &probeServer{}
	syncpb.RegisterSyncServiceServer(srv, probe)
	go srv.Serve(lis)
	defer srv.Stop()
	dial := func(opts ...grpc.DialOption) syncpb.SyncServiceClient {
		opts = append(opts, grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		conn, err := grpc.NewClient("passthrough:///bufnet", opts...)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.Close() })
		return syncpb.NewSyncServiceClient(conn)
	}
	pull := func(c syncpb.SyncServiceClient) error {
		s, err := c.StreamPull(context.Background(), &syncpb.PullRequest{TenantSlug: "acme"})
		if err != nil {
			return err
		}
		_, err = s.Recv()
		return err
	}

	if code := status.Code(pull(dial())); code != codes.Unauthenticated {
		t.Fatalf("no token: got %v", code)
	}
	bad := TokenCredentials{Token: "bogus", AllowInsecure: true}
	if code := status.Code(pull(dial(grpc.WithPerRPCCredentials(bad)))); code != codes.Unauthenticated {
		t.Fatalf("bad token: got %v", code)
	}
	good := TokenCredentials{Token: mint(t, Claims{TenantSlug: "acme", StoreID: 3, DeviceID: "pos-1"}, time.Hour, time.Now()), AllowInsecure: true}
	if err := pull(dial(grpc.WithPerRPCCredentials(good))); err == nil || err.Error() != "EOF" {
		t.Fatalf("good token should reach handler and end cleanly, got %v", err)
	}
	if probe.gotClaims == nil || probe.gotClaims.TenantSlug != "acme" || probe.gotClaims.StoreID != 3 {
		t.Fatalf("handler did not receive verified claims: %+v", probe.gotClaims)
	}
}

func TestTokenRequiresTLSUnlessAllowed(t *testing.T) {
	if !(TokenCredentials{Token: "x"}).RequireTransportSecurity() {
		t.Fatal("token must require TLS by default")
	}
	if (TokenCredentials{Token: "x", AllowInsecure: true}).RequireTransportSecurity() {
		t.Fatal("AllowInsecure should relax requirement")
	}
}
