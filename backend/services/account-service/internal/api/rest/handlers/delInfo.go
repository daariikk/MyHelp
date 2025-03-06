package handlers

import (
	"fmt"
	"github.com/daariikk/MyHelp/services/account-service/internal/api/response"
	"log/slog"
	"net/http"
	"strconv"
)

type DeletePatientWrapper interface {
	DeletePatientById(int) (bool, error)
}

func DeletePatientHandler(logger *slog.Logger, wrapper DeletePatientWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("DeletePatientHandler starting...")

		patientIDStr := r.URL.Query().Get("patientID")
		patientID, err := strconv.ParseInt(patientIDStr, 10, 64)

		logger.Debug("Handling DELETE patient request for patient", "patientID", patientID)

		isDeleted, err := wrapper.DeletePatientById(int(patientID))
		if err != nil {
			response.SendFailureResponse(w, "Error deleting patient: "+err.Error(), http.StatusInternalServerError)
		}

		if isDeleted {
			response.SendSuccessResponse(w, fmt.Sprintf("Patient with patientID=%v deleted", patientID), http.StatusNoContent)
		} else {
			response.SendSuccessResponse(w, fmt.Sprintf("Patient with patientID=%v not found", patientID), http.StatusNoContent)
		}
		logger.Info("DeletePatientHandler works successful")
	}
}
