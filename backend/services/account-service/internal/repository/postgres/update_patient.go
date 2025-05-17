package postgres

import (
	"context"
	"fmt"
	"github.com/daariikk/MyHelp/services/account-service/internal/domain"
	"github.com/daariikk/MyHelp/services/account-service/internal/repository"
	"github.com/jackc/pgx/v5"
	"strings"
)

func (s *Storage) UpdatePatientById(patient domain.Patient) (domain.Patient, error) {
	// Проверяем email, если он передан
	if patient.Email != "" {
		var exists int
		err := s.connection.QueryRow(context.Background(),
			"SELECT 1 FROM patients WHERE email = $1 AND id != $2",
			patient.Email, patient.Id).Scan(&exists)

		if err == nil {
			return domain.Patient{}, repository.ErrorEmailUnique
		} else if err != pgx.ErrNoRows {
			s.logger.Error("Failed to check email uniqueness", "error", err)
			return domain.Patient{}, fmt.Errorf("failed to check email uniqueness")
		}
	}

	query := "UPDATE patients SET "
	args := []interface{}{}
	paramCount := 1

	if patient.Surname != "" {
		query += fmt.Sprintf("surname=$%d, ", paramCount)
		args = append(args, patient.Surname)
		paramCount++
	}

	if patient.Name != "" {
		query += fmt.Sprintf("name=$%d, ", paramCount)
		args = append(args, patient.Name)
		paramCount++
	}

	if patient.Patronymic != "" {
		query += fmt.Sprintf("patronymic=$%d, ", paramCount)
		args = append(args, patient.Patronymic)
		paramCount++
	}

	if patient.Email != "" {
		query += fmt.Sprintf("email=$%d, ", paramCount)
		args = append(args, patient.Email)
		paramCount++
	}

	if patient.Polic != "" {
		query += fmt.Sprintf("polic=$%d, ", paramCount)
		args = append(args, patient.Polic)
		paramCount++
	}

	query = strings.TrimSuffix(query, ", ")

	query += fmt.Sprintf(" WHERE id=$%d", paramCount)
	args = append(args, patient.Id)

	_, err := s.connection.Exec(context.Background(), query, args...)
	if err != nil {
		s.logger.Error("Failed to update patient", "patientId", patient.Id, "error", err)
		return domain.Patient{}, fmt.Errorf("failed to update patient")
	}

	updatedPatient, err := s.GetPatientById(patient.Id)
	if err != nil {
		s.logger.Error("Failed to get updated patient", "patientId", patient.Id, "error", err)
		return domain.Patient{}, fmt.Errorf("failed to get updated patient")
	}

	return updatedPatient, nil
}
