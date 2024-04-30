package user

import "errors"

var (
	ErrUniqueEmailViolation = errors.New("unique email constraint violation")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidPassword      = errors.New("invalid password")
)
