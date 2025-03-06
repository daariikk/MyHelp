package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func (s *Storage) DeletePatientById(patientID int) (bool, error) {
	var isDeleted bool

	query := `
	UPDATE patients
	SET is_deleted=true
	WHERE id=$1
`
	_, err := s.connection.Exec(context.Background(), query, patientID)
	if err != nil {
		s.logger.Error("Failed to deleted account", "patientId", patientID, "error", err)
		return false, err
	}

	query = `
	SELECT is_deleted
	FROM patients
	WHERE id=$1
`

	err = s.connection.QueryRow(context.Background(), query, patientID).Scan(
		&isDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Debug("patient not found in database")
			return isDeleted, nil
		}
		return isDeleted, errors.Wrapf(err, "failed to get patient with id %d", patientID)
	}

	return isDeleted, nil
}
