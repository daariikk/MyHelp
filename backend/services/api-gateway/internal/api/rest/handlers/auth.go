package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/api/response"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/config"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"log/slog"
	"net/http"
	"time"
)

type LoginWrapper interface {
	GetPassword(string) (int, string, error)
	GetAdminPassword(string) (int, string, error)
	GetUser(string) (domain.User, error)
	GetAdmin(string) (domain.Admin, error)
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
			logger.Info("Произошла ошибка внутри функции GetPassword", "error", err.Error())

			if err.Error() == "user not found" {
				response.SendFailureResponse(w, "Пользователь с таким email не существует", http.StatusNotFound)
			} else {
				logger.Error(fmt.Sprintf("Ошибка в GetPassword: %s", err))
				response.SendFailureResponse(w, fmt.Sprintf("Failed to auth user: %s", err), http.StatusInternalServerError)
			}
			return
		}
		logger.Debug("GetPassword отработала успешно")
		// logger.Debug("patientId и encodedPassword", slog.Int("patientId", patientId), slog.String("encodedPassword", encodedPassword))

		// logger.Debug("Пытаемся проверить совпадают ли пароли")
		// logger.Debug("Расшифрованный пароль: ", slog.String("decodedPassword", encodedPassword))
		// logger.Debug("Присланный пароль: ", slog.String("inputPassword", request.Password))
		if encodedPassword != request.Password {
			logger.Info("Пароли не совпадают")
			// logger.Error(err.Error())
			response.SendFailureResponse(w, fmt.Sprintf("Failed to auth user: %v", "Неверный пароль"), http.StatusUnauthorized)
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

		// logger.Debug("Сформированный ответ", res)

		logger.Info("LoginHandler works successful")
		response.SendSuccessResponse(w, res, http.StatusOK)
	}
}

func GetUserHandler(logger *slog.Logger, auth LoginWrapper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("GetUserHandler starting...")
		logger.Debug("r.URL", "r.URL", r.URL)
		email := r.URL.Query().Get("email")
		logger.Debug("email", "email", email)
		if email == "" {
			logger.Error("Email is empty")
			response.SendFailureResponse(w, "Email is empty", http.StatusBadRequest)
			return
		}
		patient, err := auth.GetUser(email)
		if err != nil {
			logger.Error("Произошла ошибка внутри функции GetUser", "error", err)
			if errors.Is(err, pgx.ErrNoRows) || err.Error() == "user not found" {
				response.SendFailureResponse(w, "Пользователь с таким email не существует", http.StatusNotFound)
			} else {
				response.SendFailureResponse(w, "Failed get to user", http.StatusInternalServerError)
			}
			return
		}

		logger.Debug("GetUserHandler works successful")
		response.SendSuccessResponse(w, patient, http.StatusOK)
	}
}

func LoginAdminHandler(logger *slog.Logger, auth LoginWrapper, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("LoginHandler starting...")

		request := domain.Admin{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		logger.Debug("Body успешно распарсен")

		logger.Debug("Пытаемся получить пароль по указанному email")
		adminID, encodedPassword, err := auth.GetAdminPassword(request.Email)
		if err != nil {
			logger.Error("Произошла ошибка внутри функции GetPassword")
			//logger.Error(err.Error())
			response.SendFailureResponse(w, fmt.Sprintf("Failed to auth user: %s", err), http.StatusInternalServerError)
			return
		}
		logger.Debug("GetPassword отработала успешно")
		logger.Debug("patientId и encodedPassword", slog.Int("adminID", adminID), slog.String("encodedPassword", encodedPassword))

		logger.Debug("Пытаемся проверить совпадают ли пароли")
		logger.Debug("Расшифрованный пароль: ", slog.String("decodedPassword", encodedPassword))
		logger.Debug("Присланный пароль: ", slog.String("inputPassword", request.Password))
		if encodedPassword != request.Password {
			logger.Info("Пароли не совпадают")
			// logger.Error(err.Error())
			response.SendFailureResponse(w, fmt.Sprintf("Failed to auth user: %v", err), http.StatusUnauthorized)
			return
		}
		logger.Debug("Пароль введен успешно")

		logger.Debug("Пытаемся снегерировать токен")
		accessToken, err := generateAccessToken(cfg, adminID)
		if err != nil {
			logger.Error("Произошла ошибка внутри функции generateAccessToken")
			logger.Error("Failed to generate token", slog.String("error", err.Error()))
			response.SendFailureResponse(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		refreshToken, err := generateRefreshToken(cfg, adminID)
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
			"adminID":          adminID,
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

func GetAdminHandler(logger *slog.Logger, auth LoginWrapper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("GetAdminHandler starting...")
		logger.Debug("r.URL", "r.URL", r.URL)
		email := r.URL.Query().Get("email")
		logger.Debug("email", "email", email)
		if email == "" {
			logger.Error("Email is empty")
			response.SendFailureResponse(w, "Email is empty", http.StatusBadRequest)
		}
		admin, err := auth.GetAdmin(email)
		if err != nil {
			logger.Error("Произошла ошибка внутри функции GetAdmin")
			response.SendFailureResponse(w, "Failed get to admin", http.StatusInternalServerError)
		}

		logger.Debug("GetAdminHandler works successful")
		response.SendSuccessResponse(w, admin, http.StatusOK)
	}
}
