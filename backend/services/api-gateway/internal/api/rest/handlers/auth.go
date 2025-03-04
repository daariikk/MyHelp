package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/api/response"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/config"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/domain"
	"log/slog"
	"net/http"
	"time"
)

type LoginWrapper interface {
	GetPassword(string) (int, string, error)
}

func LoginHandler(logger *slog.Logger, auth LoginWrapper, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("LoginHandler starting...")

		request := domain.User{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		logger.Debug("Body успешно распарсен")

		logger.Debug("Пытаемся получить пароль по указанному email")
		patientId, encodedPassword, err := auth.GetPassword(request.Email)
		if err != nil {
			logger.Error("Произошла ошибка внутри функции GetPassword")
			logger.Error(err.Error())
			response.SendFailureResponse(w, fmt.Sprintf("Failed to auth user: %v", err), http.StatusInternalServerError)
			return
		}
		logger.Debug("GetPassword отработала успешно")
		logger.Debug("patientId и encodedPassword", slog.Int("patientId", patientId), slog.String("encodedPassword", encodedPassword))

		logger.Debug("Пытаемся проверить совпадают ли пароли")
		logger.Debug("Расшифрованный пароль: ", slog.String("decodedPassword", encodedPassword))
		if encodedPassword != request.Password {
			logger.Info("Пароли не совпадают")
			logger.Error(err.Error())
			response.SendFailureResponse(w, fmt.Sprintf("Failed to auth user: %v", err), http.StatusUnauthorized)
			return
		}
		logger.Debug("Пароль введен успешно")

		logger.Debug("Пытаемся снегерировать токен")
		accessToken, err := generateAccessToken(cfg, patientId)
		if err != nil {
			logger.Error("Произошла ошибка внутри функции generateAccessToken")
			logger.Error("Failed to generate token", slog.String("error", err.Error()))
			response.SendFailureResponse(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		refreshToken, err := generateRefreshToken(cfg, patientId)
		if err != nil {
			logger.Error("Произошла ошибка внутри функции generateAccessToken")
			logger.Error("Failed to generate token", slog.String("error", err.Error()))
			response.SendFailureResponse(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		logger.Debug("Токен успешно сгенерирован")

		accessLifetime := cfg.JWT.ExpireAccess
		refreshLifetime := cfg.JWT.ExpireRefresh

		logger.Debug("Access lifetime", slog.Time("access_lifetime", time.Now().Add(accessLifetime)))
		logger.Debug("Refresh lifetime", slog.Time("refresh_lifetime", time.Now().Add(refreshLifetime)))

		logger.Debug("Формируем ответ")
		res := map[string]interface{}{
			"patientID":        patientId,
			"access_token":     accessToken,
			"access_lifetime":  time.Now().Add(accessLifetime).Format(time.RFC3339),
			"refresh_token":    refreshToken,
			"refresh_lifetime": time.Now().Add(refreshLifetime).Format(time.RFC3339),
		}

		logger.Debug("Сформированный ответ", res)

		logger.Info("LoginHandler works successful")
		response.SendSuccessResponse(w, res, http.StatusOK)
	}
}
