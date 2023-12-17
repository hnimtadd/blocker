package core

import "errors"

var (
	ErrExisted     = errors.New("document existed")
	ErrNotExisted  = errors.New("document not existed")
	ErrTypeInvalid = errors.New("document type is invalid")
)
