package polyclinic_service

import (
	"fmt"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/api/rest/helper"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/config"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

func NewDoctor(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("%s/MyHelp/doctors", cfg.Services.PolyclinicService)
		helper.ForwardRequest(logger, w, r, url, "POST")
	}
}

func DeleteDoctor(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		doctorID := chi.URLParam(r, "doctorID")
		url := fmt.Sprintf("%s/MyHelp/doctors/%s", cfg.Services.PolyclinicService, doctorID)
		helper.ForwardRequest(logger, w, r, url, "DELETE")
	}
}

func GetSchedule(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		doctorID := chi.URLParam(r, "doctorID")
		url := fmt.Sprintf("%s/MyHelp/schedule/doctors/%s", cfg.Services.PolyclinicService, doctorID)
		helper.ForwardRequest(logger, w, r, url, "GET")
	}
}

func NewSchedule(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		doctorID := chi.URLParam(r, "doctorID")
		url := fmt.Sprintf("%s/MyHelp/schedule/doctors/%s", cfg.Services.PolyclinicService, doctorID)
		helper.ForwardRequest(logger, w, r, url, "POST")
	}
}
