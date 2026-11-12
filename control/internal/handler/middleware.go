package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func requestLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := middleware.GetReqID(r.Context())
			w.Header().Set("X-Request-ID", id)
			start := time.Now()
			response := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(response, r)
			logger.InfoContext(r.Context(), "http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", response.Status(),
				slog.Duration("duration", time.Since(start)),
				"request_id", id,
			)
		})
	}
}

func recoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if value := recover(); value != nil {
					logger.ErrorContext(r.Context(), "handler panic", "panic", value, "request_id", middleware.GetReqID(r.Context()))
					if response, ok := w.(interface{ Status() int }); ok && response.Status() != 0 {
						return
					}
					writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error", Code: "INTERNAL_ERROR"})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
