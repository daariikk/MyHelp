package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/api/response"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/domain"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/use_cases"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type ControlDoctorsWrapper interface {
	NewDoctor(domain.Doctor) (domain.Doctor, error)
	DeleteDoctor(int) (bool, error)
	GetDoctorById(int) (domain.Doctor, error)
	GetScheduleForDoctor(int, time.Time) ([]domain.Record, error)
	CreateNewScheduleForDoctor(int, []domain.Record) error
}

func NewDoctorHandler(logger *slog.Logger, wrapper ControlDoctorsWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("NewDoctorHandler starting...")

		var newDoctor domain.Doctor
		err := json.NewDecoder(r.Body).Decode(&newDoctor)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse body: %e", err))
			response.SendFailureResponse(w, "Error parse body", http.StatusBadRequest)
		}

		logger.Debug("newDoctor", newDoctor)

		doctor, err := wrapper.NewDoctor(newDoctor)
		if err != nil {
			logger.Error(fmt.Sprintf("Error create doctor: %e", err))
			response.SendFailureResponse(w, "Error create doctor", http.StatusInternalServerError)
		}

		logger.Debug("New doctor: ", "doctor", doctor)
		logger.Info("NewDoctorHandler works successful")

		response.SendSuccessResponse(w, doctor, http.StatusOK)
	}
}

func DeleteDoctorHandler(logger *slog.Logger, wrapper ControlDoctorsWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("DeleteDoctorHandler starting...")

		doctorIDStr := chi.URLParam(r, "doctorID")
		doctorID, err := strconv.ParseInt(doctorIDStr, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse doctorID: %e", err))
			response.SendFailureResponse(w, "Error parse doctorID", http.StatusBadRequest)
		}

		isDeleted, err := wrapper.DeleteDoctor(int(doctorID))
		if err != nil {
			logger.Error(fmt.Sprintf("Error delete doctor: %e", err))
			response.SendFailureResponse(w, "Error delete doctor", http.StatusInternalServerError)
		}

		if !isDeleted {
			logger.Info(fmt.Sprintf("Delete doctor successfully, but doctor with doctorID=%v not found", doctorID))
			response.SendSuccessResponse(w, fmt.Sprintf("Delete doctor successfully, but doctor with doctorID=%v", doctorID), http.StatusOK)
		}

		logger.Info("DeleteDoctorHandler works successful")
		response.SendSuccessResponse(w, "Delete doctor successfully", http.StatusOK)
	}
}

func GetScheduleDoctorByIdHandler(logger *slog.Logger, wrapper ControlDoctorsWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("GetScheduleDoctorByIdHandler starting...")

		doctorIDStr := chi.URLParam(r, "doctorID")
		doctorID, err := strconv.ParseInt(doctorIDStr, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse doctorID: %e", err))
			response.SendFailureResponse(w, "Error parse doctorID", http.StatusBadRequest)
		}
		dateStr := r.URL.Query().Get("date")
		if dateStr == "" {
			logger.Error("Date parameter is required")
			response.SendFailureResponse(w, "Date parameter is required", http.StatusBadRequest)
			return
		}

		date, err := time.Parse("2006-01-02", dateStr) // Формат даты: ГГГГ-ММ-ДД
		if err != nil {
			logger.Error(fmt.Sprintf("Error parsing date: %v", err))
			response.SendFailureResponse(w, "Invalid date format. Expected format: YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		logger.Info(fmt.Sprintf("Parsed date: %v", date))

		doctor, err := wrapper.GetDoctorById(int(doctorID))
		if err != nil {
			logger.Error(fmt.Sprintf("Error get doctor: %e", err))
			response.SendFailureResponse(w, "Error get doctor", http.StatusInternalServerError)
		}

		schedule, err := wrapper.GetScheduleForDoctor(doctor.Id, date)
		if err != nil {
			logger.Error(fmt.Sprintf("Error get doctor schedule: %e", err))
			response.SendFailureResponse(w, "Error get doctor schedule", http.StatusInternalServerError)
		}

		// Формируем ответ
		scheduleInfo := domain.ScheduleInfoDTO{
			Doctor:   doctor,
			Schedule: domain.Schedule{Records: schedule},
		}

		response.SendSuccessResponse(w, scheduleInfo, http.StatusOK)
	}
}

func NewScheduleHandler(logger *slog.Logger, wrapperDB ControlDoctorsWrapper, wrapper use_cases.NewScheduleWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("NewScheduleHandler starting...")

		doctorIDStr := chi.URLParam(r, "doctorID")
		doctorID, err := strconv.ParseInt(doctorIDStr, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse doctorID: %e", err))
			response.SendFailureResponse(w, "Error parse doctorID", http.StatusBadRequest)
		}

		dateStr := r.URL.Query().Get("date")
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse date: %e", err))
			response.SendFailureResponse(w, "Error parse date", http.StatusBadRequest)
		}
		startTimeStr := r.URL.Query().Get("start_time")
		startTime, err := time.Parse("15:04:05", startTimeStr)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse start_time: %e", err))
			response.SendFailureResponse(w, "Error parse start_time", http.StatusBadRequest)
		}

		endTimeStr := r.URL.Query().Get("end_time")
		endTime, err := time.Parse("15:04:05", endTimeStr)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse end_time: %e", err))
			response.SendFailureResponse(w, "Error parse end_time", http.StatusBadRequest)
		}

		receptionTimeStr := r.URL.Query().Get("reception_time")
		receptionTime, err := strconv.ParseInt(receptionTimeStr, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse reception_time: %e", err))
			response.SendFailureResponse(w, "Error parse reception_time", http.StatusBadRequest)
		}

		newSchedule, err := wrapper.CreateScheduleForDoctorById(int(doctorID), date, startTime, endTime, int(receptionTime))
		if err != nil {
			logger.Error(fmt.Sprintf("Error create doctor schedule: %e", err))
			response.SendFailureResponse(w, "Error create doctor schedule", http.StatusInternalServerError)
		}

		err = wrapperDB.CreateNewScheduleForDoctor(int(doctorID), newSchedule.Records)
		if err != nil {
			logger.Error(fmt.Sprintf("Error create doctor schedule: %e", err))
			response.SendFailureResponse(w, "Error create doctor schedule", http.StatusInternalServerError)
		}

		response.SendSuccessResponse(w, newSchedule, http.StatusOK)
	}
}
