package client

import "errors"

var (
	// ErrUnauthorized displays error message for unauthorized users
	ErrUnauthorized = errors.New("unauthorized")
)
