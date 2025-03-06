package domain

import "time"

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
