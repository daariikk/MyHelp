package domain

type User struct {
	Id         int    `json:"patientID"`
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	Patronymic string `json:"patronymic"`
	Polic      string `json:"polic"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	IsDeleted  bool   `json:"is_deleted"`
}
