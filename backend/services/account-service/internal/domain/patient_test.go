package domain

import (
	"encoding/json"
	"testing"
	"time"
)

// Тест конструктора для Patient
func TestNewPatient(t *testing.T) {
	p := NewPatient(1, "Иванов", "Иван", "Иванович", "ABCD1234", "ivan@example.com", "secret", true)
	if p.Id != 1 || p.Surname != "Иванов" || p.Name != "Иван" || p.IsDeleted != true {
		t.Error("Конструктор NewPatient заполнил не все поля корректно")
	}
}

// Тест маршалинга Patient с заполнением через конструктор
func TestPatient_JSONMarshalling(t *testing.T) {
	p := NewPatient(123, "Иванов", "Иван", "Иванович", "ABCD1234", "ivan@example.com", "secret", true)

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("не удалось маршалить Patient: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("не удалось анмаршалить JSON: %v", err)
	}

	if m["patientID"].(float64) != 123 {
		t.Errorf("ожидали patientID=123, получили %v", m["patientID"])
	}
	if m["is_deleted"].(bool) != true {
		t.Errorf("ожидали is_deleted=true, получили %v", m["is_deleted"])
	}
}

// Тест конструктора для PatientDTO
func TestNewPatientDTO(t *testing.T) {
	a := AppointmentDTO{
		Id: 1,
		// Добавьте другие необходимые поля, если они есть
	}
	dto := NewPatientDTO(2, "Петров", "Пётр", "Петрович", "XYZ9876", "petr@example.com", "pass", false, []AppointmentDTO{a})
	if dto.Id != 2 || dto.Surname != "Петров" || len(dto.Appointments) != 1 {
		t.Error("Конструктор NewPatientDTO заполнил не все поля корректно")
	}
}

// Тест маршалинга PatientDTO с вложенным appointment
func TestPatientDTO_JSONMarshalling(t *testing.T) {
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	timeStr := now.Format("15:04:05")
	app := AppointmentDTO{
		Id:                   1,
		DoctorFIO:            "Доктор",
		DoctorSpecialization: "Терапевт",
		Date:                 dateStr,
		Time:                 timeStr,
		Status:               "SCHEDULED",
		Rating:               4.5,
	}
	dto := NewPatientDTO(2, "Петров", "Пётр", "Петрович", "XYZ9876", "petr@example.com", "pass", false, []AppointmentDTO{app})

	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("не удалось маршалить PatientDTO: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("не удалось анмаршалить JSON: %v", err)
	}

	if int(m["patientID"].(float64)) != 2 {
		t.Errorf("ожидали patientID=2, получили %v", m["patientID"])
	}
	apps, ok := m["appointments"].([]interface{})
	if !ok || len(apps) != 1 {
		t.Fatalf("ожидали 1 appointment, получили %v", m["appointments"])
	}
	appMap, ok := apps[0].(map[string]interface{})
	if !ok {
		t.Fatalf("не удалось привести appointment к map")
	}
	if appMap["doctor_fio"].(string) != "Доктор" {
		t.Errorf("ожидали doctor_fio='Доктор', получили %v", appMap["doctor_fio"])
	}
}
