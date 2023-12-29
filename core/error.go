package core

import (
	"errors"
)

var (
	ErrDocExisted    = errors.New("document existed")
	ErrDocNotExisted = errors.New("document not existed")
	ErrTypeInvalid   = errors.New("document type is invalid")
	ErrSigInvalid    = errors.New("invalid signature")
	ErrSigNotExisted = errors.New("signature not found")
	ErrNonceInvalid  = errors.New("nonce is invalid")
)
