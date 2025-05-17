package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/daariikk/MyHelp/services/account-service/internal/domain"
	"github.com/daariikk/MyHelp/services/account-service/internal/repository"
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

func (s *Storage) GetPatientById(patientID int) (domain.Patient, error) {
	query := `
		SELECT id, surname, name, patronymic, email, polic, is_deleted
		FROM patients 
		WHERE id=$1
`
	var patient domain.Patient
	var surname, name, patronymic sql.NullString
	err := s.connection.QueryRow(context.Background(), query, patientID).Scan(
		&patient.Id,
		&surname,
		&name,
		&patronymic,
		&patient.Email,
		&patient.Polic,
		&patient.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Debug("Patient not found in database")
			return domain.Patient{}, fmt.Errorf("%w: patient with id %d not found", repository.ErrorNotFound, patientID)
		}
		return domain.Patient{}, errors.Wrapf(err, "failed to get patient with id %d", patientID)
	}

	if surname.Valid {
		patient.Surname = surname.String
	} else {
		patient.Surname = ""
	}

	if name.Valid {
		patient.Name = name.String
	} else {
		patient.Name = ""
	}

	if patronymic.Valid {
		patient.Patronymic = patronymic.String
	} else {
		patient.Patronymic = ""
	}

	return patient, nil
}

func (s *Storage) GetAppointmentByPatientId(patientID int) ([]domain.Appointment, error) {
	s.logger.Debug("GetAppointmentByPatientId starting...")

	// Сначала обновляем статусы прошедших записей для этого пациента
	updateQuery := `
        UPDATE appointments
        SET status_id = $2
        WHERE patient_id = $1
        AND (date < CURRENT_DATE OR (date = CURRENT_DATE AND time < CURRENT_TIME))
        AND status_id NOT IN (2, 3)  -- Не обновляем уже завершенные или отмененные
    `

	_, err := s.connection.Exec(context.Background(), updateQuery, patientID, domain.COMPLETED)
	if err != nil {
		s.logger.Error("Failed to update appointment statuses", "error", err)
		return nil, fmt.Errorf("failed to update appointment statuses: %w", err)
	}

	query := `
        SELECT appointments.id, 
               CONCAT(doctors.surname, ' ', doctors.name, ' ', doctors.patronymic) AS doctor_fio,
               specialization.specialization_doctor AS doctor_specialization, 
               appointments.date, 
               appointments.time, 
               status_appointment.code, 
               appointments.rating
        FROM appointments
        JOIN doctors ON appointments.doctor_id = doctors.id
        JOIN specialization ON doctors.specialization_id = specialization.id
        JOIN status_appointment ON appointments.status_id = status_appointment.id
        WHERE appointments.patient_id=$1
    `

	s.logger.Debug("Executing query with patientID", "patientID", patientID)

	rows, err := s.connection.Query(context.Background(), query, patientID)
	if err != nil {
		s.logger.Error("Failed to execute query", "error", err)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	appointments := make([]domain.Appointment, 0)

	for rows.Next() {
		app := domain.Appointment{}
		var rating sql.NullFloat64

		err := rows.Scan(
			&app.Id,
			&app.DoctorFIO,
			&app.DoctorSpecialization,
			&app.Date,
			&app.Time,
			&app.Status,
			&rating,
		)
		if err != nil {
			s.logger.Error("Failed to scan row", "error", err)
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if rating.Valid {
			app.Rating = rating.Float64
		} else {
			app.Rating = 5.0
		}

		appointments = append(appointments, app)
	}

	s.logger.Debug("Found appointments", "count", len(appointments))

	if err = rows.Err(); err != nil {
		s.logger.Error("Error iterating rows", "error", err)
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return appointments, nil
}
