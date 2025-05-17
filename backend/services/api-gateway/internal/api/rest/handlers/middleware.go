package handlers

import (
	"context"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"net/http"
	"strings"
)

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(logger *slog.Logger, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debug("AuthMiddleware starting...")

			// Извлекаем токен из заголовка Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Error("Authorization header is missing")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			logger.Debug("Токен успешно извлечен из заголовка")
			// Проверяем, что заголовок начинается с "Bearer "
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				logger.Error("Invalid Authorization header format")
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			// Проверяем токен
			token, err := verifyToken(tokenString, []byte(cfg.JWT.AccessSecretKey))
			if err != nil {
				logger.Error("Failed to verify token", slog.String("error", err.Error()))
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Проверяем, валиден ли токен
			if !token.Valid {
				logger.Error("Invalid token")
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Извлекаем claims из токена
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				logger.Error("Failed to extract claims from token")
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Извлекаем patient_id из claims
			patientID, ok := claims["patient_id"].(float64)
			if !ok {
				logger.Error("Failed to extract patient_id from token")
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Добавляем patient_id в контекст запроса
			ctx := context.WithValue(r.Context(), "patient_id", int64(patientID))
			next.ServeHTTP(w, r.WithContext(ctx))
			logger.Info("AuthMiddleware works successful")
		})
	}
}
