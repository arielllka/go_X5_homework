package main

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type requestIDKey struct{}

var requestCounter atomic.Uint64

func requestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strconv.FormatUint(requestCounter.Add(1), 10)
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		response := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			status := response.Status()
			if status == 0 {
				status = http.StatusOK
			}
			slog.InfoContext(r.Context(), "http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"duration_ms", float64(time.Since(start))/float64(time.Millisecond),
				"request_id", requestIDFromContext(r.Context()),
			)
		}()
		next.ServeHTTP(response, r)
	})
}

func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if value := recover(); value != nil {
				slog.ErrorContext(r.Context(), "handler panic", "request_id", requestIDFromContext(r.Context()))
				if response, ok := w.(interface{ Status() int }); ok && response.Status() != 0 {
					return
				}
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
