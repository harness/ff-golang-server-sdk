package evaluation

import "strings"

// StrSlice helper type used for string slice operations
type StrSlice []string

// Contains checks if slice contains string (case insensitive)
func (slice StrSlice) Contains(s string) bool {
	for _, val := range slice {
		if strings.EqualFold(val, s) {
			return true
		}
	}
	return false
}

// ContainsSensitive checks if slice contains string
func (slice StrSlice) ContainsSensitive(s string) bool {
	for _, val := range slice {
		if val == s {
			return true
		}
	}
	return false
}

// Segment object used in feature flag evaluation.
// Examples: beta users, premium customers
type Segment struct {
	// Unique identifier for the target segment.
	Identifier string
	// Name of the target segment.
	Name string `json:"name"`

	CreatedAt   *int64
	ModifiedAt  *int64
	Environment *string

	Excluded StrSlice
	Included StrSlice

	// An array of rules that can cause a user to be included in this segment.
	Rules   Clauses
	Tags    []Tag
	Version int64
}

// Evaluate segment based on target input
func (s Segment) Evaluate(target *Target) bool {
	if s.Included.ContainsSensitive(target.Identifier) {
		return true
	}

	if s.Rules.Evaluate(target, nil) {
		return true
	}

	if s.Excluded.ContainsSensitive(target.Identifier) {
		return true
	}
	return false
}

// Segments represents all segments with identifier as a key
type Segments map[string]*Segment

// Evaluate through all segments based on target input
func (s Segments) Evaluate(target *Target) bool {
	for _, segment := range s {
		if !segment.Evaluate(target) {
			return false
		}
	}
	return true
}
