package postgres

import (
	"context"
	"fmt"
	"github.com/daariikk/MyHelp/services/appointment-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"log/slog"
	"time"
)

type Storage struct {
	connection *pgx.Conn
	logger     *slog.Logger
}

func New(ctx context.Context, logger *slog.Logger, url string) (*Storage, error) {
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		logger.Error("Failed to connect to postgres", "error", err)
		return nil, errors.Wrapf(err, "failed to connect to postgres")
	}

	return &Storage{conn, logger}, nil
}

func (s *Storage) Close() error {
	if s.connection != nil {
		return s.connection.Close(context.Background())
	}
	return nil
}

func (s *Storage) checkAvailableRecord(doctorID int, date time.Time, time time.Time) (bool, error) {
	var isAvailable bool

	query := `
	select is_available from doctor_schedules
	where doctor_id = $1 and date = $2 and start_time = $3 
`
	err := s.connection.QueryRow(context.Background(), query,
		doctorID,
		date,
		time).Scan(&isAvailable)
	if err != nil {
		s.logger.Error("impossible to check the availability of the record")
		return false, err
	}
	return isAvailable, nil
}

func (s *Storage) NewAppointment(appointment domain.Appointment) error {
	isAvailable, err := s.checkAvailableRecord(appointment.DoctorID, appointment.Date, appointment.Time)
	if err != nil {
		return err
	}

	if !isAvailable {
		s.logger.Error(fmt.Sprintf("Recording for doctorID=%v date=%v and time=%v is not available because it is already busy", appointment.DoctorID, appointment.Date, appointment.Time))
		return errors.New("Record is busy")
	}

	tx, err := s.connection.Begin(context.Background())
	if err != nil {
		s.logger.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()

	query := `
	INSERT INTO appointments (doctor_id, patient_id, date, time, status_id)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
`
	var appointmentID int
	err = s.connection.QueryRow(context.Background(), query,
		appointment.DoctorID,
		appointment.PatientID,
		appointment.Date,
		appointment.Time,
		domain.SCHEDULED).Scan(&appointmentID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error create appointment in database: doctorID=%v patientID=%v date=%v time=%v",
			appointment.DoctorID,
			appointment.PatientID,
			appointment.Date,
			appointment.Time))
		return err
	}
	updateQuery := `
        UPDATE doctor_schedules
        SET is_available = false
        WHERE doctor_id = $1 
        AND date = $2 
        AND start_time = $3 
    `
	_, err = tx.Exec(context.Background(), updateQuery,
		appointment.DoctorID,
		appointment.Date,
		appointment.Time)
	if err != nil {
		s.logger.Error("Failed to update doctor schedule availability",
			"doctorID", appointment.DoctorID,
			"date", appointment.Date,
			"time", appointment.Time,
			"error", err)
		return err
	}

	// Завершаем транзакцию
	err = tx.Commit(context.Background())
	if err != nil {
		s.logger.Error("Failed to commit transaction", "error", err)
		return err
	}

	s.logger.Info("New appointment", "id", appointmentID)

	return nil
}

func (s *Storage) UpdateAppointment(appointment domain.Appointment) error {
	query := `
	UPDATE appointments
	SET rating = $1
	WHERE id = $2
`

	_, err := s.connection.Exec(context.Background(), query, appointment.Rating, appointment.Id)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error update rating for appointment with id=%v", appointment.Id))
		return err
	}

	return nil

}

func (s *Storage) DeleteAppointment(appointmentID int) error {
	query := `
	UPDATE appointments
	SET status_id = $2
	WHERE id = $1
`
	tx, err := s.connection.Begin(context.Background())
	if err != nil {
		s.logger.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()
	_, err = s.connection.Exec(context.Background(), query, appointmentID, domain.CANCELED)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error delete appointment with id=%v in database", appointmentID))
		return err
	}

	appointment, err := s.GetAppointment(appointmentID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error get appointment with id=%v in database", appointmentID))
		return err
	}

	updateQuery := `
        UPDATE doctor_schedules
        SET is_available = true
        WHERE doctor_id = $1 
        AND date = $2 
        AND start_time <= $3 
    `
	_, err = tx.Exec(context.Background(), updateQuery,
		appointment.DoctorID,
		appointment.Date,
		appointment.Time)
	if err != nil {
		s.logger.Error("Failed to update doctor schedule availability",
			"doctorID", appointment.DoctorID,
			"date", appointment.Date,
			"time", appointment.Time,
			"error", err)
		return err
	}

	// Завершаем транзакцию
	err = tx.Commit(context.Background())
	if err != nil {
		s.logger.Error("Failed to commit transaction", "error", err)
		return err
	}
	return nil
}

func (s *Storage) GetAppointment(appointmentID int) (*domain.Appointment, error) {
	var appointment domain.Appointment
	query := `
	select id, doctor_id, patient_id, date, time from appointments
	where id = $1
`
	err := s.connection.QueryRow(context.Background(), query, appointmentID).Scan(
		&appointment.Id,
		&appointment.DoctorID,
		&appointment.PatientID,
		&appointment.Date,
		&appointment.Time)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error get appointment with id=%v in database", appointmentID))
		return nil, err
	}

	return &appointment, nil
}
