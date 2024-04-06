package errors

import "errors"

var (
	ErrSomethingWentWrong = errors.New("something went wrong on server side, please, try again later")
)
