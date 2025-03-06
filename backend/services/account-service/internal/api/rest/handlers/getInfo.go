package handlers

import (
	"errors"
	"fmt"
	"github.com/daariikk/MyHelp/services/account-service/internal/api/response"
	"github.com/daariikk/MyHelp/services/account-service/internal/domain"
	"github.com/daariikk/MyHelp/services/account-service/internal/lib/logger/sl"
	"github.com/daariikk/MyHelp/services/account-service/internal/repository"
	"log/slog"
	"net/http"
	"strconv"
)

type GetPatientWrapper interface {
	GetPatientById(int) (domain.Patient, error)
	GetAppointmentByPatientId(int) ([]domain.Appointment, error)
}

func GetPatientByIdHandler(logger *slog.Logger, wrapper GetPatientWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("GetPatientByIdHandler starting...")

		patientIDStr := r.URL.Query().Get("patientID")
		patientID, err := strconv.ParseInt(patientIDStr, 10, 64)

		logger.Debug("Handling GET patient request for patient", "patientID", patientID)

		patient, err := wrapper.GetPatientById(int(patientID))
		if err != nil {
			if errors.Is(err, repository.ErrorNotFound) {
				logger.Debug("Patient not found", sl.Err(err))
				response.SendFailureResponse(w, fmt.Sprintf("Patient with patientID=%v not found", patientID), http.StatusNotFound)
			} else {
				logger.Debug(fmt.Sprintf("Error get info for patient with patientID=%v", patientID), sl.Err(err))
				response.SendFailureResponse(w, "Failed to get patient", http.StatusInternalServerError)
			}
		}

		appointments, err := wrapper.GetAppointmentByPatientId(int(patientID))
		if err != nil {
			logger.Debug(fmt.Sprintf("Error get appointment by patient with patientID=%v", patientID), sl.Err(err))
		}

		formattedAppointments := make([]domain.AppointmentDTO, len(appointments))

		for i, app := range appointments {
			formattedAppointments[i] = domain.AppointmentDTO{
				Id:                   app.Id,
				DoctorFIO:            app.DoctorFIO,
				DoctorSpecialization: app.DoctorSpecialization,
				Date:                 app.Date.Format("2006-01-02"),
				Time:                 app.Time.Format("15:04:05"),
				Status:               app.Status,
				Rating:               app.Rating,
			}
		}

		// Формируем ответ
		patientInfo := domain.PatientDTO{
			Id:           patient.Id,
			Surname:      patient.Surname,
			Name:         patient.Name,
			Patronymic:   patient.Patronymic,
			Polic:        patient.Polic,
			Email:        patient.Email,
			IsDeleted:    patient.IsDeleted,
			Appointments: formattedAppointments,
		}

		logger.Info("GetPatientByIdHandler works successful")

		response.SendSuccessResponse(w, patientInfo, http.StatusOK)
	}
}
