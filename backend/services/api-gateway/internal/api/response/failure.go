package response

import (
	"encoding/json"
	"net/http"
)

const ContentTypeJSON = "application/json"

type FailureResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func SendFailureResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(FailureResponse{
		Status:  "failure",
		Message: message,
	})
}
