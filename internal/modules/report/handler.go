package report

import (
	"errors"
	"fmt"
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
	r.Get("/summary", handler.downloadSummaryReport)
	return r
}

func (handler *Handler) downloadSummaryReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserID(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing user context")
		return
	}

	data, fileName, err := handler.service.GenerateSummaryReport(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, "not_found", "User not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "report_generation_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
