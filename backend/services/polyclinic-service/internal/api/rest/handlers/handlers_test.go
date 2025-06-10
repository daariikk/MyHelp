package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/domain"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type MockControlDoctorsWrapper struct {
	NewDoctorFn                  func(domain.Doctor) (domain.Doctor, error)
	DeleteDoctorFn               func(int) (bool, error)
	GetDoctorByIdFn              func(int) (domain.Doctor, error)
	GetScheduleForDoctorFn       func(int, time.Time) ([]domain.Record, error)
	CreateNewScheduleForDoctorFn func(int, []domain.Record) error
}

func (m *MockControlDoctorsWrapper) NewDoctor(d domain.Doctor) (domain.Doctor, error) {
	if m.NewDoctorFn != nil {
		return m.NewDoctorFn(d)
	}
	return domain.Doctor{}, nil
}
func (m *MockControlDoctorsWrapper) DeleteDoctor(id int) (bool, error) {
	if m.DeleteDoctorFn != nil {
		return m.DeleteDoctorFn(id)
	}
	return false, nil
}
func (m *MockControlDoctorsWrapper) GetDoctorById(id int) (domain.Doctor, error) {
	if m.GetDoctorByIdFn != nil {
		return m.GetDoctorByIdFn(id)
	}
	return domain.Doctor{}, nil
}
func (m *MockControlDoctorsWrapper) GetScheduleForDoctor(id int, date time.Time) ([]domain.Record, error) {
	if m.GetScheduleForDoctorFn != nil {
		return m.GetScheduleForDoctorFn(id, date)
	}
	return nil, nil
}
func (m *MockControlDoctorsWrapper) CreateNewScheduleForDoctor(id int, records []domain.Record) error {
	if m.CreateNewScheduleForDoctorFn != nil {
		return m.CreateNewScheduleForDoctorFn(id, records)
	}
	return nil
}

type MockNewScheduleWrapper struct {
	CreateScheduleForDoctorByIdFn func(int, time.Time, time.Time, time.Time, int) (domain.Schedule, error)
}

func (m *MockNewScheduleWrapper) CreateScheduleForDoctorById(
	doctorID int, date, start, end time.Time, receptionTime int,
) (domain.Schedule, error) {
	return m.CreateScheduleForDoctorByIdFn(doctorID, date, start, end, receptionTime)
}

type MockSpecializationWrapper struct {
	GetAllSpecializationsFn      func() ([]domain.Specialization, error)
	GetSpecializationAllDoctorFn func(int) ([]domain.Doctor, error)
	CreateNewSpecializationFn    func(domain.Specialization) (int, error)
	DeleteSpecializationFn       func(int) (bool, error)
}

func (m *MockSpecializationWrapper) GetAllSpecializations() ([]domain.Specialization, error) {
	return m.GetAllSpecializationsFn()
}
func (m *MockSpecializationWrapper) GetSpecializationAllDoctor(id int) ([]domain.Doctor, error) {
	return m.GetSpecializationAllDoctorFn(id)
}
func (m *MockSpecializationWrapper) CreateNewSpecialization(s domain.Specialization) (int, error) {
	return m.CreateNewSpecializationFn(s)
}
func (m *MockSpecializationWrapper) DeleteSpecialization(id int) (bool, error) {
	return m.DeleteSpecializationFn(id)
}

func TestNewDoctorHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	validDoctor := domain.Doctor{
		Id:             1,
		Surname:        "Иванов",
		Name:           "Иван",
		Patronymic:     "Иванович",
		Specialization: "Терапевт",
	}
	tests := []struct {
		name     string
		body     interface{}
		mockFunc func(domain.Doctor) (domain.Doctor, error)
		wantCode int
		wantBody string
	}{
		{
			name: "успех",
			body: validDoctor,
			mockFunc: func(d domain.Doctor) (domain.Doctor, error) {
				return validDoctor, nil
			},
			wantCode: http.StatusOK,
			wantBody: `"surname":"Иванов"`,
		},
		{
			name: "невалидный json",
			body: "{invalid",
			mockFunc: func(d domain.Doctor) (domain.Doctor, error) {
				t.Fatal("mock не должен вызываться")
				return domain.Doctor{}, nil
			},
			wantCode: http.StatusBadRequest,
			wantBody: `Error parse body`,
		},
		{
			name: "db ошибка",
			body: validDoctor,
			mockFunc: func(d domain.Doctor) (domain.Doctor, error) {
				return domain.Doctor{}, errors.New("fail")
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `Error create doctor`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(v)
				assert.NoError(t, err)
			}
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			w := httptest.NewRecorder()
			mock := &MockControlDoctorsWrapper{NewDoctorFn: tt.mockFunc}
			handler := NewDoctorHandler(logger, mock)
			handler(w, req)
			assert.Equal(t, tt.wantCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
		})
	}
}

func TestDeleteDoctorHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	tests := []struct {
		name       string
		doctorID   string
		mockDelete func(int) (bool, error)
		wantCode   int
		wantBody   string
	}{
		{
			name:       "bad id",
			doctorID:   "abc",
			mockDelete: func(id int) (bool, error) { return false, nil },
			wantCode:   http.StatusBadRequest,
			wantBody:   "Error parse doctorID",
		},
		{
			name:       "успех",
			doctorID:   "7",
			mockDelete: func(id int) (bool, error) { return true, nil },
			wantCode:   http.StatusOK,
			wantBody:   "Delete doctor successfully",
		},
		{
			name:       "not found",
			doctorID:   "8",
			mockDelete: func(id int) (bool, error) { return false, nil },
			wantCode:   http.StatusOK,
			wantBody:   "but doctor with doctorID=8",
		},
		{
			name:       "db error",
			doctorID:   "9",
			mockDelete: func(id int) (bool, error) { return false, errors.New("fail") },
			wantCode:   http.StatusInternalServerError,
			wantBody:   "Error delete doctor",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/", nil)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("doctorID", tt.doctorID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
			w := httptest.NewRecorder()
			mock := &MockControlDoctorsWrapper{DeleteDoctorFn: tt.mockDelete}
			handler := DeleteDoctorHandler(logger, mock)
			handler(w, req)
			assert.Equal(t, tt.wantCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
		})
	}
}

func TestGetScheduleDoctorByIdHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	doctor := domain.Doctor{
		Id:             2,
		Surname:        "Смирнов",
		Name:           "Олег",
		Patronymic:     "Петрович",
		Specialization: "Хирург",
		Education:      "РНИМУ",
		Progress:       "Медцентр 2020-2024",
		Rating:         4.7,
		PhotoPath:      "/img/smirnov.jpg",
	}
	date := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	record := domain.Record{
		ID:          10,
		DoctorId:    doctor.Id,
		Date:        date,
		Start:       time.Date(2025, 7, 1, 9, 0, 0, 0, time.UTC),
		End:         time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC),
		IsAvailable: true,
	}
	tests := []struct {
		name           string
		doctorID       string
		dateStr        string
		mockGetDoctor  func(int) (domain.Doctor, error)
		mockGetRecords func(int, time.Time) ([]domain.Record, error)
		wantCode       int
		wantBody       string
	}{
		{
			name:          "успех",
			doctorID:      "2",
			dateStr:       "2025-07-01",
			mockGetDoctor: func(id int) (domain.Doctor, error) { return doctor, nil },
			mockGetRecords: func(id int, date time.Time) ([]domain.Record, error) {
				return []domain.Record{record}, nil
			},
			wantCode: http.StatusOK,
			wantBody: `"surname":"Смирнов"`,
		},
		{
			name:     "bad doctor id",
			doctorID: "abc",
			dateStr:  "2025-07-01",
			wantCode: http.StatusBadRequest,
			wantBody: "Error parse doctorID",
		},
		{
			name:     "empty date param",
			doctorID: "2",
			dateStr:  "",
			wantCode: http.StatusBadRequest,
			wantBody: "Date parameter is required",
		},
		{
			name:     "bad date format",
			doctorID: "2",
			dateStr:  "2025.07.01",
			wantCode: http.StatusBadRequest,
			wantBody: "Invalid date format",
		},
		{
			name:     "doctor not found",
			doctorID: "2",
			dateStr:  "2025-07-01",
			mockGetDoctor: func(id int) (domain.Doctor, error) {
				return domain.Doctor{}, repository.ErrorDoctorNotFound
			},
			wantCode: http.StatusNotFound,
			wantBody: "Doctor not found",
		},
		{
			name:     "db error get doctor",
			doctorID: "2",
			dateStr:  "2025-07-01",
			mockGetDoctor: func(id int) (domain.Doctor, error) {
				return domain.Doctor{}, errors.New("fail")
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "Error get doctor",
		},
		{
			name:          "db error get schedule",
			doctorID:      "2",
			dateStr:       "2025-07-01",
			mockGetDoctor: func(id int) (domain.Doctor, error) { return doctor, nil },
			mockGetRecords: func(id int, date time.Time) ([]domain.Record, error) {
				return nil, errors.New("fail schedule")
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "Error get doctor schedule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?date="+tt.dateStr, nil)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("doctorID", tt.doctorID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			mock := &MockControlDoctorsWrapper{
				GetDoctorByIdFn:        tt.mockGetDoctor,
				GetScheduleForDoctorFn: tt.mockGetRecords,
			}
			handler := GetScheduleDoctorByIdHandler(logger, mock)
			handler(w, req)
			assert.Equal(t, tt.wantCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
		})
	}
}

func TestNewScheduleHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	date := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	start := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)
	schedule := domain.Schedule{Records: []domain.Record{
		{
			ID:          11,
			DoctorId:    2,
			Date:        date,
			Start:       start,
			End:         end,
			IsAvailable: true,
		},
	}}
	tests := []struct {
		name               string
		doctorID           string
		dateStr            string
		startTimeStr       string
		endTimeStr         string
		receptionTimeStr   string
		mockCreateSchedule func(int, time.Time, time.Time, time.Time, int) (domain.Schedule, error)
		mockSaveSchedule   func(int, []domain.Record) error
		wantCode           int
		wantBody           string
	}{
		{
			name:             "успех",
			doctorID:         "2",
			dateStr:          "2025-07-01",
			startTimeStr:     "09:00:00",
			endTimeStr:       "10:00:00",
			receptionTimeStr: "20",
			mockCreateSchedule: func(id int, date, start, end time.Time, reception int) (domain.Schedule, error) {
				return schedule, nil
			},
			mockSaveSchedule: func(id int, recs []domain.Record) error { return nil },
			wantCode:         http.StatusOK,
			wantBody:         `"recordID":11`,
		},
		{
			name:             "bad date",
			doctorID:         "2",
			dateStr:          "2025.07.01",
			startTimeStr:     "09:00:00",
			endTimeStr:       "10:00:00",
			receptionTimeStr: "20",
			wantCode:         http.StatusBadRequest,
			wantBody:         "Error parse date",
		},
		{
			name:             "bad start_time",
			doctorID:         "2",
			dateStr:          "2025-07-01",
			startTimeStr:     "bad",
			endTimeStr:       "10:00:00",
			receptionTimeStr: "20",
			wantCode:         http.StatusBadRequest,
			wantBody:         "Error parse start_time",
		},
		{
			name:             "bad end_time",
			doctorID:         "2",
			dateStr:          "2025-07-01",
			startTimeStr:     "09:00:00",
			endTimeStr:       "bad",
			receptionTimeStr: "20",
			wantCode:         http.StatusBadRequest,
			wantBody:         "Error parse end_time",
		},
		{
			name:             "bad reception_time",
			doctorID:         "2",
			dateStr:          "2025-07-01",
			startTimeStr:     "09:00:00",
			endTimeStr:       "10:00:00",
			receptionTimeStr: "bad",
			wantCode:         http.StatusBadRequest,
			wantBody:         "Error parse reception_time",
		},
		{
			name:             "fail create schedule",
			doctorID:         "2",
			dateStr:          "2025-07-01",
			startTimeStr:     "09:00:00",
			endTimeStr:       "10:00:00",
			receptionTimeStr: "20",
			mockCreateSchedule: func(id int, date, start, end time.Time, reception int) (domain.Schedule, error) {
				return domain.Schedule{}, errors.New("fail create")
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "Error create doctor schedule",
		},
		{
			name:             "fail save schedule",
			doctorID:         "2",
			dateStr:          "2025-07-01",
			startTimeStr:     "09:00:00",
			endTimeStr:       "10:00:00",
			receptionTimeStr: "20",
			mockCreateSchedule: func(id int, date, start, end time.Time, reception int) (domain.Schedule, error) {
				return schedule, nil
			},
			mockSaveSchedule: func(id int, recs []domain.Record) error { return errors.New("fail save") },
			wantCode:         http.StatusInternalServerError,
			wantBody:         "Error create doctor schedule",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Формируем реквест с нужными query params
			req := httptest.NewRequest(http.MethodPost, "/?date="+tt.dateStr+"&start_time="+tt.startTimeStr+
				"&end_time="+tt.endTimeStr+"&reception_time="+tt.receptionTimeStr, nil)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("doctorID", tt.doctorID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			mockWrapper := &MockNewScheduleWrapper{
				CreateScheduleForDoctorByIdFn: tt.mockCreateSchedule,
			}
			mockDB := &MockControlDoctorsWrapper{
				CreateNewScheduleForDoctorFn: tt.mockSaveSchedule,
			}
			handler := NewScheduleHandler(logger, mockDB, mockWrapper)
			handler(w, req)
			assert.Equal(t, tt.wantCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
		})
	}
}

func TestGetPolyclinicInfoHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	// Успех
	t.Run("success", func(t *testing.T) {
		mock := &MockSpecializationWrapper{
			GetAllSpecializationsFn: func() ([]domain.Specialization, error) {
				return []domain.Specialization{{ID: 1, SpecializationDoctor: "Терапевт"}}, nil
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler := GetPolyclinicInfoHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("want %v, got %v", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Терапевт") {
			t.Errorf("response body missing specialization")
		}
	})

	// Ошибка сервиса
	t.Run("service error", func(t *testing.T) {
		mock := &MockSpecializationWrapper{
			GetAllSpecializationsFn: func() ([]domain.Specialization, error) {
				return nil, errors.New("fail")
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler := GetPolyclinicInfoHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("want %v, got %v", http.StatusInternalServerError, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error get specialization") {
			t.Errorf("response body missing error message")
		}
	})

	// Пустой список
	t.Run("nil specialization list", func(t *testing.T) {
		mock := &MockSpecializationWrapper{
			GetAllSpecializationsFn: func() ([]domain.Specialization, error) {
				return nil, nil
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler := GetPolyclinicInfoHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("want %v, got %v", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), "[]") {
			t.Errorf("response body should be empty array")
		}
	})
}

// -----------------------------
// GET /specializations/{specializationID}/doctors
// -----------------------------
func TestGetSpecializationDoctorHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	t.Run("success", func(t *testing.T) {
		mock := &MockSpecializationWrapper{
			GetSpecializationAllDoctorFn: func(id int) ([]domain.Doctor, error) {
				return []domain.Doctor{{Id: 3, Name: "Доктор Хаус"}}, nil
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("specializationID", "5")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		handler := GetSpecializationDoctorHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("want %v, got %v", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Доктор Хаус") {
			t.Errorf("response body missing doctor")
		}
	})

	t.Run("bad specialization id", func(t *testing.T) {
		mock := &MockSpecializationWrapper{}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("specializationID", "badid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		handler := GetSpecializationDoctorHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("want %v, got %v", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error parse specializationID") {
			t.Errorf("response body missing error message")
		}
	})

	t.Run("service error", func(t *testing.T) {
		mock := &MockSpecializationWrapper{
			GetSpecializationAllDoctorFn: func(id int) ([]domain.Doctor, error) {
				return nil, errors.New("fail")
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("specializationID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		handler := GetSpecializationDoctorHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("want %v, got %v", http.StatusInternalServerError, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error get list doctors") {
			t.Errorf("response body missing error message")
		}
	})

	t.Run("nil doctors list", func(t *testing.T) {
		mock := &MockSpecializationWrapper{
			GetSpecializationAllDoctorFn: func(id int) ([]domain.Doctor, error) {
				return nil, nil
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("specializationID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		handler := GetSpecializationDoctorHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("want %v, got %v", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), "[]") {
			t.Errorf("response body should be empty array")
		}
	})
}

// -----------------------------
// POST /specializations
// -----------------------------
func TestCreateNewSpecializationHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	t.Run("success", func(t *testing.T) {
		mock := &MockSpecializationWrapper{
			CreateNewSpecializationFn: func(s domain.Specialization) (int, error) {
				return 77, nil
			},
		}
		spec := domain.Specialization{ID: 0, SpecializationDoctor: "Лор"}
		body, _ := json.Marshal(spec)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler := CreateNewSpecializationHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("want %v, got %v", http.StatusCreated, w.Code)
		}
		if !strings.Contains(w.Body.String(), "77") {
			t.Errorf("response body missing specialization id")
		}
	})

	t.Run("bad body", func(t *testing.T) {
		mock := &MockSpecializationWrapper{}
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{bad json"))
		w := httptest.NewRecorder()

		handler := CreateNewSpecializationHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("want %v, got %v", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error parse body") {
			t.Errorf("response body missing error message")
		}
	})

	t.Run("service error", func(t *testing.T) {
		mock := &MockSpecializationWrapper{
			CreateNewSpecializationFn: func(s domain.Specialization) (int, error) {
				return 0, errors.New("fail")
			},
		}
		spec := domain.Specialization{SpecializationDoctor: "Лор"}
		body, _ := json.Marshal(spec)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler := CreateNewSpecializationHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("want %v, got %v", http.StatusInternalServerError, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error create specialization") {
			t.Errorf("response body missing error message")
		}
	})
}

// -----------------------------
// DELETE /specializations/{specializationID}
// -----------------------------
func TestDeleteSpecializationHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	t.Run("success deleted", func(t *testing.T) {
		mock := &MockSpecializationWrapper{
			DeleteSpecializationFn: func(id int) (bool, error) {
				return true, nil
			},
		}
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("specializationID", "4")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		handler := DeleteSpecializationHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("want %v, got %v", http.StatusNoContent, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Deleted specialization") {
			t.Errorf("response body missing deleted message")
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock := &MockSpecializationWrapper{
			DeleteSpecializationFn: func(id int) (bool, error) {
				return false, nil
			},
		}
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("specializationID", "4")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		handler := DeleteSpecializationHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("want %v, got %v", http.StatusNotFound, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Not found specialization") {
			t.Errorf("response body missing not found message")
		}
	})

	t.Run("bad id", func(t *testing.T) {
		mock := &MockSpecializationWrapper{}
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("specializationID", "bad")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		handler := DeleteSpecializationHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("want %v, got %v", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error parse specializationID") {
			t.Errorf("response body missing error message")
		}
	})

	t.Run("service error", func(t *testing.T) {
		mock := &MockSpecializationWrapper{
			DeleteSpecializationFn: func(id int) (bool, error) {
				return false, errors.New("fail")
			},
		}
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("specializationID", "4")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		handler := DeleteSpecializationHandler(logger, mock)
		handler(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("want %v, got %v", http.StatusInternalServerError, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Error delete specialization") {
			t.Errorf("response body missing error message")
		}
	})
}
