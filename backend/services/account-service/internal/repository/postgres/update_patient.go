package postgres

import (
	"context"
	"github.com/daariikk/MyHelp/services/account-service/internal/domain"
)

func (s *Storage) UpdatePatientById(patient domain.Patient) (domain.Patient, error) {

	var updatedPatient domain.Patient

	query := `
	UPDATE patients
	SET surname=$1, 
	    name=$2, 
	    patronymic=$3, 
	    email=$4, 
	    polic=$5
	WHERE id=$6
`
	_, err := s.connection.Exec(context.Background(), query,
		patient.Surname,
		patient.Name,
		patient.Patronymic,
		patient.Email,
		patient.Polic,
		patient.Id,
	)
	if err != nil {
		s.logger.Error("Failed to updated account", "patientId", patient.Id, "error", err)
		return domain.Patient{}, err
	}

	updatedPatient, err = s.GetPatientById(patient.Id)

	return updatedPatient, nil
}
