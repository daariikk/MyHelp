package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/domain"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockRegisterWrapper struct {
	RegisterUserFn func(user domain.User) (domain.User, error)
}

func (m *MockRegisterWrapper) RegisterUser(user domain.User) (domain.User, error) {
	return m.RegisterUserFn(user)
}

func TestRegisterHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	validUser := domain.User{
		Surname:    "Иванов",
		Name:       "Иван",
		Patronymic: "Иванович",
		Polic:      "12345",
		Email:      "test@mail.com",
		Password:   "123456",
	}

	t.Run("bad body", func(t *testing.T) {
		mock := &MockRegisterWrapper{}
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("bad_json"))
		w := httptest.NewRecorder()
		handler := RegisterHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Invalid request body") {
			t.Errorf("ответ не содержит Invalid request body")
		}
	})

	t.Run("missing email", func(t *testing.T) {
		user := validUser
		user.Email = ""
		body, _ := json.Marshal(user)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mock := &MockRegisterWrapper{}
		handler := RegisterHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Email is required") {
			t.Errorf("ответ не содержит Email is required")
		}
	})

	t.Run("missing polic", func(t *testing.T) {
		user := validUser
		user.Polic = ""
		body, _ := json.Marshal(user)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mock := &MockRegisterWrapper{}
		handler := RegisterHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Polic is required") {
			t.Errorf("ответ не содержит Polic is required")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		user := validUser
		user.Name = ""
		body, _ := json.Marshal(user)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mock := &MockRegisterWrapper{}
		handler := RegisterHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Name is required") {
			t.Errorf("ответ не содержит Name is required")
		}
	})

	t.Run("missing password", func(t *testing.T) {
		user := validUser
		user.Password = ""
		body, _ := json.Marshal(user)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mock := &MockRegisterWrapper{}
		handler := RegisterHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("ожидал %d, получил %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Password is required") {
			t.Errorf("ответ не содержит Password is required")
		}
	})

	t.Run("internal error on register", func(t *testing.T) {
		user := validUser
		body, _ := json.Marshal(user)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mock := &MockRegisterWrapper{
			RegisterUserFn: func(user domain.User) (domain.User, error) {
				return domain.User{}, errors.New("fail to create user")
			},
		}
		handler := RegisterHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("ожидал %d, получил %d", http.StatusInternalServerError, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Failed to create user") {
			t.Errorf("ответ не содержит Failed to create user")
		}
	})

	t.Run("success", func(t *testing.T) {
		user := validUser
		body, _ := json.Marshal(user)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mock := &MockRegisterWrapper{
			RegisterUserFn: func(user domain.User) (domain.User, error) {
				user.Id = 42
				return user, nil
			},
		}
		handler := RegisterHandler(logger, mock)
		handler(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("ожидал %d, получил %d", http.StatusCreated, w.Code)
		}
		if !strings.Contains(w.Body.String(), `"patientID":42`) {
			t.Errorf("ответ не содержит patientID")
		}
		if strings.Contains(w.Body.String(), validUser.Password) {
			t.Errorf("пароль не должен возвращаться")
		}
	})
}
