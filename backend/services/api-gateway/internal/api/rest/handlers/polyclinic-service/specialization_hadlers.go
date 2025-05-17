package polyclinic_service

import (
	"fmt"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/api/rest/helper"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/config"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

func GetPolyclinic(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("%s/MyHelp/specializations", cfg.Services.PolyclinicService)
		helper.ForwardRequest(logger, w, r, url, "GET")
	}
}

func GetDoctorsBySpecialization(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		specializationID := chi.URLParam(r, "specializationID")
		url := fmt.Sprintf("%s/MyHelp/specializations/{%s}", cfg.Services.PolyclinicService, specializationID)
		helper.ForwardRequest(logger, w, r, url, "GET")
	}
}

func NewSpecialization(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("%s/MyHelp/specializations", cfg.Services.PolyclinicService)
		helper.ForwardRequest(logger, w, r, url, "POST")
	}
}

func DeleteSpecialization(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		specializationID := chi.URLParam(r, "specializationID")
		url := fmt.Sprintf("%s/MyHelp/specializations/{%s}", cfg.Services.PolyclinicService, specializationID)
		helper.ForwardRequest(logger, w, r, url, "DELETE")
	}
}
