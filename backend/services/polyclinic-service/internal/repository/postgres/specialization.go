package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/domain"
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

func (s *Storage) GetAllSpecializations() ([]domain.Specialization, error) {
	query := `
		SELECT id, specialization, specialization_doctor, description
		FROM specialization 
`
	var specializations []domain.Specialization

	rows, err := s.connection.Query(context.Background(), query)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error execution sql query: %v", err))
		return nil, errors.Wrapf(err, "Error executing sql query: %v", query)
	}

	for rows.Next() {
		var specialization domain.Specialization
		err = rows.Scan(
			&specialization.ID,
			&specialization.Specialization,
			&specialization.SpecializationDoctor,
			&specialization.Description,
		)
		specializations = append(specializations, specialization)
	}
	return specializations, nil
}

func (s *Storage) GetSpecializationAllDoctor(specializationID int) ([]domain.Doctor, error) {
	query := `
		SELECT doctors.id, 
		       surname,
		       name, 
		       patronymic, 
		       specialization.specialization_doctor, 
		       education, 
		       progress, 
		       rating
		FROM doctors
		JOIN specialization ON doctors.specialization_id = specialization.id
		WHERE specialization_id = $1
`
	var doctors []domain.Doctor
	rows, err := s.connection.Query(context.Background(), query, specializationID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error execution sql query: %v", err))
		return nil, errors.Wrapf(err, "Error executing sql query: %v", query)
	}
	for rows.Next() {
		var doctor domain.Doctor
		var education sql.NullString
		var progress sql.NullString

		err = rows.Scan(
			&doctor.Id,
			&doctor.Surname,
			&doctor.Name,
			&doctor.Patronymic,
			&doctor.Specialization,
			&education,
			&progress,
			&doctor.Rating,
		)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Error execution sql query: %v", err))
			return nil, errors.Wrapf(err, "Error executing sql query: %v", query)
		}

		if education.Valid {
			doctor.Education = education.String
		}
		if progress.Valid {
			doctor.Progress = progress.String
		}

		doctors = append(doctors, doctor)
	}

	return doctors, nil
}

func (s *Storage) CreateNewSpecialization(specialization domain.Specialization) (int, error) {
	query := `
	INSERT INTO specialization (specialization, specialization_doctor, description ) 
	VALUES ($1, $2, $3)
	RETURNING id
`
	var specializationId int

	err := s.connection.QueryRow(context.Background(), query,
		specialization.Specialization,
		specialization.SpecializationDoctor,
		specialization.Description,
	).Scan(&specializationId)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error execution sql query: %v", err))
		return 0, errors.Wrapf(err, "Error executing sql query: %v", query)
	}
	s.logger.Debug("specializationId", "specializationId", specializationId)

	return specializationId, nil
}
func (s *Storage) DeleteSpecialization(specializationID int) (bool, error) {
	query := `
	DELETE FROM specialization WHERE id = $1
`
	_, err := s.connection.Exec(context.Background(), query, specializationID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error executing sql query: %v", err))
		return false, errors.Wrapf(err, "Error executing sql query: %v", query)
	}

	return true, nil
}
