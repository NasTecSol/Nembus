package auth

import (
	"context"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	mdAuthorization = "authorization"
	mdAPIKey        = "x-api-key"
	mdAgentKey      = "x-agent-key"
)

// UnaryServerInterceptor validates agent JWT or API key on unary RPCs.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := authorize(ctx); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor validates agent JWT or API key on streaming RPCs.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := authorize(ss.Context()); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

func authorize(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	apiKey := first(md, mdAPIKey)
	if apiKey == "" {
		apiKey = first(md, mdAgentKey)
	}

	expectedKey := os.Getenv("AGENT_API_KEY")
	if expectedKey == "" {
		expectedKey = os.Getenv("MIGRATION_API_KEY")
	}
	if expectedKey != "" && apiKey != "" && apiKey == expectedKey {
		return nil
	}

	bearer := bearerToken(first(md, mdAuthorization))
	if bearer != "" {
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret != "" {
			token, err := jwt.Parse(bearer, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, status.Error(codes.Unauthenticated, "unexpected signing method")
				}
				return []byte(jwtSecret), nil
			})
			if err == nil && token.Valid {
				return nil
			}
		}
	}

	env := os.Getenv("ENV")
	if expectedKey == "" && (env == "development" || env == "dev" || env == "") {
		return nil
	}

	return status.Error(codes.Unauthenticated, "valid x-api-key or Bearer JWT required")
}

func first(md metadata.MD, key string) string {
	vals := md.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

func bearerToken(authHeader string) string {
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return ""
}
