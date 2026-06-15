package middleware

import (
	"context"
	"net/http"
	"strings"

	"be-fittracker/internal/utils"
)

type contextKey string

const userIDKey contextKey = "user_id"

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing bearer token")
				return
			}

			claims, err := utils.ParseAccessToken(jwtSecret, strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Invalid bearer token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok && userID != ""
}
