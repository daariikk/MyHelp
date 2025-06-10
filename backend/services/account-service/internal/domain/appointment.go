package domain

import "time"

// статусы записи
const (
	SCHEDULED = 1
	COMPLETED = 2
	CANCELED  = 3
)

type Appointment struct {
	Id                   int       `json:"appointmentID"`
	DoctorFIO            string    `json:"doctor_fio"`
	DoctorSpecialization string    `json:"doctor_specialization"`
	Date                 time.Time `json:"date"`
	Time                 time.Time `json:"time"`
	Status               string    `json:"status"`
	Rating               float64   `json:"rating"`
}

type AppointmentDTO struct {
	Id                   int     `json:"id"`
	DoctorFIO            string  `json:"doctor_fio"`
	DoctorSpecialization string  `json:"doctor_specialization"`
	Date                 string  `json:"date"` // Формат YYYY-MM-DD
	Time                 string  `json:"time"` // Формат HH:MM:SS
	Status               string  `json:"status"`
	Rating               float64 `json:"rating"`
}

func NewAppointment(id int, doctorFIO, doctorSpec string, date, timeVal time.Time, status string, rating float64) *Appointment {
	return &Appointment{
		Id:                   id,
		DoctorFIO:            doctorFIO,
		DoctorSpecialization: doctorSpec,
		Date:                 date,
		Time:                 timeVal,
		Status:               status,
		Rating:               rating,
	}
}

func NewAppointmentDTO(id int, doctorFIO, doctorSpec, date, timeStr, status string, rating float64) *AppointmentDTO {
	return &AppointmentDTO{
		Id:                   id,
		DoctorFIO:            doctorFIO,
		DoctorSpecialization: doctorSpec,
		Date:                 date,
		Time:                 timeStr,
		Status:               status,
		Rating:               rating,
	}
}
