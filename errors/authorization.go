package errors

import "errors"

var (
	ErrTokenIsInvalid    = errors.New("invalid token provided")
	ErrNoTokenProvided   = errors.New("token was not provided")
	ErrAdminRequired     = errors.New("admin role needed to get access to the endpoint")
	ErrAlreadyRegistered = errors.New("user with this login is already registered")
	ErrUserNotFound      = errors.New("user not found")
	ErrWrongPassword     = errors.New("wrong password provided")
	ErrNoLoginOrPassword = errors.New("no login or password provided")
	ErrWrongAdminHeader  = errors.New("wrong admin header")
)
