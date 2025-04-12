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
	isExistPatient, err := s.CheckUserByEmail(user.Email)
	if isExistPatient {
		return domain.User{}, errors.New("User already exists")
	}

	query := `
		INSERT INTO patients (name, polic, email, password)
        VALUES ($1, $2, $3, $4)
        RETURNING id
`
	var patientId int
	err = s.connection.QueryRow(context.Background(), query,
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
	SELECT id, password FROM patients WHERE email=$1 and is_deleted=false
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

func (s *Storage) GetAdminPassword(email string) (int, string, error) {
	query := `
	SELECT id, password FROM admins WHERE email=$1 and is_active=true
`
	rows, err := s.connection.Query(context.Background(), query, email)
	if err != nil {
		return 0, "", errors.Wrap(err, "failed to query database")
	}
	defer rows.Close()

	user := domain.Admin{}

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

func (s *Storage) CheckUserByEmail(email string) (bool, error) {
	query := `
	select email
	from patients
	where email=$1 and is_deleted=false
`
	rows, err := s.connection.Query(context.Background(), query, email)
	if err != nil {
		s.logger.Error("Failed to query database", "email", email, "error", err)
		return false, errors.Wrap(err, "failed to query database: attempt to check user by email")
	}
	defer rows.Close()

	if rows.Next() {
		s.logger.Debug("User found", "email", email)
		return true, nil
	}

	s.logger.Debug("User not found", "email", email)
	return false, nil
}

func (s *Storage) GetUser(email string) (domain.User, error) {
	var user domain.User

	query := `
	select email, password
	from patients
	where email=$1 and is_deleted=false
`
	err := s.connection.QueryRow(context.Background(), query, email).Scan(&user.Email, &user.Password)
	if err == pgx.ErrNoRows {
		return domain.User{}, errors.New("user not found")
	}
	if err != nil {
		s.logger.Error("Failed to query database", "email", email, "error", err)
		return domain.User{}, errors.Wrap(err, "failed to query database: attempt to get user")
	}

	return user, nil
}

func (s *Storage) GetAdmin(email string) (domain.Admin, error) {
	var user domain.Admin

	query := `
	select email, password
	from admins
	where email=$1 and is_active=true
`
	err := s.connection.QueryRow(context.Background(), query, email).Scan(&user.Email, &user.Password)
	if err == pgx.ErrNoRows {
		return domain.Admin{}, errors.New("admin not found")
	}
	if err != nil {
		s.logger.Error("Failed to query database", "email", email, "error", err)
		return domain.Admin{}, errors.Wrap(err, "failed to query database: attempt to get admin")
	}

	return user, nil
}
