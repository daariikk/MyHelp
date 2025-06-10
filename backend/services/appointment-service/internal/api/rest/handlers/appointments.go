package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/daariikk/MyHelp/services/appointment-service/internal/api/response"
	"github.com/daariikk/MyHelp/services/appointment-service/internal/domain"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type AppointmentWrapper interface {
	NewAppointment(appointment domain.Appointment) error
	UpdateAppointment(appointment domain.Appointment) error
	DeleteAppointment(appointmentID int) error
}

func CreateAppointmentHandler(logger *slog.Logger, wrapper AppointmentWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("CreateAppointmentHandler starting...")

		var newAppointment domain.AppointmentDTO
		var appointment domain.Appointment

		err := json.NewDecoder(r.Body).Decode(&newAppointment)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse body: %e", err))
			response.SendFailureResponse(w, "Error parse body", http.StatusBadRequest)
			return
		}

		if newAppointment.DoctorID == 0 {
			response.SendFailureResponse(w, "doctorID is missing", http.StatusBadRequest)
			return
		}
		appointment.DoctorID = newAppointment.DoctorID

		if newAppointment.PatientID == 0 {
			response.SendFailureResponse(w, "PatientID is missing", http.StatusBadRequest)
			return
		}
		appointment.PatientID = newAppointment.PatientID

		appointment.Date, err = time.Parse("2006-01-02", newAppointment.Date) // Формат даты: ГГГГ-ММ-ДД
		if err != nil {
			logger.Error(fmt.Sprintf("Error parsing date: %v", err))
			response.SendFailureResponse(w, "Invalid date format. Expected format: YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		appointment.Time, err = time.Parse("15:04:05", newAppointment.Time) // Формат времени: HH:MM:SS
		if err != nil {
			logger.Error(fmt.Sprintf("Error parsing time: %v", err))
			response.SendFailureResponse(w, "Invalid time format. Expected format: HH:MM:SS", http.StatusBadRequest)
			return
		}

		err = wrapper.NewAppointment(appointment)
		if err != nil {
			if err.Error() == "Record is busy" {
				logger.Error(fmt.Sprintf("Error create appointment: %e", err))
				response.SendFailureResponse(w, "Record is busy", http.StatusInternalServerError)
				return
			}
			logger.Error(fmt.Sprintf("Error create appointment: %e", err))
			response.SendFailureResponse(w, "Error create appointment", http.StatusInternalServerError)
			return
		}
		logger.Info("CreateAppointmentHandler end...")
		response.SendSuccessResponse(w, "Appointment created", http.StatusCreated)
	}
}

func UpdateAppointmentHandler(logger *slog.Logger, wrapper AppointmentWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("UpdateAppointmentHandler starting...")
		appointmentIDStr := chi.URLParam(r, "appointmentID")
		appointmentID, err := strconv.ParseInt(appointmentIDStr, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse appointmentID: %e", err))
			response.SendFailureResponse(w, "Error parse appointmentID", http.StatusBadRequest)
			return
		}
		var appointment domain.Appointment
		err = json.NewDecoder(r.Body).Decode(&appointment)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse body: %e", err))
			response.SendFailureResponse(w, "Error parse body", http.StatusBadRequest)
			return
		}
		if appointment.Rating == 0 {
			response.SendFailureResponse(w, "Rating is missing", http.StatusBadRequest)
			return
		}

		appointment.Id = int(appointmentID)
		err = wrapper.UpdateAppointment(appointment)
		if err != nil {
			logger.Error(fmt.Sprintf("Error update appointment: %e", err))
			response.SendFailureResponse(w, "Error update appointment", http.StatusInternalServerError)
			return
		}
		logger.Info("UpdateAppointmentHandler end...")
		response.SendSuccessResponse(w, "Appointment updated", http.StatusOK)
	}
}

func CancelAppointmentHandler(logger *slog.Logger, wrapper AppointmentWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("CancelAppointmentHandler starting...")
		appointmentIDStr := chi.URLParam(r, "appointmentID")
		appointmentID, err := strconv.ParseInt(appointmentIDStr, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse appointmentID: %e", err))
			response.SendFailureResponse(w, "Error parse appointmentID", http.StatusBadRequest)
			return
		}

		err = wrapper.DeleteAppointment(int(appointmentID))
		if err != nil {
			logger.Error(fmt.Sprintf("Error cancel appointment: %e", err))
			response.SendFailureResponse(w, "Error cancel appointment", http.StatusInternalServerError)
			return
		}
		logger.Info("CancelAppointmentHandler end...")
		response.SendSuccessResponse(w, "Appointment cancelled", http.StatusOK)
	}
}
