package postgres

import (
	"context"
	"fmt"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/domain"
	"github.com/pkg/errors"
	"time"
)

func (s *Storage) CreateNewScheduleForDoctor(doctorID int, records []domain.Record) error {
	query := `
		INSERT INTO doctor_schedules (doctor_id, date, start_time, end_time, is_available)
		VALUES ($1, $2, $3, $4, $5)
	`
	ctx := context.Background()

	tx, err := s.connection.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, record := range records {
		_, err := s.connection.Exec(ctx, query,
			doctorID,           // doctor_id
			record.Date,        // date
			record.Start,       // start_time
			record.End,         // end_time
			record.IsAvailable, // is_available
		)
		if err != nil {
			// Откатываем транзакцию в случае ошибки
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to insert record: %w", err)
		}
	}
	// Фиксируем транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Storage) GetScheduleForDoctor(doctorID int, date time.Time) ([]domain.Record, error) {
	query := `
	SELECT  id,
	        doctor_id,
	        date,
	        start_time, 
	   		end_time, 
	    	is_available
	FROM doctor_schedules 
	WHERE doctor_id = $1 AND date >= $2
`
	rows, err := s.connection.Query(context.Background(), query, doctorID, date)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error execution sql query: %v", err))
		return nil, errors.Wrapf(err, "Error executing sql query: %v", query)
	}

	var records []domain.Record

	for rows.Next() {
		var record domain.Record
		err = rows.Scan(
			&record.ID,
			&record.DoctorId,
			&record.Date,
			&record.Start,
			&record.End,
			&record.IsAvailable,
		)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Error execution sql query: %v", err))
			return nil, errors.Wrapf(err, "Error executing sql query: %v", query)
		}

		records = append(records, record)
	}

	return records, nil
}
