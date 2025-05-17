package account_service

import (
	"fmt"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/api/rest/helper"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/config"
	"log/slog"
	"net/http"
)

func GetPatient(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		patientIDStr := r.URL.Query().Get("patientID")
		url := fmt.Sprintf("%s/MyHelp/account?patientID=%s", cfg.Services.AccountService, patientIDStr)
		helper.ForwardRequest(logger, w, r, url, "GET")
	}
}

func UpdatePatient(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		patientIDStr := r.URL.Query().Get("patientID")
		url := fmt.Sprintf("%s/MyHelp/account?patientID=%s", cfg.Services.AccountService, patientIDStr)
		helper.ForwardRequest(logger, w, r, url, "PUT")
	}
}

func DeletePatient(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		patientIDStr := r.URL.Query().Get("patientID")
		url := fmt.Sprintf("%s/MyHelp/account?patientID=%s", cfg.Services.AccountService, patientIDStr)
		helper.ForwardRequest(logger, w, r, url, "DELETE")
	}
}
