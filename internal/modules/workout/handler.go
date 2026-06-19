package workout

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

func (handler *Handler) PlanRoutes(jwtSecret string) chi.Router {
	r := chi.NewRouter()
	r.Use(appmiddleware.Auth(jwtSecret))
	r.Get("/", handler.listPlans)
	r.Post("/", handler.createPlan)
	r.Get("/today", handler.todayPlans)
	r.Get("/{id}", handler.findPlan)
	r.Put("/{id}", handler.updatePlan)
	r.Delete("/{id}", handler.deletePlan)
	r.Post("/{id}/exercises", handler.createExercise)
	return r
}

func (handler *Handler) ExerciseRoutes(jwtSecret string) chi.Router {
	r := chi.NewRouter()
	r.Use(appmiddleware.Auth(jwtSecret))
	r.Put("/{id}", handler.updateExercise)
	r.Delete("/{id}", handler.deleteExercise)
	return r
}

func (handler *Handler) listPlans(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	items, err := handler.service.ListPlans(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "workout_plans_failed", "Failed to load workout plans")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (handler *Handler) todayPlans(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	items, err := handler.service.TodayPlans(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "today_workout_plans_failed", "Failed to load today's workout plans")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (handler *Handler) createPlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	var req workoutPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}
	item, err := handler.service.CreatePlan(r.Context(), userID, req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "workout_plan_create_failed", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (handler *Handler) findPlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	item, err := handler.service.FindPlan(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		writeNotFoundOrServerError(w, err, "workout_plan_not_found", "Workout plan not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (handler *Handler) updatePlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	var req workoutPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}
	item, err := handler.service.UpdatePlan(r.Context(), chi.URLParam(r, "id"), userID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, "workout_plan_not_found", "Workout plan not found")
			return
		}
		utils.WriteError(w, http.StatusBadRequest, "workout_plan_update_failed", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (handler *Handler) deletePlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	if err := handler.service.DeletePlan(r.Context(), chi.URLParam(r, "id"), userID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "workout_plan_delete_failed", "Failed to delete workout plan")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": map[string]bool{"success": true}})
}

func (handler *Handler) createExercise(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	var req exerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}
	item, err := handler.service.CreateExercise(r.Context(), chi.URLParam(r, "id"), userID, req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "exercise_create_failed", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (handler *Handler) updateExercise(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	var req exerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}
	item, err := handler.service.UpdateExercise(r.Context(), chi.URLParam(r, "id"), userID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, "exercise_not_found", "Exercise not found")
			return
		}
		utils.WriteError(w, http.StatusBadRequest, "exercise_update_failed", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (handler *Handler) deleteExercise(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	if err := handler.service.DeleteExercise(r.Context(), chi.URLParam(r, "id"), userID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "exercise_delete_failed", "Failed to delete exercise")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": map[string]bool{"success": true}})
}

func userIDFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return "", false
	}
	return userID, true
}

func writeNotFoundOrServerError(w http.ResponseWriter, err error, code string, message string) {
	if errors.Is(err, pgx.ErrNoRows) {
		utils.WriteError(w, http.StatusNotFound, code, message)
		return
	}
	utils.WriteError(w, http.StatusInternalServerError, code, message)
}

func (handler *Handler) SessionRoutes(jwtSecret string) chi.Router {
	r := chi.NewRouter()
	r.Use(appmiddleware.Auth(jwtSecret))
	r.Post("/start", handler.startSession)
	r.Get("/", handler.listSessions)
	r.Get("/{id}", handler.findSession)
	r.Post("/{id}/sets", handler.logSet)
	r.Put("/{id}/finish", handler.finishSession)
	r.Delete("/{id}", handler.deleteSession)
	return r
}

func (handler *Handler) startSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	var req startSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}
	item, err := handler.service.StartSession(r.Context(), userID, req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "workout_session_start_failed", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (handler *Handler) listSessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	filter := listSessionsFilter{
		From:   r.URL.Query().Get("from"),
		To:     r.URL.Query().Get("to"),
		Status: r.URL.Query().Get("status"),
	}
	items, err := handler.service.ListSessions(r.Context(), userID, filter)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "workout_sessions_failed", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (handler *Handler) findSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	item, err := handler.service.FindSession(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		writeNotFoundOrServerError(w, err, "workout_session_not_found", "Workout session not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (handler *Handler) logSet(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	var req logSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}
	item, err := handler.service.LogSet(r.Context(), chi.URLParam(r, "id"), userID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, "workout_session_not_found", "Workout session not found")
			return
		}
		utils.WriteError(w, http.StatusBadRequest, "workout_set_log_failed", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (handler *Handler) finishSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	var req finishSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}
	item, err := handler.service.FinishSession(r.Context(), chi.URLParam(r, "id"), userID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, "workout_session_not_found", "Workout session not found")
			return
		}
		utils.WriteError(w, http.StatusBadRequest, "workout_session_finish_failed", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (handler *Handler) deleteSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	err := handler.service.DeleteSession(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, "workout_session_not_found", "Workout session not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "workout_session_delete_failed", "Failed to delete workout session")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": map[string]bool{"success": true}})
}

