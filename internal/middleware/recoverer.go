package middleware

import (
	"log/slog"
	"net/http"

	"be-fittracker/internal/utils"
)

func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("panic recovered", "value", recovered, "path", r.URL.Path)
					utils.WriteError(w, http.StatusInternalServerError, "internal_server_error", "Internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
