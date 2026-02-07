package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

// NewLoggingInterceptor returns a Connect interceptor that logs request details.
func NewLoggingInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()
			procedure := req.Spec().Procedure
			slog.InfoContext(ctx, "rpc started",
				"procedure", procedure,
				"protocol", req.Peer().Protocol,
			)

			resp, err := next(ctx, req)
			duration := time.Since(start)

			if err != nil {
				slog.ErrorContext(ctx, "rpc failed",
					"procedure", procedure,
					"duration", duration,
					"error", err,
				)
			} else {
				slog.InfoContext(ctx, "rpc completed",
					"procedure", procedure,
					"duration", duration,
				)
			}

			return resp, err
		}
	}
}

// NewRecoveryInterceptor returns a Connect interceptor that recovers from panics.
func NewRecoveryInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
			defer func() {
				if r := recover(); r != nil {
					slog.ErrorContext(ctx, "panic recovered",
						"procedure", req.Spec().Procedure,
						"panic", r,
					)
					err = connect.NewError(connect.CodeInternal, fmt.Errorf("internal error"))
				}
			}()
			return next(ctx, req)
		}
	}
}
