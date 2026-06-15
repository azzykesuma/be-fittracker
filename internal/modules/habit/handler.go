package habit

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

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
	r.Post("/{id}/complete", handler.completeToday)
	r.Delete("/{id}/complete", handler.uncompleteToday)
	return r
}

func (handler *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}

	items, err := handler.service.List(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "habits_failed", "Failed to load habits")
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

	var req createHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}

	item, err := handler.service.Create(r.Context(), userID, req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "habit_create_failed", err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (handler *Handler) completeToday(w http.ResponseWriter, r *http.Request) {
	handler.toggleComplete(w, r, true)
}

func (handler *Handler) uncompleteToday(w http.ResponseWriter, r *http.Request) {
	handler.toggleComplete(w, r, false)
}

func (handler *Handler) toggleComplete(w http.ResponseWriter, r *http.Request, completed bool) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}

	habitID := chi.URLParam(r, "id")
	var err error
	if completed {
		err = handler.service.CompleteToday(r.Context(), habitID, userID)
	} else {
		err = handler.service.UncompleteToday(r.Context(), habitID, userID)
	}
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "habit_not_found", "Habit not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": map[string]bool{"success": true}})
}
