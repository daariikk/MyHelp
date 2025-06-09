package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"log/slog"
)

type Storage struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func New(ctx context.Context, logger *slog.Logger, url string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		logger.Error("Failed to create pgxpool", "error", err)
		return nil, errors.Wrap(err, "failed to create pgxpool")
	}

	logger.Info("Successfully connected to postgres")
	return &Storage{pool: pool, logger: logger}, nil
}

func (s *Storage) Close() error {
	s.pool.Close()
	s.logger.Info("Postgres pool closed successfully")
	return nil
}

// Дополнительный метод для проверки соединения
func (s *Storage) Ping(ctx context.Context) error {
	if s.pool == nil {
		return errors.New("connection pool is nil")
	}
	return s.pool.Ping(ctx)
}
