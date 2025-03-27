package postgres

import (
	"context"
	"fmt"
	"github.com/daariikk/MyHelp/services/appointment-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"log/slog"
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

func (s *Storage) NewAppointment(appointment domain.Appointment) error {
	query := `
	INSERT INTO appointments (doctor_id, patient_id, date, time, status_id)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
`
	var appointmentID int
	err := s.connection.QueryRow(context.Background(), query,
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
	_, err := s.connection.Exec(context.Background(), query, appointmentID, domain.CANCELED)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error delete appointment with id=%v in database", appointmentID))
		return err
	}

	return nil
}
