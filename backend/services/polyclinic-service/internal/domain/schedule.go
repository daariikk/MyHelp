package domain

import "time"

type Record struct {
	ID          int       `json:"recordID"`
	DoctorId    int       `json:"doctorID"`
	Date        time.Time `json:"date"`
	Start       time.Time `json:"start_time"`
	End         time.Time `json:"end_time"`
	IsAvailable bool      `json:"is_available"`
}

type Schedule struct {
	Records []Record
}

type ScheduleInfoDTO struct {
	Doctor   Doctor   `json:"doctor"`
	Schedule Schedule `json:"schedule"`
}
