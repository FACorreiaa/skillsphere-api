package interceptors

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

// NewLoggingInterceptor creates a new logging interceptor
func NewLoggingInterceptor(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()

			logger.Info("RPC started",
				"procedure", req.Spec().Procedure,
				"peer", req.Peer().Addr,
			)

			resp, err := next(ctx, req)

			duration := time.Since(start)

			if err != nil {
				logger.Error("RPC failed",
					"procedure", req.Spec().Procedure,
					"duration", duration,
					"error", err,
				)
			} else {
				logger.Info("RPC completed",
					"procedure", req.Spec().Procedure,
					"duration", duration,
				)
			}

			return resp, err
		}
	}
}
