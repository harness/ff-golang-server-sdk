package evaluation

import "strings"

type StrSlice []string

func (slice StrSlice) Contains(identifier string) bool {
	for _, val := range slice {
		if strings.ToLower(val) == strings.ToLower(identifier) {
			return true
		}
	}
	return false
}

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

func (s Segment) Evaluate(target *Target) bool {
	if s.Included.Contains(target.Identifier) {
		return true
	}

	if s.Rules.Evaluate(target, nil) {
		return true
	}

	if s.Excluded.Contains(target.Identifier) {
		return true
	}
	return false
}

type Segments map[string]*Segment

func (s Segments) Evaluate(target *Target) bool {
	for _, segment := range s {
		if !segment.Evaluate(target) {
			return false
		}
	}
	return true
}
