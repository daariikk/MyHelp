package repository

import "errors"

var (
	ErrorAlreadyExists = errors.New("patient with this data already exists")
	ErrorNotFound      = errors.New("patient is not found")
)
