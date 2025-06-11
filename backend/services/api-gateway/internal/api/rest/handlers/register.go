package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/api/response"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/domain"
	"log/slog"
	"net/http"
)

type RegisterWrapper interface {
	RegisterUser(user domain.User) (domain.User, error)
}

func RegisterHandler(logger *slog.Logger, register RegisterWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("RegisterHandler starting...")

		request := domain.User{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		// logger.Debug("request: ", request)

		// Проверяем обязательные поля
		if request.Email == "" {
			response.SendFailureResponse(w, "Email is required", http.StatusBadRequest)
			return
		}
		if request.Polic == "" {
			response.SendFailureResponse(w, "Polic is required", http.StatusBadRequest)
			return
		}
		if request.Name == "" {
			response.SendFailureResponse(w, "Name is required", http.StatusBadRequest)
			return
		}
		if request.Password == "" {
			response.SendFailureResponse(w, "Password is required", http.StatusBadRequest)
			return
		}

		newUser, err := register.RegisterUser(request)
		if err != nil {
			response.SendFailureResponse(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
			return
		}
		logger.Info("RegisterHandler works successful")
		newUser.Password = ""
		response.SendSuccessResponse(w, newUser, http.StatusCreated)
	}
}
