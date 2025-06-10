package domain

type Patient struct {
	Id         int    `json:"patientID"`
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	Patronymic string `json:"patronymic"`
	Polic      string `json:"polic"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	IsDeleted  bool   `json:"is_deleted"`
}

type PatientDTO struct {
	Id           int              `json:"patientID"`
	Surname      string           `json:"surname"`
	Name         string           `json:"name"`
	Patronymic   string           `json:"patronymic"`
	Polic        string           `json:"polic"`
	Email        string           `json:"email"`
	Password     string           `json:"password"`
	IsDeleted    bool             `json:"is_deleted"`
	Appointments []AppointmentDTO `json:"appointments"`
}

func NewPatient(id int, surname, name, patronymic, polic, email, password string, isDeleted bool) *Patient {
	return &Patient{
		Id:         id,
		Surname:    surname,
		Name:       name,
		Patronymic: patronymic,
		Polic:      polic,
		Email:      email,
		Password:   password,
		IsDeleted:  isDeleted,
	}
}

func NewPatientDTO(
	id int, surname, name, patronymic, polic, email, password string, isDeleted bool, appointments []AppointmentDTO,
) *PatientDTO {
	return &PatientDTO{
		Id:           id,
		Surname:      surname,
		Name:         name,
		Patronymic:   patronymic,
		Polic:        polic,
		Email:        email,
		Password:     password,
		IsDeleted:    isDeleted,
		Appointments: appointments,
	}
}
