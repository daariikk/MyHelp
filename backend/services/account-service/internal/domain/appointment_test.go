package domain

import (
	"encoding/json"
	"testing"
	"time"
)

// Тест конструктора для Appointment
func TestNewAppointment(t *testing.T) {
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	timeVal := time.Date(0, 1, 1, 14, 30, 0, 0, time.UTC)
	a := NewAppointment(1, "Доктор Кто", "Терапевт", date, timeVal, "SCHEDULED", 4.9)

	if a.Id != 1 || a.DoctorFIO != "Доктор Кто" || a.DoctorSpecialization != "Терапевт" {
		t.Error("NewAppointment некорректно заполняет поля")
	}
	if a.Date != date || a.Time.Hour() != 14 || a.Time.Minute() != 30 {
		t.Error("NewAppointment неверно заполняет Date/Time")
	}
	if a.Status != "SCHEDULED" || a.Rating != 4.9 {
		t.Error("NewAppointment некорректно заполняет Status/Rating")
	}
}

// Тест маршалинга Appointment
func TestAppointment_JSONMarshalling(t *testing.T) {
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	timeVal := time.Date(0, 1, 1, 14, 30, 0, 0, time.UTC)
	a := NewAppointment(2, "Доктор Стрэндж", "Хирург", date, timeVal, "COMPLETED", 5.0)

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("не удалось маршалить Appointment: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("не удалось анмаршалить JSON: %v", err)
	}

	if int(m["appointmentID"].(float64)) != 2 {
		t.Errorf("ожидали appointmentID=2, получили %v", m["appointmentID"])
	}
	if m["doctor_fio"].(string) != "Доктор Стрэндж" {
		t.Errorf("ожидали doctor_fio='Доктор Стрэндж', получили %v", m["doctor_fio"])
	}
}

// Тест конструктора для AppointmentDTO
func TestNewAppointmentDTO(t *testing.T) {
	a := NewAppointmentDTO(3, "Доктор Хаус", "Диагност", "2024-06-01", "09:00:00", "CANCELED", 3.7)

	if a.Id != 3 || a.DoctorFIO != "Доктор Хаус" || a.DoctorSpecialization != "Диагност" {
		t.Error("NewAppointmentDTO некорректно заполняет поля")
	}
	if a.Date != "2024-06-01" || a.Time != "09:00:00" {
		t.Error("NewAppointmentDTO некорректно заполняет Date/Time")
	}
	if a.Status != "CANCELED" || a.Rating != 3.7 {
		t.Error("NewAppointmentDTO некорректно заполняет Status/Rating")
	}
}

// Тест маршалинга AppointmentDTO
func TestAppointmentDTO_JSONMarshalling(t *testing.T) {
	a := NewAppointmentDTO(4, "Доктор Дулиттл", "Ветеринар", "2024-06-02", "12:00:00", "SCHEDULED", 4.1)
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("не удалось маршалить AppointmentDTO: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("не удалось анмаршалить JSON: %v", err)
	}
	if int(m["id"].(float64)) != 4 {
		t.Errorf("ожидали id=4, получили %v", m["id"])
	}
	if m["doctor_fio"].(string) != "Доктор Дулиттл" {
		t.Errorf("ожидали doctor_fio='Доктор Дулиттл', получили %v", m["doctor_fio"])
	}
}
