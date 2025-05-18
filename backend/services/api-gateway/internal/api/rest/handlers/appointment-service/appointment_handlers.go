package appointment_service

import (
	"fmt"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/api/rest/helper"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/config"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

func NewAppointment(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("%s/MyHelp/schedule/appointments", cfg.Services.AppointmentService)
		helper.ForwardRequest(logger, w, r, url, "POST")
	}
}

func UpdateAppointment(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appointmentIDStr := chi.URLParam(r, "appointmentID")
		url := fmt.Sprintf("%s/MyHelp/schedule/appointments/%s", cfg.Services.AppointmentService, appointmentIDStr)
		helper.ForwardRequest(logger, w, r, url, "PATCH")
	}
}

func DeleteAppointment(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appointmentIDStr := chi.URLParam(r, "appointmentID")
		url := fmt.Sprintf("%s/MyHelp/schedule/appointments/%s", cfg.Services.AppointmentService, appointmentIDStr)
		helper.ForwardRequest(logger, w, r, url, "DELETE")
	}
}
