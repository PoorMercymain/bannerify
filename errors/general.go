package errors

import "errors"

var (
	ErrSomethingWentWrong = errors.New("something went wrong on server side, please, try again later")
	ErrNoRowsAffected     = errors.New("something went wrong while using database: no rows were affected by insert/update")
)
