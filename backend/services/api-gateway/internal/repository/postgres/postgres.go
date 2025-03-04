package postgres

import (
	"context"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/domain"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/repository"
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

func (s *Storage) RegisterUser(user domain.User) (domain.User, error) {
	query := `
		INSERT INTO patients (name, polic, email, password)
        VALUES ($1, $2, $3, $4)
        RETURNING id
`
	var patientId int
	err := s.connection.QueryRow(context.Background(), query,
		user.Name,
		user.Polic,
		user.Email,
		user.Password,
	).Scan(&patientId)
	if err != nil {
		return domain.User{}, errors.Wrap(err, "failed to register user")
	}
	user.Id = patientId
	return user, nil
}

func (s *Storage) GetPassword(email string) (int, string, error) {
	query := `
	SELECT id, password FROM patients WHERE email=$1
`
	rows, err := s.connection.Query(context.Background(), query, email)
	if err != nil {
		return 0, "", errors.Wrap(err, "failed to query database")
	}
	defer rows.Close()

	user := domain.User{}

	if rows.Next() {
		err = rows.Scan(&user.Id, &user.Password)
		if err != nil {
			return 0, "", errors.Wrap(err, "failed to scan row")
		}
	} else {
		return 0, "", errors.New("user not found")
	}
	return user.Id, user.Password, nil
}

func (s *Storage) UpdatePassword(email string, password string) error {

	s.logger.Debug("Updating password", "email", email, "password", password)

	query := `
	UPDATE patients
	SET password=$1
	WHERE email=$2
`
	_, err := s.connection.Exec(context.Background(), query, password, email)
	if err != nil {
		s.logger.Error("Failed to update password", "email", email, "error", err)
		return errors.Wrap(err, repository.ErrorNotFound.Error())
	}

	s.logger.Debug("Successfully updated password", "email", email, "password", password)
	return nil
}
