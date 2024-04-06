package errors

import "errors"

var (
	ErrDuplicateInJSON       = errors.New("duplicate key found in JSON")
	ErrWrongMIME             = errors.New("wrong MIME type used")
	ErrWrongJSON             = errors.New("something is wrong in json")
	ErrNothingProvidedInJSON = errors.New("nothing provided in JSON")
)
