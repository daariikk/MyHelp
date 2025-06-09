package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/domain"
	"github.com/pkg/errors"
)

func (s *Storage) GetAllSpecializations() ([]domain.Specialization, error) {
	query := `
		SELECT id, specialization, specialization_doctor, description
		FROM specialization 
`
	var specializations []domain.Specialization

	rows, err := s.pool.Query(context.Background(), query)
	if err != nil {
		if rows != nil {
			rows.Close()
		}
		s.logger.Error(fmt.Sprintf("Error execution sql query: %v", err))
		return nil, errors.Wrapf(err, "Error executing sql query: %v", query)
	}

	defer rows.Close()
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

	// Проверяем ошибки, которые могли возникнуть во время итерации
	if err := rows.Err(); err != nil {
		s.logger.Error(fmt.Sprintf("Error during rows iteration: %v", err))
		return nil, errors.Wrap(err, "Error during rows iteration")
	}

	return specializations, nil
}

func (s *Storage) GetSpecializationAllDoctor(specializationID int) ([]domain.Doctor, error) {
	//err := s.CalculateRating(nil, &specializationID)
	//if err != nil {
	//	s.logger.Error(fmt.Sprintf("Error calculating rating: %v", err))
	//}
	query := `
		SELECT doctors.id, 
		       surname,
		       name, 
		       patronymic, 
		       specialization.specialization_doctor, 
		       education, 
		       progress, 
		       rating,
		       photo_path
		FROM doctors
		JOIN specialization ON doctors.specialization_id = specialization.id
		WHERE specialization_id = $1
`
	var doctors []domain.Doctor
	rows, err := s.pool.Query(context.Background(), query, specializationID)
	if err != nil {
		if rows != nil {
			rows.Close()
		}
		s.logger.Error(fmt.Sprintf("Error execution sql query: %v", err))
		return nil, errors.Wrapf(err, "Error executing sql query: %v", query)
	}
	defer rows.Close()

	for rows.Next() {
		var doctor domain.Doctor
		var education sql.NullString
		var progress sql.NullString
		var photoPath sql.NullString

		err = rows.Scan(
			&doctor.Id,
			&doctor.Surname,
			&doctor.Name,
			&doctor.Patronymic,
			&doctor.Specialization,
			&education,
			&progress,
			&doctor.Rating,
			&photoPath,
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
		if photoPath.Valid {
			doctor.PhotoPath = photoPath.String
		}

		doctors = append(doctors, doctor)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error(fmt.Sprintf("Error during rows iteration: %v", err))
		return nil, errors.Wrap(err, "Error during rows iteration")
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

	err := s.pool.QueryRow(context.Background(), query,
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
	_, err := s.pool.Exec(context.Background(), query, specializationID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error executing sql query: %v", err))
		return false, errors.Wrapf(err, "Error executing sql query: %v", query)
	}

	return true, nil
}
