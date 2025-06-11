package domain

import "testing"

func TestNewUser(t *testing.T) {
	user := NewUser(1, "Иванов", "Иван", "Иванович", "1234567890", "test@example.com", "pass", true)
	if user == nil {
		t.Fatal("NewUser вернул nil")
	}
	if user.Id != 1 || user.Surname != "Иванов" || user.Name != "Иван" || user.Patronymic != "Иванович" {
		t.Errorf("ФИО заполнены неверно: %+v", user)
	}
	if user.Polic != "1234567890" {
		t.Errorf("Polic неверен: %s", user.Polic)
	}
	if user.Email != "test@example.com" || user.Password != "pass" {
		t.Errorf("Email или Password неверны: %s %s", user.Email, user.Password)
	}
	if user.IsDeleted != true {
		t.Errorf("IsDeleted неверен: %v", user.IsDeleted)
	}
}

func TestNewUser_EmptyFields(t *testing.T) {
	user := NewUser(0, "", "", "", "", "", "", false)
	if user == nil {
		t.Fatal("NewUser вернул nil")
	}
	if user.Id != 0 || user.Surname != "" || user.Name != "" || user.Patronymic != "" {
		t.Errorf("Поля не пустые: %+v", user)
	}
	if user.Polic != "" || user.Email != "" || user.Password != "" || user.IsDeleted != false {
		t.Errorf("Поля не пустые: %+v", user)
	}
}

func TestNewAdmin(t *testing.T) {
	admin := NewAdmin(2, "admin", "admin@admin.com", "secret", true)
	if admin == nil {
		t.Fatal("NewAdmin вернул nil")
	}
	if admin.Id != 2 {
		t.Errorf("Id неверен: %d", admin.Id)
	}
	if admin.Username != "admin" {
		t.Errorf("Username неверен: %s", admin.Username)
	}
	if admin.Email != "admin@admin.com" {
		t.Errorf("Email неверен: %s", admin.Email)
	}
	if admin.Password != "secret" {
		t.Errorf("Password неверен: %s", admin.Password)
	}
	if admin.IsActive != true {
		t.Errorf("IsActive неверен: %v", admin.IsActive)
	}
}

func TestNewAdmin_EmptyFields(t *testing.T) {
	admin := NewAdmin(0, "", "", "", false)
	if admin == nil {
		t.Fatal("NewAdmin вернул nil")
	}
	if admin.Id != 0 || admin.Username != "" || admin.Email != "" || admin.Password != "" || admin.IsActive != false {
		t.Errorf("Поля не пустые: %+v", admin)
	}
}
