// Package syncauth authenticates POS devices on the gRPC sync service.
//
// A device is issued a signed token bound to one tenant (and optionally one
// store). The server derives tenant/store authority from the verified token
// instead of trusting the fields inside each SyncEvent.
//
// Tokens are HS256 JWTs with a dedicated audience, so a user/session JWT issued
// by the REST API can never be used as a sync credential.
package syncauth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// Audience is the JWT audience reserved for sync device tokens.
	Audience = "nembus-sync"
	// Issuer identifies tokens minted by the cloud tooling.
	Issuer = "nembus-cloud"
	// MinSecretLen is the minimum accepted signing secret length in bytes.
	MinSecretLen = 32
	// SyncServicePrefix is the gRPC full-method prefix that requires auth.
	SyncServicePrefix = "/sync.SyncService/"

	metadataKey = "authorization"
)

// Claims is the verified identity of a sync device.
type Claims struct {
	TenantSlug string `json:"tenant_slug"`
	StoreID    int32  `json:"store_id,omitempty"` // 0 = not bound to a single store
	DeviceID   string `json:"-"`                  // JWT subject
}

type tokenClaims struct {
	TenantSlug string `json:"tenant_slug"`
	StoreID    int32  `json:"store_id,omitempty"`
	jwt.RegisteredClaims
}

// IssueToken mints a device token. ttl must be positive.
func IssueToken(secret string, c Claims, ttl time.Duration, now time.Time) (string, error) {
	if len(secret) < MinSecretLen {
		return "", fmt.Errorf("sync signing secret must be at least %d bytes", MinSecretLen)
	}
	if strings.TrimSpace(c.TenantSlug) == "" {
		return "", errors.New("tenant slug is required")
	}
	if strings.TrimSpace(c.DeviceID) == "" {
		return "", errors.New("device id is required")
	}
	if ttl <= 0 {
		return "", errors.New("ttl must be positive")
	}
	tc := tokenClaims{
		TenantSlug: c.TenantSlug,
		StoreID:    c.StoreID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			Subject:   c.DeviceID,
			Audience:  jwt.ClaimStrings{Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, tc).SignedString([]byte(secret))
}

// Verifier validates device tokens.
type Verifier struct {
	secret  []byte
	revoked map[string]struct{}
	now     func() time.Time
}

// NewVerifier builds a verifier. revokedDevices lists device ids (token
// subjects) that are rejected even if their token is otherwise valid.
func NewVerifier(secret string, revokedDevices []string) (*Verifier, error) {
	if len(secret) < MinSecretLen {
		return nil, fmt.Errorf("sync signing secret must be at least %d bytes", MinSecretLen)
	}
	v := &Verifier{secret: []byte(secret), revoked: map[string]struct{}{}, now: time.Now}
	for _, d := range revokedDevices {
		if d = strings.TrimSpace(d); d != "" {
			v.revoked[d] = struct{}{}
		}
	}
	return v, nil
}

// Verify parses and validates a raw token string.
func (v *Verifier) Verify(raw string) (*Claims, error) {
	raw = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "Bearer "))
	if raw == "" {
		return nil, errors.New("missing token")
	}
	var tc tokenClaims
	_, err := jwt.ParseWithClaims(raw, &tc, func(*jwt.Token) (any, error) { return v.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithAudience(Audience),
		jwt.WithIssuer(Issuer),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(time.Minute),
		jwt.WithTimeFunc(v.now),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	if strings.TrimSpace(tc.TenantSlug) == "" || strings.TrimSpace(tc.Subject) == "" {
		return nil, errors.New("token lacks tenant or device")
	}
	if _, bad := v.revoked[tc.Subject]; bad {
		return nil, errors.New("device is revoked")
	}
	return &Claims{TenantSlug: tc.TenantSlug, StoreID: tc.StoreID, DeviceID: tc.Subject}, nil
}

type ctxKey struct{}

// FromContext returns the verified claims placed by the interceptors.
func FromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(*Claims)
	return c, ok && c != nil
}

// WithClaims attaches claims to a context (used by interceptors and tests).
func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

// Authorize checks a requested tenant/store against verified claims.
// Empty requested values are not allowed to bypass the check for tenant.
func Authorize(c *Claims, tenant string, store int32) error {
	if c == nil {
		return status.Error(codes.Unauthenticated, "authentication required")
	}
	if tenant != c.TenantSlug {
		return status.Error(codes.PermissionDenied, "tenant not permitted for this device")
	}
	if c.StoreID != 0 && store != c.StoreID {
		return status.Error(codes.PermissionDenied, "store not permitted for this device")
	}
	return nil
}

func authenticate(ctx context.Context, v *Verifier) (*Claims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get(metadataKey)) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization metadata")
	}
	claims, err := v.Verify(md.Get(metadataKey)[0])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return claims, nil
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context { return w.ctx }

// StreamInterceptor authenticates streams of the sync service. Other services
// on the same server (e.g. backup) pass through unchanged.
func StreamInterceptor(v *Verifier) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, h grpc.StreamHandler) error {
		if !strings.HasPrefix(info.FullMethod, SyncServicePrefix) {
			return h(srv, ss)
		}
		claims, err := authenticate(ss.Context(), v)
		if err != nil {
			return err
		}
		return h(srv, &wrappedStream{ServerStream: ss, ctx: WithClaims(ss.Context(), claims)})
	}
}

// UnaryInterceptor authenticates unary calls of the sync service.
func UnaryInterceptor(v *Verifier) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
		if !strings.HasPrefix(info.FullMethod, SyncServicePrefix) {
			return h(ctx, req)
		}
		claims, err := authenticate(ctx, v)
		if err != nil {
			return nil, err
		}
		return h(WithClaims(ctx, claims), req)
	}
}

// ServerTLS loads a server TLS credential (TLS 1.2+).
func ServerTLS(certFile, keyFile string) (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load TLS key pair: %w", err)
	}
	return credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}), nil
}

// ClientTLS builds a client TLS credential. caFile may be empty to use the
// system roots; serverName may be empty to use the dialed host.
func ClientTLS(caFile, serverName string) (credentials.TransportCredentials, error) {
	cfg := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: serverName}
	if caFile != "" {
		pem, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("read CA file: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, errors.New("CA file contains no valid certificates")
		}
		cfg.RootCAs = pool
	}
	return credentials.NewTLS(cfg), nil
}

// TokenCredentials attaches the device token to every RPC.
type TokenCredentials struct {
	Token string
	// AllowInsecure permits sending the token without transport security.
	// Only for loopback development.
	AllowInsecure bool
}

func (t TokenCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{metadataKey: "Bearer " + t.Token}, nil
}

func (t TokenCredentials) RequireTransportSecurity() bool { return !t.AllowInsecure }
