package interceptors

import (
	"context"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
)

type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
)

// NewAuthInterceptor creates a new auth interceptor
// For demo purposes, this is a simple implementation
func NewAuthInterceptor(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Skip auth for health checks
			if strings.HasSuffix(req.Spec().Procedure, "Health") {
				return next(ctx, req)
			}

			// Get authorization header
			authHeader := req.Header().Get("Authorization")
			if authHeader == "" {
				logger.Warn("missing authorization header",
					"procedure", req.Spec().Procedure,
				)
				// For demo, we'll allow requests without auth
				// In production, return an error here
				return next(ctx, req)
			}

			// Extract token (Bearer <token>)
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				logger.Warn("invalid authorization header format",
					"procedure", req.Spec().Procedure,
				)
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			// Validate token (simplified for demo)
			// In production, use JWT validation
			if token == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			// Add user ID to context
			// In production, extract from JWT claims
			ctx = context.WithValue(ctx, UserIDKey, "demo-user")

			logger.Info("request authenticated",
				"procedure", req.Spec().Procedure,
				"user_id", "demo-user",
			)

			return next(ctx, req)
		}
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
