package middleware

import (
	"drssr/internal/pkg/ctx_utils"
	"net/http"

	"github.com/google/uuid"
)

func WithRequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.NewString()
		w.Header().Set("X-Request-ID", reqID)
		r = r.WithContext(ctx_utils.SetReqID(r.Context(), reqID))
		h.ServeHTTP(w, r)
	})
}
