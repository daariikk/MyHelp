package response

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

func SendSuccessResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   data,
	})
}
