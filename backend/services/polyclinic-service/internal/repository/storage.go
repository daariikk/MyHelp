package repository

import "errors"

var (
	ErrorAlreadyExists          = errors.New("patient with this data already exists")
	ErrorDoctorNotFound         = errors.New("doctor is not found")
	ErrorSpecializationNotFound = errors.New("specialization is not found")
)
