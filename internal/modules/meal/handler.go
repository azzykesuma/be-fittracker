package meal

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	appmiddleware "be-fittracker/internal/middleware"
	"be-fittracker/internal/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) Routes(jwtSecret string) chi.Router {
	r := chi.NewRouter()
	r.Use(appmiddleware.Auth(jwtSecret))
	r.Get("/", handler.list)
	r.Post("/", handler.create)
	r.Get("/calories", handler.calorieSummary)
	r.Get("/{id}", handler.find)
	r.Put("/{id}", handler.update)
	r.Delete("/{id}", handler.delete)
	return r
}

func (handler *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}
	items, err := handler.service.List(r.Context(), userID, r.URL.Query().Get("from"), r.URL.Query().Get("to"), r.URL.Query().Get("meal_type"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "meal_logs_failed", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (handler *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}
	var req mealLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}
	item, err := handler.service.Create(r.Context(), userID, req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "meal_log_create_failed", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (handler *Handler) find(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}
	item, err := handler.service.Find(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, "meal_log_not_found", "Meal log not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "meal_log_failed", "Failed to load meal log")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (handler *Handler) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}
	var req mealLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}
	item, err := handler.service.Update(r.Context(), chi.URLParam(r, "id"), userID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, "meal_log_not_found", "Meal log not found")
			return
		}
		utils.WriteError(w, http.StatusBadRequest, "meal_log_update_failed", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (handler *Handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}
	if err := handler.service.Delete(r.Context(), chi.URLParam(r, "id"), userID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "meal_log_delete_failed", "Failed to delete meal log")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": map[string]bool{"success": true}})
}

func (handler *Handler) calorieSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}
	summary, err := handler.service.CalorieSummary(r.Context(), userID, r.URL.Query().Get("date"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "calorie_summary_failed", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": summary})
}
