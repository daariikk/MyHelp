package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/daariikk/MyHelp/services/appointment-service/internal/domain"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type MockAppointmentWrapper struct {
	NewAppointmentFn    func(domain.Appointment) error
	UpdateAppointmentFn func(domain.Appointment) error
	DeleteAppointmentFn func(int) error
}

func (m *MockAppointmentWrapper) NewAppointment(a domain.Appointment) error {
	if m.NewAppointmentFn != nil {
		return m.NewAppointmentFn(a)
	}
	return nil
}
func (m *MockAppointmentWrapper) UpdateAppointment(a domain.Appointment) error {
	if m.UpdateAppointmentFn != nil {
		return m.UpdateAppointmentFn(a)
	}
	return nil
}
func (m *MockAppointmentWrapper) DeleteAppointment(id int) error {
	if m.DeleteAppointmentFn != nil {
		return m.DeleteAppointmentFn(id)
	}
	return nil
}

func TestCreateAppointmentHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// success
	t.Run("success", func(t *testing.T) {
		mock := &MockAppointmentWrapper{
			NewAppointmentFn: func(a domain.Appointment) error {
				return nil
			},
		}
		dto := domain.AppointmentDTO{
			DoctorID:  1,
			PatientID: 2,
			Date:      "2025-06-10",
			Time:      "14:30:00",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler := CreateAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("ожидал %d, получил %d", http.StatusCreated, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Appointment created") {
			t.Errorf("ответ не содержит успешного сообщения")
		}
	})

	t.Run("bad body", func(t *testing.T) {
		mock := &MockAppointmentWrapper{}
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("bad json"))
		w := httptest.NewRecorder()
		handler := CreateAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error parse body") {
			t.Errorf("ответ не содержит ошибки парсинга")
		}
	})

	t.Run("missing doctorID", func(t *testing.T) {
		mock := &MockAppointmentWrapper{}
		dto := domain.AppointmentDTO{
			PatientID: 2,
			Date:      "2025-06-10",
			Time:      "14:30:00",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler := CreateAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "doctorID is missing") {
			t.Errorf("ответ не содержит doctorID is missing")
		}
	})

	t.Run("missing patientID", func(t *testing.T) {
		mock := &MockAppointmentWrapper{}
		dto := domain.AppointmentDTO{
			DoctorID: 1,
			Date:     "2025-06-10",
			Time:     "14:30:00",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler := CreateAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "PatientID is missing") {
			t.Errorf("ответ не содержит PatientID is missing")
		}
	})

	t.Run("bad date format", func(t *testing.T) {
		mock := &MockAppointmentWrapper{}
		dto := domain.AppointmentDTO{
			DoctorID:  1,
			PatientID: 2,
			Date:      "10.06.2025",
			Time:      "14:30:00",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler := CreateAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Invalid date format") {
			t.Errorf("ответ не содержит Invalid date format")
		}
	})

	t.Run("bad time format", func(t *testing.T) {
		mock := &MockAppointmentWrapper{}
		dto := domain.AppointmentDTO{
			DoctorID:  1,
			PatientID: 2,
			Date:      "2025-06-10",
			Time:      "bad",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler := CreateAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Invalid time format") {
			t.Errorf("ответ не содержит Invalid time format")
		}
	})

	t.Run("busy record", func(t *testing.T) {
		mock := &MockAppointmentWrapper{
			NewAppointmentFn: func(a domain.Appointment) error {
				return errors.New("Record is busy")
			},
		}
		dto := domain.AppointmentDTO{
			DoctorID:  1,
			PatientID: 2,
			Date:      "2025-06-10",
			Time:      "14:30:00",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler := CreateAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("ожидал %d, получил %d", http.StatusInternalServerError, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Record is busy") {
			t.Errorf("ответ не содержит Record is busy")
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mock := &MockAppointmentWrapper{
			NewAppointmentFn: func(a domain.Appointment) error {
				return errors.New("unexpected error")
			},
		}
		dto := domain.AppointmentDTO{
			DoctorID:  1,
			PatientID: 2,
			Date:      "2025-06-10",
			Time:      "14:30:00",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler := CreateAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("ожидал %d, получил %d", http.StatusInternalServerError, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error create appointment") {
			t.Errorf("ответ не содержит Error create appointment")
		}
	})
}

func TestUpdateAppointmentHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// success
	t.Run("success", func(t *testing.T) {
		mock := &MockAppointmentWrapper{
			UpdateAppointmentFn: func(a domain.Appointment) error {
				return nil
			},
		}
		appointment := domain.Appointment{
			DoctorID:  1,
			PatientID: 2,
			Date:      time.Now(),
			Time:      time.Now(),
			Rating:    4.5,
		}
		body, _ := json.Marshal(appointment)
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("appointmentID", "10")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()
		handler := UpdateAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("ожидал %d, получил %d", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Appointment updated") {
			t.Errorf("ответ не содержит Appointment updated")
		}
	})

	t.Run("bad id", func(t *testing.T) {
		mock := &MockAppointmentWrapper{}
		appointment := domain.Appointment{}
		body, _ := json.Marshal(appointment)
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("appointmentID", "bad")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()
		handler := UpdateAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error parse appointmentID") {
			t.Errorf("ответ не содержит Error parse appointmentID")
		}
	})

	t.Run("bad body", func(t *testing.T) {
		mock := &MockAppointmentWrapper{}
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("bad json"))
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("appointmentID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()
		handler := UpdateAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error parse body") {
			t.Errorf("ответ не содержит Error parse body")
		}
	})

	t.Run("missing rating", func(t *testing.T) {
		mock := &MockAppointmentWrapper{}
		appointment := domain.Appointment{}
		body, _ := json.Marshal(appointment)
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("appointmentID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()
		handler := UpdateAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Rating is missing") {
			t.Errorf("ответ не содержит Rating is missing")
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mock := &MockAppointmentWrapper{
			UpdateAppointmentFn: func(a domain.Appointment) error {
				return errors.New("fail update")
			},
		}
		appointment := domain.Appointment{Rating: 5}
		body, _ := json.Marshal(appointment)
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("appointmentID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()
		handler := UpdateAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("ожидал %d, получил %d", http.StatusInternalServerError, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error update appointment") {
			t.Errorf("ответ не содержит Error update appointment")
		}
	})
}

func TestCancelAppointmentHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	t.Run("success", func(t *testing.T) {
		mock := &MockAppointmentWrapper{
			DeleteAppointmentFn: func(id int) error { return nil },
		}
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("appointmentID", "7")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()
		handler := CancelAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("ожидал %d, получил %d", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Appointment cancelled") {
			t.Errorf("ответ не содержит Appointment cancelled")
		}
	})

	t.Run("bad id", func(t *testing.T) {
		mock := &MockAppointmentWrapper{}
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("appointmentID", "bad")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()
		handler := CancelAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error parse appointmentID") {
			t.Errorf("ответ не содержит Error parse appointmentID")
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mock := &MockAppointmentWrapper{
			DeleteAppointmentFn: func(id int) error { return errors.New("fail delete") },
		}
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("appointmentID", "5")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()
		handler := CancelAppointmentHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("ожидал %d, получил %d", http.StatusInternalServerError, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error cancel appointment") {
			t.Errorf("ответ не содержит Error cancel appointment")
		}
	})
}
