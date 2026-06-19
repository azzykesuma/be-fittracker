package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	appmiddleware "be-fittracker/internal/middleware"
	"be-fittracker/internal/utils"
)

const refreshTokenCookieName = "fitflow_refresh_token"

type Handler struct {
	service   *Service
	jwtSecret string
}

func NewHandler(service *Service, jwtSecret string) *Handler {
	return &Handler{service: service, jwtSecret: jwtSecret}
}

func (handler *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", handler.register)
	r.Post("/login", handler.login)
	r.Post("/refresh", handler.refresh)
	r.Post("/logout", handler.logout)
	r.With(appmiddleware.Auth(handler.jwtSecret)).Get("/me", handler.me)
	return r
}

func (handler *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}

	res, err := handler.service.Register(r.Context(), req)
	if err != nil {
		slog.Warn("auth_register_failed", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_auth_request", "Invalid auth request")
		return
	}

	setRefreshTokenCookie(w, r, res.RefreshToken)
	utils.WriteJSON(w, http.StatusCreated, map[string]any{"data": res})
}

func (handler *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}

	res, err := handler.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrInvalidAuthRequest) {
			slog.Warn("auth_login_rejected", "reason", err)
			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
			return
		}
		slog.Error("auth_login_failed", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "login_failed", "Login failed")
		return
	}

	setRefreshTokenCookie(w, r, res.RefreshToken)
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": res})
}

func (handler *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.ContentLength > 0 {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}
	if req.RefreshToken == "" {
		req.RefreshToken = refreshTokenFromCookie(r)
	}

	res, err := handler.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		slog.Warn("auth_refresh_rejected", "reason", err)
		utils.WriteError(w, http.StatusUnauthorized, "invalid_refresh_token", "Invalid refresh token")
		return
	}

	setRefreshTokenCookie(w, r, res.RefreshToken)
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": res})
}

func (handler *Handler) logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.RefreshToken == "" {
		req.RefreshToken = refreshTokenFromCookie(r)
	}
	if err := handler.service.Logout(r.Context(), req.RefreshToken); err != nil {
		slog.Error("auth_logout_failed", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "logout_failed", "Logout failed")
		return
	}
	clearRefreshTokenCookie(w, r)
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": map[string]bool{"success": true}})
}

func (handler *Handler) me(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}

	user, err := handler.service.Me(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, "not_found", "User not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "profile_failed", "Failed to load profile")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": user})
}

func refreshTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func setRefreshTokenCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		Path:     "/api/auth",
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearRefreshTokenCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
}

func (handler *Handler) UserRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(appmiddleware.Auth(handler.jwtSecret))
	r.Get("/me", handler.getUserProfile)
	r.Patch("/me", handler.patchUserProfile)
	return r
}

func (handler *Handler) getUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}

	user, err := handler.service.Me(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, "not_found", "User not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "profile_failed", "Failed to load profile")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": user})
}

func (handler *Handler) patchUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}

	var req patchUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}

	user, err := handler.service.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "profile_update_failed", err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": user})
}
