package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/domain"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func (s *Storage) NewDoctor(newDoctor domain.Doctor) (domain.Doctor, error) {
	subQuery := `
	select id from specialization
	where specialization_doctor=$1
`
	var specializationID int
	err := s.pool.QueryRow(context.Background(), subQuery, newDoctor.Specialization).Scan(
		&specializationID,
	)
	s.logger.Info("specializationID", "specializationID", specializationID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Not found specialization=%v", newDoctor.Specialization))
		return domain.Doctor{}, err
	}

	query := `
	INSERT INTO doctors (surname, name, patronymic, specialization_id, education, progress, photo_path)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id
`
	var doctorID int64
	err = s.pool.QueryRow(context.Background(), query,
		newDoctor.Surname,
		newDoctor.Name,
		newDoctor.Patronymic,
		specializationID,
		newDoctor.Education,
		newDoctor.Progress,
		newDoctor.PhotoPath).Scan(&doctorID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error create doctor in database: %s %s %s", newDoctor.Surname, newDoctor.Name, newDoctor.Patronymic))
		return domain.Doctor{}, err
	}

	s.logger.Info("New doctor", "id", doctorID)

	createdDoctor, err := s.GetDoctorById(int(doctorID))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Not found doctor=%v", doctorID))
		return domain.Doctor{}, errors.Wrapf(err, "doctor with id %d not found", doctorID)
	}

	return createdDoctor, nil
}

func (s *Storage) DeleteDoctor(doctorID int) (bool, error) {
	var isDeleted bool

	query := `
	DELETE FROM doctors
	WHERE id = $1
`
	_, err := s.pool.Exec(context.Background(), query, doctorID)
	if err != nil {
		s.logger.Error("Failed to deleted doctor", "doctorID", doctorID, "error", err)
		return false, err
	}
	isDeleted = true

	return isDeleted, nil

}

func (s *Storage) GetDoctorById(doctorID int) (domain.Doctor, error) {
	//err := s.CalculateRating(&doctorID, nil)
	//if err != nil {
	//	s.logger.Error(err.Error())
	//}

	query := `
	SELECT  d.id, 
	        surname, 
	        name, 
	        patronymic, 
	        s.specialization_doctor, 
	        education, 
	        progress, 
	        rating,
	        photo_path
	    FROM doctors d
	    join specialization s on s.id = d.specialization_id
	    WHERE d.id = $1
`
	var doctor domain.Doctor
	var surname, name, patronymic, photoPath sql.NullString
	err := s.pool.QueryRow(context.Background(), query, doctorID).Scan(
		&doctor.Id,
		&surname,
		&name,
		&patronymic,
		&doctor.Specialization,
		&doctor.Education,
		&doctor.Progress,
		&doctor.Rating,
		&photoPath,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Debug("Doctor not found in database")
			return domain.Doctor{}, fmt.Errorf("%w: doctor with id %d not found", repository.ErrorDoctorNotFound, doctorID)

		}
		s.logger.Error(fmt.Sprintf("Not found doctor=%v with err=%s", doctorID, err))
		return domain.Doctor{}, errors.Wrapf(err, "doctor with id %d not found", doctorID)
	}

	if surname.Valid {
		doctor.Surname = surname.String
	} else {
		doctor.Surname = ""
	}

	if name.Valid {
		doctor.Name = name.String
	} else {
		doctor.Name = ""
	}

	if patronymic.Valid {
		doctor.Patronymic = patronymic.String
	} else {
		doctor.Patronymic = ""
	}
	if photoPath.Valid {
		doctor.PhotoPath = photoPath.String
	} else {
		doctor.Patronymic = ""
	}

	return doctor, nil
}
