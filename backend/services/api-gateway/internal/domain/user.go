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

type Admin struct {
	Id       int    `json:"adminID"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IsActive bool   `json:"isActive"`
}

func NewUser(id int, surname, name, patronymic, polic, email, password string, isDeleted bool) *User {
	return &User{
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

func NewAdmin(id int, username, email, password string, isActive bool) *Admin {
	return &Admin{
		Id:       id,
		Username: username,
		Email:    email,
		Password: password,
		IsActive: isActive,
	}
}
