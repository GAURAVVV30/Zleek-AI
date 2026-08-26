package middleware

import (
	"net/http"
	"time"

	"github.com/hcl-backend/services/api-go/internal/platform/logger"
	"github.com/hcl-backend/services/api-go/internal/platform/tracing"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			// Inject logger into context for downstream handlers
			ctxLogger := log.With("request_id", tracing.RequestIDFromContext(r.Context()))
			ctx := ctxLogger.WithContext(r.Context())

			next.ServeHTTP(rw, r.WithContext(ctx))

			ctxLogger.Info("HTTP request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
