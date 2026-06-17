package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			requestID := chimiddleware.GetReqID(r.Context())

			logger.Info("api_endpoint_started",
				"service", "fitflow-api",
				"method", r.Method,
				"path", r.URL.Path,
				"request_id", requestID,
				"remote_ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			next.ServeHTTP(ww, r)

			logger.Info("api_endpoint_completed",
				"service", "fitflow-api",
				"method", r.Method,
				"path", r.URL.Path,
				"route", routePattern(r),
				"status", ww.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", requestID,
			)
		})
	}
}

func routePattern(r *http.Request) string {
	routeContext := chi.RouteContext(r.Context())
	if routeContext == nil {
		return ""
	}
	return routeContext.RoutePattern()
}
