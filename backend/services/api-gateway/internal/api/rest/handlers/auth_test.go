package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/daariikk/MyHelp/services/api-gateway/internal/api/response"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/config"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
)

type MockLoginWrapper struct {
	mock.Mock
}

func (m *MockLoginWrapper) GetPassword(email string) (int, string, error) {
	args := m.Called(email)
	return args.Int(0), args.String(1), args.Error(2)
}

func (m *MockLoginWrapper) GetAdminPassword(email string) (int, string, error) {
	args := m.Called(email)
	return args.Int(0), args.String(1), args.Error(2)
}

func (m *MockLoginWrapper) GetUser(email string) (domain.User, error) {
	args := m.Called(email)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockLoginWrapper) GetAdmin(email string) (domain.Admin, error) {
	args := m.Called(email)
	return args.Get(0).(domain.Admin), args.Error(1)
}

func TestLoginHandler(t *testing.T) {
	logger := slog.Default()
	cfg := &config.Config{
		JWT: config.JWT{
			AccessSecretKey:  "test_access_secret",
			RefreshSecretKey: "test_refresh_secret",
			ExpireAccess:     time.Minute * 15,
			ExpireRefresh:    time.Hour * 24,
		},
	}

	tests := []struct {
		name           string
		input          interface{}
		mockSetup      func(*MockLoginWrapper)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "successful login",
			input: domain.User{
				Email:    "test@example.com",
				Password: "correctpassword",
			},
			mockSetup: func(m *MockLoginWrapper) {
				m.On("GetPassword", "test@example.com").
					Return(1, "correctpassword", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"patientID":        float64(1),
				"access_token":     mock.AnythingOfType("string"),
				"access_lifetime":  mock.AnythingOfType("string"),
				"refresh_token":    mock.AnythingOfType("string"),
				"refresh_lifetime": mock.AnythingOfType("string"),
			},
		},
		{
			name: "invalid json",
			input: `{
				"email": "test@example.com",
				"password": "password123",
			}`, // trailing comma makes it invalid
			mockSetup:      func(m *MockLoginWrapper) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body",
		},
		{
			name: "user not found",
			input: domain.User{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockLoginWrapper) {
				m.On("GetPassword", "nonexistent@example.com").
					Return(0, "", errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: response.FailureResponse{
				Message: "Пользователь с таким email не существует",
			},
		},
		{
			name: "wrong password",
			input: domain.User{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(m *MockLoginWrapper) {
				m.On("GetPassword", "test@example.com").
					Return(1, "correctpassword", nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: response.FailureResponse{
				Message: "Failed to auth user: Неверный пароль",
			},
		},
		{
			name: "internal server error",
			input: domain.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockLoginWrapper) {
				m.On("GetPassword", "test@example.com").
					Return(0, "", errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: response.FailureResponse{
				Message: "Failed to auth user: database error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogin := new(MockLoginWrapper)
			tt.mockSetup(mockLogin)

			handler := LoginHandler(logger, mockLogin, cfg)

			var reqBody []byte
			switch v := tt.input.(type) {
			case string:
				reqBody = []byte(v)
			default:
				reqBody, _ = json.Marshal(v)
			}

			req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusBadRequest && tt.name == "invalid json" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody.(string))
				return
			}

			var responseBody interface{}
			if tt.expectedStatus >= http.StatusBadRequest {
				responseBody = &response.FailureResponse{}
			} else {
				responseBody = &map[string]interface{}{}
			}

			err = json.Unmarshal(rr.Body.Bytes(), responseBody)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				respMap := *responseBody.(*map[string]interface{})
				assert.Equal(t, tt.expectedBody.(map[string]interface{})["patientID"], respMap["patientID"])
				assert.NotEmpty(t, respMap["access_token"])
				assert.NotEmpty(t, respMap["refresh_token"])
			} else {
				assert.Equal(t, tt.expectedBody, responseBody)
			}

			mockLogin.AssertExpectations(t)
		})
	}
}

func TestGetUserHandler(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name           string
		email          string
		mockSetup      func(*MockLoginWrapper)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:  "successful get user",
			email: "test@example.com",
			mockSetup: func(m *MockLoginWrapper) {
				m.On("GetUser", "test@example.com").
					Return(domain.User{
						Email: "test@example.com",
						Name:  "Test User",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: response.SuccessResponse{
				Data: domain.User{
					Email: "test@example.com",
					Name:  "Test User",
				},
			},
		},
		{
			name:           "empty email",
			email:          "",
			mockSetup:      func(m *MockLoginWrapper) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: response.FailureResponse{
				Message: "Email is empty",
			},
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			mockSetup: func(m *MockLoginWrapper) {
				m.On("GetUser", "nonexistent@example.com").
					Return(domain.User{}, pgx.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: response.FailureResponse{
				Message: "Пользователь с таким email не существует",
			},
		},
		{
			name:  "internal server error",
			email: "test@example.com",
			mockSetup: func(m *MockLoginWrapper) {
				m.On("GetUser", "test@example.com").
					Return(domain.User{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: response.FailureResponse{
				Message: "Failed get to user",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogin := new(MockLoginWrapper)
			tt.mockSetup(mockLogin)

			handler := GetUserHandler(logger, mockLogin)

			req, err := http.NewRequest("GET", "/user?email="+tt.email, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var responseBody interface{}
			if tt.expectedStatus >= http.StatusBadRequest {
				responseBody = &response.FailureResponse{}
			} else {
				responseBody = &response.SuccessResponse{}
			}

			err = json.Unmarshal(rr.Body.Bytes(), responseBody)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, responseBody)

			mockLogin.AssertExpectations(t)
		})
	}
}

func TestLoginAdminHandler(t *testing.T) {
	logger := slog.Default()
	cfg := &config.Config{
		JWT: config.JWT{
			AccessSecretKey:  "test_access_secret",
			RefreshSecretKey: "test_refresh_secret",
			ExpireAccess:     time.Minute * 15,
			ExpireRefresh:    time.Hour * 24,
		},
	}

	tests := []struct {
		name           string
		input          interface{}
		mockSetup      func(*MockLoginWrapper)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "successful admin login",
			input: domain.Admin{
				Email:    "admin@example.com",
				Password: "correctpassword",
			},
			mockSetup: func(m *MockLoginWrapper) {
				m.On("GetAdminPassword", "admin@example.com").
					Return(1, "correctpassword", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"adminID":          float64(1),
				"access_token":     mock.AnythingOfType("string"),
				"access_lifetime":  mock.AnythingOfType("string"),
				"refresh_token":    mock.AnythingOfType("string"),
				"refresh_lifetime": mock.AnythingOfType("string"),
			},
		},
		{
			name: "invalid json",
			input: `{
				"email": "admin@example.com",
				"password": "password123",
			}`, // trailing comma makes it invalid
			mockSetup:      func(m *MockLoginWrapper) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body",
		},
		{
			name: "wrong password",
			input: domain.Admin{
				Email:    "admin@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(m *MockLoginWrapper) {
				m.On("GetAdminPassword", "admin@example.com").
					Return(1, "correctpassword", nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: response.FailureResponse{
				Message: "Failed to auth user: invalid token",
			},
		},
		{
			name: "internal server error",
			input: domain.Admin{
				Email:    "admin@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockLoginWrapper) {
				m.On("GetAdminPassword", "admin@example.com").
					Return(0, "", errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: response.FailureResponse{
				Message: "Failed to auth user: database error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogin := new(MockLoginWrapper)
			tt.mockSetup(mockLogin)

			handler := LoginAdminHandler(logger, mockLogin, cfg)

			var reqBody []byte
			switch v := tt.input.(type) {
			case string:
				reqBody = []byte(v)
			default:
				reqBody, _ = json.Marshal(v)
			}

			req, err := http.NewRequest("POST", "/admin/login", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusBadRequest && tt.name == "invalid json" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody.(string))
				return
			}

			var responseBody interface{}
			if tt.expectedStatus >= http.StatusBadRequest {
				responseBody = &response.FailureResponse{}
			} else {
				responseBody = &map[string]interface{}{}
			}

			err = json.Unmarshal(rr.Body.Bytes(), responseBody)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				respMap := *responseBody.(*map[string]interface{})
				assert.Equal(t, tt.expectedBody.(map[string]interface{})["adminID"], respMap["adminID"])
				assert.NotEmpty(t, respMap["access_token"])
				assert.NotEmpty(t, respMap["refresh_token"])
			} else {
				assert.Equal(t, tt.expectedBody, responseBody)
			}

			mockLogin.AssertExpectations(t)
		})
	}
}

func TestGetAdminHandler(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name           string
		email          string
		mockSetup      func(*MockLoginWrapper)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:  "successful get admin",
			email: "admin@example.com",
			mockSetup: func(m *MockLoginWrapper) {
				m.On("GetAdmin", "admin@example.com").
					Return(domain.Admin{
						Email:    "admin@example.com",
						Username: "Admin User",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: response.SuccessResponse{
				Data: domain.Admin{
					Email:    "admin@example.com",
					Username: "Admin User",
				},
			},
		},
		{
			name:           "empty email",
			email:          "",
			mockSetup:      func(m *MockLoginWrapper) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: response.FailureResponse{
				Message: "Email is empty",
			},
		},
		{
			name:  "internal server error",
			email: "admin@example.com",
			mockSetup: func(m *MockLoginWrapper) {
				m.On("GetAdmin", "admin@example.com").
					Return(domain.Admin{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: response.FailureResponse{
				Message: "Failed get to admin",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogin := new(MockLoginWrapper)
			tt.mockSetup(mockLogin)

			handler := GetAdminHandler(logger, mockLogin)

			req, err := http.NewRequest("GET", "/admin?email="+tt.email, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var responseBody interface{}
			if tt.expectedStatus >= http.StatusBadRequest {
				responseBody = &response.FailureResponse{}
			} else {
				responseBody = &response.SuccessResponse{}
			}

			err = json.Unmarshal(rr.Body.Bytes(), responseBody)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, responseBody)

			mockLogin.AssertExpectations(t)
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	logger := slog.Default()
	cfg := &config.Config{
		JWT: config.JWT{
			AccessSecretKey: "test_secret_key",
		},
	}

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing authorization header",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
		},
		{
			name:           "invalid token format",
			token:          "InvalidTokenFormat",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid Authorization header format",
		},
		{
			name:           "invalid token",
			token:          "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token",
		},
		{
			name: "valid token",
			token: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"patient_id": 123,
					"exp":        time.Now().Add(time.Hour).Unix(),
				})
				tokenString, _ := token.SignedString([]byte("test_secret_key"))
				return "Bearer " + tokenString
			}(),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем, что patient_id добавлен в контекст
				patientID := r.Context().Value("patient_id")
				if tt.expectedStatus == http.StatusOK {
					assert.Equal(t, int64(123), patientID)
				}
				w.WriteHeader(http.StatusOK)
			})

			middleware := AuthMiddleware(logger, cfg)(nextHandler)

			req, err := http.NewRequest("GET", "/protected", nil)
			assert.NoError(t, err)

			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}

			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestCorsMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:           "regular request",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			},
		},
		{
			name:           "preflight request",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := CorsMiddleware(nextHandler)

			req, err := http.NewRequest(tt.method, "/", nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			for key, value := range tt.expectedHeaders {
				assert.Equal(t, value, rr.Header().Get(key))
			}
		})
	}
}
