package middleware

import (
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/platform/tracing"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")

		ctx := tracing.WithRequestID(r.Context(), reqID)

		// If the client didn't provide one, get the generated one
		if reqID == "" {
			reqID = tracing.RequestIDFromContext(ctx)
		}

		w.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
