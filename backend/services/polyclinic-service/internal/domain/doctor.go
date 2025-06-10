package domain

type Doctor struct {
	Id             int     `json:"doctorID"`
	Surname        string  `json:"surname"`
	Name           string  `json:"name"`
	Patronymic     string  `json:"patronymic"`
	Specialization string  `json:"specialization"`
	Education      string  `json:"education"`
	Progress       string  `json:"progress"`
	Rating         float64 `json:"rating"`
	PhotoPath      string  `json:"photo"`
}

func NewDoctor(
	id int,
	surname string,
	name string,
	patronymic string,
	specialization string,
	education string,
	progress string,
	rating float64,
	photoPath string,
) Doctor {
	return Doctor{
		Id:             id,
		Surname:        surname,
		Name:           name,
		Patronymic:     patronymic,
		Specialization: specialization,
		Education:      education,
		Progress:       progress,
		Rating:         rating,
		PhotoPath:      photoPath,
	}
}
