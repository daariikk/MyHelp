package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/api/response"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/domain"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
)

type SpecializationWrapper interface {
	GetAllSpecializations() ([]domain.Specialization, error)
	GetSpecializationAllDoctor(int) ([]domain.Doctor, error)
	CreateNewSpecialization(specialization domain.Specialization) (int, error)
	DeleteSpecialization(int) (bool, error)
}

func GetPolyclinicInfoHandler(logger *slog.Logger, wrapper SpecializationWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("GetPolyclinicInfoHandler starting...")
		schedule, err := wrapper.GetAllSpecializations()
		if err != nil {
			logger.Error(fmt.Sprintf("Error get specialization: %s", err))
			response.SendFailureResponse(w, "Error get specialization", http.StatusInternalServerError)
			return
		}

		if schedule == nil {
			schedule = []domain.Specialization{}
		}

		logger.Debug("Schedule: ", "schedule", schedule)
		logger.Info("GetPolyclinicInfoHandler works successful")

		response.SendSuccessResponse(w, schedule, http.StatusOK)
	}
}

func GetSpecializationDoctorHandler(logger *slog.Logger, wrapper SpecializationWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("GetSpecializationDoctorHandler starting...")
		specializationIDStr := chi.URLParam(r, "specializationID")
		specializationID, err := strconv.ParseInt(specializationIDStr, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse specializationID: %s", err))
			response.SendFailureResponse(w, "Error parse specializationID", http.StatusBadRequest)
			return
		}

		doctors, err := wrapper.GetSpecializationAllDoctor(int(specializationID))
		if err != nil {
			logger.Error(fmt.Sprintf("Error get list doctors for specialization with specializationID=%v: %s", specializationID, err))
			response.SendFailureResponse(w, "Error get list doctors", http.StatusInternalServerError)
			return
		}

		if doctors == nil {
			doctors = []domain.Doctor{}
		}

		logger.Debug("doctors list", "doctors", doctors)
		logger.Info("GetPolyclinicInfoHandler works successful")
		response.SendSuccessResponse(w, doctors, http.StatusOK)
	}
}

func CreateNewSpecializationHandler(logger *slog.Logger, wrapper SpecializationWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("CreateNewSpecializationHandler starting...")

		var newSpecialization domain.Specialization
		err := json.NewDecoder(r.Body).Decode(&newSpecialization)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse body: %s", err))
			response.SendFailureResponse(w, "Error parse body", http.StatusBadRequest)
			return
		}
		logger.Debug("newSpecialization", "newSpecialization", newSpecialization)

		specializationID, err := wrapper.CreateNewSpecialization(newSpecialization)
		if err != nil {
			logger.Error(fmt.Sprintf("Error create specialization: %s", err))
			response.SendFailureResponse(w, "Error create specialization", http.StatusInternalServerError)
			return
		}
		logger.Debug("specializationID", "specializationID", specializationID)

		logger.Info("CreateNewSpecializationHandler works successful")
		response.SendSuccessResponse(w, specializationID, http.StatusCreated)
	}
}

func DeleteSpecializationHandler(logger *slog.Logger, wrapper SpecializationWrapper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("DeleteSpecializationHandler starting...")

		specializationIDStr := chi.URLParam(r, "specializationID")
		specializationID, err := strconv.ParseInt(specializationIDStr, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parse specializationID: %s", err))
			response.SendFailureResponse(w, "Error parse specializationID", http.StatusBadRequest)
			return
		}

		logger.Debug("specializationID", "specializationID", specializationID)
		isDeleted, err := wrapper.DeleteSpecialization(int(specializationID))
		if err != nil {
			logger.Error(fmt.Sprintf("Error delete specialization: %s", err))
			response.SendFailureResponse(w, "Error delete specialization", http.StatusInternalServerError)
			return
		}

		if isDeleted {
			response.SendSuccessResponse(w, "Deleted specialization", http.StatusNoContent)
		} else {
			response.SendSuccessResponse(w, fmt.Sprintf("Not found specialization with specializationID=%v", specializationID), http.StatusNotFound)
		}
	}
}
