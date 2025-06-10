package domain

import (
	"testing"
	"time"
)

func TestNewAppointment(t *testing.T) {
	id := 1
	doctorID := 2
	patientID := 3
	date := time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC)
	timeVal := time.Date(1, 1, 1, 9, 30, 0, 0, time.UTC)
	status := "SCHEDULED"
	rating := 4.8

	app := NewAppointment(id, doctorID, patientID, date, timeVal, status, rating)
	if app == nil {
		t.Fatal("NewAppointment вернул nil")
	}
	if app.Id != id {
		t.Errorf("Ожидал Id %d, получил %d", id, app.Id)
	}
	if app.DoctorID != doctorID {
		t.Errorf("Ожидал DoctorID %d, получил %d", doctorID, app.DoctorID)
	}
	if app.PatientID != patientID {
		t.Errorf("Ожидал PatientID %d, получил %d", patientID, app.PatientID)
	}
	if !app.Date.Equal(date) {
		t.Errorf("Ожидал Date %v, получил %v", date, app.Date)
	}
	if !app.Time.Equal(timeVal) {
		t.Errorf("Ожидал Time %v, получил %v", timeVal, app.Time)
	}
	if app.Status != status {
		t.Errorf("Ожидал Status %s, получил %s", status, app.Status)
	}
	if app.Rating != rating {
		t.Errorf("Ожидал Rating %v, получил %v", rating, app.Rating)
	}
}

func TestNewAppointment_EmptyFields(t *testing.T) {
	app := NewAppointment(0, 0, 0, time.Time{}, time.Time{}, "", 0)
	if app == nil {
		t.Fatal("NewAppointment вернул nil для пустых значений")
	}
	if app.Id != 0 {
		t.Errorf("Ожидал Id 0, получил %d", app.Id)
	}
	if app.DoctorID != 0 {
		t.Errorf("Ожидал DoctorID 0, получил %d", app.DoctorID)
	}
	if app.PatientID != 0 {
		t.Errorf("Ожидал PatientID 0, получил %d", app.PatientID)
	}
	if !app.Date.IsZero() {
		t.Errorf("Ожидал пустой Date, получил %v", app.Date)
	}
	if !app.Time.IsZero() {
		t.Errorf("Ожидал пустой Time, получил %v", app.Time)
	}
	if app.Status != "" {
		t.Errorf("Ожидал пустой Status, получил %q", app.Status)
	}
	if app.Rating != 0 {
		t.Errorf("Ожидал Rating 0, получил %v", app.Rating)
	}
}
