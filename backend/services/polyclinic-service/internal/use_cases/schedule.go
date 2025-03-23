package use_cases

import (
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/domain"
	"github.com/pkg/errors"
	"log/slog"
	"time"
)

type ScheduleUseCase struct {
	logger *slog.Logger
}

func NewScheduleUseCase(logger *slog.Logger) *ScheduleUseCase {
	return &ScheduleUseCase{logger: logger}
}

type NewScheduleWrapper interface {
	CreateScheduleForDoctorById(
		doctorID int,
		date time.Time,
		startTime time.Time,
		endTime time.Time,
		receptionTime int,
	) (domain.Schedule, error)
}

func (uc *ScheduleUseCase) CreateScheduleForDoctorById(doctorID int,
	date time.Time,
	startTime time.Time,
	endTime time.Time,
	receptionTime int,
) (domain.Schedule, error) {

	uc.logger.Info("CreateScheduleForDoctorById starting...")
	defer uc.logger.Info("CreateScheduleForDoctorById done")

	if startTime.After(endTime) {
		uc.logger.Error("CreateScheduleForDoctorById start time must be before end time")
		return domain.Schedule{}, errors.New("start time cannot be after end time")
	}

	if receptionTime <= 0 {
		uc.logger.Error("CreateScheduleForDoctorById receptionTime must be greater than zero")
		return domain.Schedule{}, errors.New("reception time must be greater than 0")
	}

	var records []domain.Record
	currentTime := startTime
	recordID := 1

	for currentTime.Before(endTime) {
		nextTime := currentTime.Add(time.Duration(receptionTime) * time.Minute)

		if nextTime.After(endTime) {
			break
		}

		record := domain.Record{
			ID:          recordID,
			DoctorId:    doctorID,
			Date:        date,
			Start:       currentTime,
			End:         nextTime,
			IsAvailable: true, // По умолчанию запись доступна
		}

		records = append(records, record)
		recordID++
		currentTime = nextTime
	}

	// uc.logger.Debug("Records", "records", records)

	return domain.Schedule{
		Records: records,
	}, nil
}
