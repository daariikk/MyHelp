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
		logger.Debug("request: ", request)

		newUser, err := register.RegisterUser(request)
		if err != nil {
			response.SendFailureResponse(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
			return
		}
		logger.Info("RegisterHandler works successful")

		response.SendSuccessResponse(w, newUser, http.StatusCreated)
	}
}
