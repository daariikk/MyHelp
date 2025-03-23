package domain

type Specialization struct {
	ID                   int    `json:"specializationID"`
	Specialization       string `json:"specialization"`
	SpecializationDoctor string `json:"specialization_doctor"`
	Description          string `json:"description"`
}
