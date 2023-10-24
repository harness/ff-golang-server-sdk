package client

import (
	"errors"
	"fmt"
)

var (
	EmptySDKKeyError              = errors.New("default variation was returned")
	DefaultVariationReturnedError = errors.New("default variation was returned")
	FetchFlagsError               = errors.New("fetching flags failed")
)

type NonRetryableAuthError struct {
	StatusCode string
	Message    string
}

func (e NonRetryableAuthError) Error() string {
	return fmt.Sprintf("unauthorized: %s: %s", e.StatusCode, e.Message)
}

type RetryableAuthError struct {
	StatusCode string
	Message    string
}

func (e RetryableAuthError) Error() string {
	return fmt.Sprintf("server error: %s: %s", e.StatusCode, e.Message)
}

type InitializeTimeoutError struct {
}

func (e InitializeTimeoutError) Error() string {
	return fmt.Sprintf("timeout waiting to initialize")
}
