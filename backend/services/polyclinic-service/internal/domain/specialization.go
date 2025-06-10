package domain

type Specialization struct {
	ID                   int    `json:"specializationID"`
	Specialization       string `json:"specialization"`
	SpecializationDoctor string `json:"specialization_doctor"`
	Description          string `json:"description"`
}

func NewSpecialization(id int, specialization, specializationDoctor, description string) Specialization {
	return Specialization{
		ID:                   id,
		Specialization:       specialization,
		SpecializationDoctor: specializationDoctor,
		Description:          description,
	}
}
