package models

import (
	"errors"
)

var (

	ErrNoRecord = errors.New("models: no matching record found")

	ErrDuplicateEmail = errors.New("models: duplicate email")

	ErrInvalidCredentials = errors.New("models: invalid credentials")

)