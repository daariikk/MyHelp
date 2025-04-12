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
