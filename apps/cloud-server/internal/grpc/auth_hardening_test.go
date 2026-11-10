package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestBindStreamTenant(t *testing.T) {
	s := &SyncServer{
		requireTokenAuth: false, // test binding logic independent of DB auth
	}

	ctx := context.Background()
	var bound string

	// 1. Initial bind with empty tenant should fail
	err := s.bindStreamTenant(ctx, &bound, "", "dev-1", 1)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for empty initial tenant, got %v", err)
	}

	// 2. Initial bind with tenant-a should succeed
	err = s.bindStreamTenant(ctx, &bound, "tenant-a", "dev-1", 1)
	if err != nil {
		t.Fatalf("unexpected error binding tenant-a: %v", err)
	}
	if bound != "tenant-a" {
		t.Fatalf("expected bound to be 'tenant-a', got %q", bound)
	}

	// 3. Subsequent event for the same tenant should succeed
	err = s.bindStreamTenant(ctx, &bound, "tenant-a", "dev-1", 1)
	if err != nil {
		t.Fatalf("unexpected error re-verifying same tenant: %v", err)
	}

	// 4. Subsequent event with different tenant MUST be rejected with PermissionDenied
	err = s.bindStreamTenant(ctx, &bound, "tenant-b", "dev-1", 1)
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied when switching tenant to tenant-b, got %v", err)
	}

	// 5. Subsequent event with empty tenant should fail
	err = s.bindStreamTenant(ctx, &bound, "", "dev-1", 1)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for empty tenant on bound stream, got %v", err)
	}
}

func TestProcessPushBatchRejectsMismatchedTenants(t *testing.T) {
	s := &SyncServer{}
	ctx := context.Background()

	events := []*syncpb.SyncEvent{
		{
			Id:         1,
			BatchId:    "batch-1",
			TenantSlug: "tenant-a",
			EntityType: "pos_transaction",
			EntityId:   101,
		},
		{
			Id:         2,
			BatchId:    "batch-1",
			TenantSlug: "tenant-b", // mismatched tenant in same batch!
			EntityType: "pos_transaction",
			EntityId:   102,
		},
	}

	ack := s.processPushBatch(ctx, events)
	if ack.Success {
		t.Fatal("expected processPushBatch to fail for mixed tenant events")
	}
	if ack.FailedSeq != 2 {
		t.Fatalf("expected FailedSeq to be 2, got %d", ack.FailedSeq)
	}
}

func TestBackupValidateToken(t *testing.T) {
	jwtSecret := "super-secure-test-secret-123456"
	server := NewBackupServer(nil, jwtSecret, "")

	ctx := context.Background()

	// 1. Empty token
	err := server.validateToken(ctx, "", "tenant-a")
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated for empty token, got %v", err)
	}

	// 2. Empty tenant slug
	err = server.validateToken(ctx, "some-token", "")
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for empty tenant slug, got %v", err)
	}

	// 3. Arbitrary non-empty token string (not valid JWT, no DB pool)
	err = server.validateToken(ctx, "arbitrary-token-string", "tenant-a")
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated for arbitrary token string, got %v", err)
	}

	// 4. Valid JWT signed for tenant-a
	claimsA := jwt.MapClaims{
		"tenant_slug": "tenant-a",
		"exp":         time.Now().Add(time.Hour).Unix(),
	}
	tokenObjA := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsA)
	tokenStringA, err := tokenObjA.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("failed to sign test JWT: %v", err)
	}

	// Valid JWT accessing tenant-a should pass
	err = server.validateToken(ctx, tokenStringA, "tenant-a")
	if err != nil {
		t.Fatalf("expected valid JWT for tenant-a to succeed, got %v", err)
	}

	// Valid JWT for tenant-a attempting to access tenant-b MUST fail with PermissionDenied
	err = server.validateToken(ctx, tokenStringA, "tenant-b")
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied when token for tenant-a accesses tenant-b, got %v", err)
	}
}
