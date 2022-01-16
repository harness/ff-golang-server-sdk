package repository

import (
	"errors"
)

var (
	// ErrFeatureConfigNotFound ...
	ErrFeatureConfigNotFound = errors.New("feature config not found")
	// ErrSegmentNotFound ...
	ErrSegmentNotFound = errors.New("target group not found")
)
