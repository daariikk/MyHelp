package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/domain"
	"github.com/pkg/errors"
	"time"
)

func (s *Storage) CalculateRating(doctorID *int, specializationID *int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Определяем условие WHERE в зависимости от входных параметров
	var whereClause string
	var args []interface{}

	if doctorID != nil {
		whereClause = "WHERE d.id = $1"
		args = append(args, *doctorID)
	} else if specializationID != nil {
		whereClause = "WHERE d.specialization_id = $1"
		args = append(args, *specializationID)
	} else {
		return errors.New("either doctorID or specializationID must be provided")
	}

	// Получаем список врачей для обновления
	doctorsQuery := `
        SELECT d.id FROM doctors d
        ` + whereClause
	rows, err := s.connection.Query(ctx, doctorsQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to get doctors list: %w", err)
	}
	defer rows.Close()

	var doctorIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan doctor ID: %w", err)
		}
		doctorIDs = append(doctorIDs, id)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating doctors list: %w", err)
	}

	// Для каждого врача рассчитываем и обновляем рейтинг
	for _, id := range doctorIDs {
		if err := s.calculateAndUpdateRatingForDoctor(id); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to update rating for doctor %d: %v", id, err))
			// Продолжаем для остальных врачей, даже если один не удался
			continue
		}
	}

	return nil
}

func (s *Storage) calculateAndUpdateRatingForDoctor(doctorID int) error {
	s.logger.Debug("Calculating rating for doctor", doctorID)
	query := `
	SELECT avg(rating) FROM appointments
	WHERE doctor_id = $1 AND rating IS NOT NULL ;
`
	var avgRatingRaw sql.NullFloat64

	err := s.connection.QueryRow(context.Background(), query, doctorID).Scan(&avgRatingRaw)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error query in database for calculate rating for doctorID=%v", doctorID))
		return fmt.Errorf("failed to calculate average rating: %w", err)
	}
	if !avgRatingRaw.Valid {
		s.logger.Debug(fmt.Sprintf("No valid ratings for doctorID=%v", doctorID))
		return fmt.Errorf("failed to calculate average rating: no valid ratings for doctorID=%v", doctorID)
	}

	s.logger.Info(fmt.Sprintf("Calculated rating for doctorID=%v: rating=%v", doctorID, avgRatingRaw.Float64))

	query = `
	UPDATE doctors
	SET rating = $1
	WHERE id = $2
`
	_, err = s.connection.Exec(context.Background(), query, avgRatingRaw.Float64, doctorID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error updating rating for doctorID=%v", doctorID))
		return fmt.Errorf("failed to update doctor rating: %w", err)
	}
	s.logger.Info(fmt.Sprintf("Updated rating for doctorID=%v", doctorID))
	return nil
}

func (s *Storage) NewDoctor(newDoctor domain.Doctor) (domain.Doctor, error) {
	subQuery := `
	select id from specialization
	where specialization_doctor=$1
`
	var specializationID int
	err := s.connection.QueryRow(context.Background(), subQuery, newDoctor.Specialization).Scan(
		&specializationID,
	)
	s.logger.Info("specializationID", "specializationID", specializationID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Not found specialization=%v", newDoctor.Specialization))
		return domain.Doctor{}, err
	}

	query := `
	INSERT INTO doctors (surname, name, patronymic, specialization_id, education, progress)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id
`
	var doctorID int64
	err = s.connection.QueryRow(context.Background(), query,
		newDoctor.Surname,
		newDoctor.Name,
		newDoctor.Patronymic,
		specializationID,
		newDoctor.Education,
		newDoctor.Progress).Scan(&doctorID)
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
	_, err := s.connection.Exec(context.Background(), query, doctorID)
	if err != nil {
		s.logger.Error("Failed to deleted doctor", "doctorID", doctorID, "error", err)
		return false, err
	}
	isDeleted = true

	return isDeleted, nil

}

func (s *Storage) GetDoctorById(doctorID int) (domain.Doctor, error) {
	err := s.CalculateRating(&doctorID, nil)
	if err != nil {
		s.logger.Error(err.Error())
	}

	query := `
	SELECT  d.id, 
	        surname, 
	        name, 
	        patronymic, 
	        s.specialization_doctor, 
	        education, 
	        progress, 
	        rating
	    FROM doctors d
	    join specialization s on s.id = d.specialization_id
	    WHERE d.id = $1
`
	var doctor domain.Doctor
	var surname, name, patronymic sql.NullString
	err = s.connection.QueryRow(context.Background(), query, doctorID).Scan(
		&doctor.Id,
		&surname,
		&name,
		&patronymic,
		&doctor.Specialization,
		&doctor.Education,
		&doctor.Progress,
		&doctor.Rating,
	)

	if err != nil {
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

	return doctor, nil
}
