package handlers

import (
	"fmt"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// Генерация Access токена
func generateAccessToken(cfg *config.Config, patientID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"patient_id": patientID,
		"exp":        time.Now().Add(time.Minute * 15).Unix(), // Access токен на 15 минут
	})

	secretKey := []byte(cfg.JWT.AccessSecretKey)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}



// Генерация Refresh токена
func generateRefreshToken(cfg *config.Config, patientID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"patient_id": patientID,
		"exp":        time.Now().Add(time.Hour * 24).Unix(), // Refresh токен на 1 дней
	})

	secretKey := []byte(cfg.JWT.RefreshSecretKey)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Валидация токена
func verifyToken(tokenString string, secretKey []byte) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

// Обновление Access токена
func refreshAccessToken(cfg *config.Config, refreshTokenString string) (string, error) {
	// Валидируем Refresh токен
	refreshToken, err := verifyToken(refreshTokenString, []byte(cfg.JWT.RefreshSecretKey))
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %v", err)
	}

	// Извлекаем claims из Refresh токена
	claims, ok := refreshToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Получаем patient_id из claims
	patientID, ok := claims["patient_id"].(float64)
	if !ok {
		return "", fmt.Errorf("invalid patient_id in token")
	}

	// Генерируем новый Access токен
	accessToken, err := generateAccessToken(cfg, int(patientID))
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %v", err)
	}

	return accessToken, nil
}
