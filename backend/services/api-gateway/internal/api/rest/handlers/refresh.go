package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/api/response"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/config"
	"log/slog"
	"net/http"
	"time"
)

func RefreshHandler(logger *slog.Logger, cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("RefreshHandler starting...")

		var request struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		logger.Debug("request: ", request)

		newAccessToken, err := refreshAccessToken(cfg, request.RefreshToken)
		if err != nil {
			response.SendFailureResponse(w, fmt.Sprintf("Failed to refresh access token: %v", err), http.StatusUnauthorized)
			return
		}
		// Формируем ответ
		accessLifetime := time.Duration(cfg.JWT.ExpireAccess) * time.Minute

		res := map[string]interface{}{
			"access_token":    newAccessToken,
			"access_lifetime": time.Now().Add(accessLifetime).Format(time.RFC3339),
		}

		logger.Info("RefreshHandler works successful")

		response.SendSuccessResponse(w, res, http.StatusCreated)
	}
}
