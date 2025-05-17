package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"log/slog"
	"time"
)

type Storage struct {
	//pool       *pgxpool.Pool
	connection *pgx.Conn
	logger     *slog.Logger
}

func New(ctx context.Context, logger *slog.Logger, url string) (*Storage, error) {
	// Парсим URL для получения конфигурации
	config, err := pgx.ParseConfig(url)
	if err != nil {
		logger.Error("Failed to parse postgres connection string", "error", err)
		return nil, errors.Wrap(err, "failed to parse postgres connection string")
	}

	// Устанавливаем соединение с конфигурацией
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		logger.Error("Failed to connect to postgres", "error", err)
		return nil, errors.Wrapf(err, "failed to connect to postgres")
	}

	// Проверяем соединение
	if err := conn.Ping(ctx); err != nil {
		logger.Error("Failed to ping postgres", "error", err)
		return nil, errors.Wrap(err, "failed to ping postgres")
	}

	logger.Info("Successfully connected to postgres")
	return &Storage{connection: conn, logger: logger}, nil
}

func (s *Storage) Close() error {
	if s.connection == nil {
		return nil
	}

	// Создаем контекст с таймаутом для закрытия соединения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.connection.Close(ctx); err != nil {
		s.logger.Error("Failed to close postgres connection", "error", err)
		return errors.Wrap(err, "failed to close postgres connection")
	}

	s.logger.Info("Postgres connection closed successfully")
	return nil
}

// Дополнительный метод для проверки соединения
func (s *Storage) Ping(ctx context.Context) error {
	if s.connection == nil {
		return errors.New("connection is nil")
	}
	return s.connection.Ping(ctx)
}

func (s *Storage) GetAllSpecializations() ([]domain.Specialization, error) {
	query := `
		SELECT id, specialization, specialization_doctor, description
		FROM specialization 
`
	var specializations []domain.Specialization

	rows, err := s.connection.Query(context.Background(), query)
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
	rows, err := s.connection.Query(context.Background(), query, specializationID)
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
