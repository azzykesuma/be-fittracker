package progress

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
	r.Get("/body-measurements", handler.bodyMeasurements)
	r.Post("/body-measurements", handler.createBodyMeasurement)
	return r
}

func (handler *Handler) BodyMeasurementRoutes(jwtSecret string) chi.Router {
	r := chi.NewRouter()
	r.Use(appmiddleware.Auth(jwtSecret))
	r.Post("/", handler.createBodyMeasurement)
	return r
}

func (handler *Handler) bodyMeasurements(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}

	items, err := handler.service.BodyMeasurements(r.Context(), userID, bodyMeasurementQuery{
		From: r.URL.Query().Get("from"),
		To:   r.URL.Query().Get("to"),
	})
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "body_measurements_failed", err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (handler *Handler) createBodyMeasurement(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}

	var req createBodyMeasurementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}

	measurement, err := handler.service.CreateBodyMeasurement(r.Context(), userID, req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "body_measurement_create_failed", err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{"data": measurement})
}
