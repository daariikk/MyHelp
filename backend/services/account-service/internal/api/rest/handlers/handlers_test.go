package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/daariikk/MyHelp/services/account-service/internal/domain"
	"github.com/daariikk/MyHelp/services/account-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockGetPatientWrapper struct {
	GetPatientByIdFn            func(int) (domain.Patient, error)
	GetAppointmentByPatientIdFn func(int) ([]domain.Appointment, error)
}

func (m *MockGetPatientWrapper) GetPatientById(id int) (domain.Patient, error) {
	return m.GetPatientByIdFn(id)
}
func (m *MockGetPatientWrapper) GetAppointmentByPatientId(id int) ([]domain.Appointment, error) {
	return m.GetAppointmentByPatientIdFn(id)
}

type MockDeletePatientWrapper struct {
	DeletePatientByIdFn func(int) (bool, error)
}

func (m *MockDeletePatientWrapper) DeletePatientById(id int) (bool, error) {
	return m.DeletePatientByIdFn(id)
}

type MockUpdatePatientWrapper struct {
	UpdatePatientByIdFn func(domain.Patient) (domain.Patient, error)
}

func (m *MockUpdatePatientWrapper) UpdatePatientById(p domain.Patient) (domain.Patient, error) {
	return m.UpdatePatientByIdFn(p)
}

func TestGetPatientByIdHandler(t *testing.T) {
	now := time.Now()
	mockPatient := domain.Patient{
		Id:         42,
		Surname:    "Иванов",
		Name:       "Иван",
		Patronymic: "Иванович",
		Polic:      "123456",
		Email:      "test@example.com",
		IsDeleted:  false,
	}

	mockAppointment := domain.Appointment{
		Id:                   1,
		DoctorFIO:            "Dr. Smith",
		DoctorSpecialization: "Cardiology",
		Date:                 now,
		Time:                 now,
		Status:               "COMPLETED",
		Rating:               4.5,
	}

	tests := []struct {
		name           string
		patientID      string
		setupWrapper   func() *MockGetPatientWrapper
		expectedStatus int
		expectedInBody string
	}{
		{
			name:      "success",
			patientID: "42",
			setupWrapper: func() *MockGetPatientWrapper {
				return &MockGetPatientWrapper{
					GetPatientByIdFn: func(id int) (domain.Patient, error) {
						assert.Equal(t, 42, id)
						return mockPatient, nil
					},
					GetAppointmentByPatientIdFn: func(id int) ([]domain.Appointment, error) {
						assert.Equal(t, 42, id)
						return []domain.Appointment{mockAppointment}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			expectedInBody: `"patientID":42`,
		},
		{
			name:      "not found",
			patientID: "777",
			setupWrapper: func() *MockGetPatientWrapper {
				return &MockGetPatientWrapper{
					GetPatientByIdFn: func(id int) (domain.Patient, error) {
						return domain.Patient{}, repository.ErrorNotFound
					},
					GetAppointmentByPatientIdFn: func(id int) ([]domain.Appointment, error) {
						return nil, nil
					},
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedInBody: `not found`,
		},
		{
			name:      "db error",
			patientID: "5",
			setupWrapper: func() *MockGetPatientWrapper {
				return &MockGetPatientWrapper{
					GetPatientByIdFn: func(id int) (domain.Patient, error) {
						return domain.Patient{}, errors.New("fail")
					},
					GetAppointmentByPatientIdFn: func(id int) ([]domain.Appointment, error) {
						return nil, nil
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedInBody: `Failed to get patient`,
		},
		{
			name:      "bad patientID",
			patientID: "abc",
			setupWrapper: func() *MockGetPatientWrapper {
				return &MockGetPatientWrapper{}
			},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: `Invalid patientID`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём реквест
			req := httptest.NewRequest(http.MethodGet, "/?patientID="+tt.patientID, nil)
			w := httptest.NewRecorder()
			logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
			handler := GetPatientByIdHandler(logger, tt.setupWrapper())
			handler(w, req)

			resp := w.Result()
			body := w.Body.String()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, body)
			assert.Contains(t, body, tt.expectedInBody)
		})
	}
}

func TestDeletePatientHandler(t *testing.T) {
	tests := []struct {
		name           string
		patientID      string
		mockWrapper    *MockDeletePatientWrapper
		expectedStatus int
		expectedInBody string
	}{
		{
			name:      "invalid patientID",
			patientID: "abc",
			mockWrapper: &MockDeletePatientWrapper{
				DeletePatientByIdFn: func(id int) (bool, error) {
					panic("should not be called")
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "Invalid patientID",
		},
		{
			name:      "deleted",
			patientID: "42",
			mockWrapper: &MockDeletePatientWrapper{
				DeletePatientByIdFn: func(id int) (bool, error) {
					assert.Equal(t, 42, id)
					return true, nil
				},
			},
			expectedStatus: http.StatusNoContent,
			expectedInBody: "deleted",
		},
		{
			name:      "not found",
			patientID: "101",
			mockWrapper: &MockDeletePatientWrapper{
				DeletePatientByIdFn: func(id int) (bool, error) {
					assert.Equal(t, 101, id)
					return false, nil
				},
			},
			expectedStatus: http.StatusNoContent,
			expectedInBody: "not found",
		},
		{
			name:      "error deleting",
			patientID: "55",
			mockWrapper: &MockDeletePatientWrapper{
				DeletePatientByIdFn: func(id int) (bool, error) {
					assert.Equal(t, 55, id)
					return false, errors.New("db fail")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedInBody: "Error deleting patient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/?patientID="+tt.patientID, nil)
			w := httptest.NewRecorder()
			logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
			handler := DeletePatientHandler(logger, tt.mockWrapper)
			handler(w, req)

			resp := w.Result()
			body := w.Body.String()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, body)
			assert.Contains(t, body, tt.expectedInBody)
		})
	}
}

func TestUpdatePatientInfoHandler(t *testing.T) {
	basePatient := domain.Patient{
		Surname:    "Иванов",
		Name:       "Иван",
		Patronymic: "Иванович",
		Polic:      "123456",
		Email:      "ivan@example.com",
	}

	tests := []struct {
		name           string
		patientID      string
		requestBody    interface{}
		setupWrapper   func() *MockUpdatePatientWrapper
		expectedStatus int
		expectedInBody string
	}{
		{
			name:        "invalid patientID",
			patientID:   "abc",
			requestBody: basePatient,
			setupWrapper: func() *MockUpdatePatientWrapper {
				return &MockUpdatePatientWrapper{
					UpdatePatientByIdFn: func(p domain.Patient) (domain.Patient, error) {
						panic("should not be called")
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "Invalid patientID",
		},
		{
			name:        "invalid json",
			patientID:   "10",
			requestBody: "{not valid json", // строка, не объект
			setupWrapper: func() *MockUpdatePatientWrapper {
				return &MockUpdatePatientWrapper{
					UpdatePatientByIdFn: func(p domain.Patient) (domain.Patient, error) {
						panic("should not be called")
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "Invalid request data",
		},
		{
			name:        "email not unique",
			patientID:   "12",
			requestBody: basePatient,
			setupWrapper: func() *MockUpdatePatientWrapper {
				return &MockUpdatePatientWrapper{
					UpdatePatientByIdFn: func(p domain.Patient) (domain.Patient, error) {
						return domain.Patient{}, repository.ErrorEmailUnique
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "Email already exists",
		},
		{
			name:        "internal error",
			patientID:   "15",
			requestBody: basePatient,
			setupWrapper: func() *MockUpdatePatientWrapper {
				return &MockUpdatePatientWrapper{
					UpdatePatientByIdFn: func(p domain.Patient) (domain.Patient, error) {
						return domain.Patient{}, errors.New("something failed")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedInBody: "Error updating patient",
		},
		{
			name:        "success",
			patientID:   "42",
			requestBody: basePatient,
			setupWrapper: func() *MockUpdatePatientWrapper {
				return &MockUpdatePatientWrapper{
					UpdatePatientByIdFn: func(p domain.Patient) (domain.Patient, error) {
						assert.Equal(t, 42, p.Id)
						return domain.Patient{
							Id:         42,
							Surname:    "Иванов",
							Name:       "Иван",
							Patronymic: "Иванович",
							Polic:      "123456",
							Email:      "ivan@example.com",
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			expectedInBody: `"surname":"Иванов"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				var err error
				bodyBytes, err = json.Marshal(v)
				assert.NoError(t, err)
			}
			req := httptest.NewRequest(http.MethodPut, "/?patientID="+tt.patientID, bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()
			logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
			handler := UpdatePatientInfoHandler(logger, tt.setupWrapper())
			handler(w, req)

			resp := w.Result()
			body := w.Body.String()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, body)
			assert.Contains(t, body, tt.expectedInBody)
		})
	}
}
