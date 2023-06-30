package evaluation

import (
	"strings"

	"github.com/harness/ff-golang-server-sdk/log"
)

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

// SegmentRules is a set of clauses to determine if a target should be included in the segment.
type SegmentRules Clauses

// Evaluate SegmentRules.  This determines if a segment rule is being used to include a target with
// the segment.  SegmentRules are similar to ServingRules except a ServingRule can contain multiple clauses
// but a Segment rule only contains one clause.
func (c SegmentRules) Evaluate(target *Target, segments Segments) bool {
	// OR operation
	for _, clause := range c {
		// operator should be evaluated based on type of attribute
		op, err := target.GetOperator(clause.Attribute)
		if err != nil {
			log.Warn(err)
			continue
		}
		if clause.Evaluate(target, segments, op) {
			return true
		}

		// continue on next rule
	}
	// it means that there was no matching rule
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
	Rules   SegmentRules
	Tags    []Tag
	Version int64
}

// Evaluate segment based on target input
func (s Segment) Evaluate(target *Target) bool {

	// is target excluded from segment via the exclude list
	if s.Excluded.ContainsSensitive(target.Identifier) {
		log.Debugf("target %s excluded from segment %s via exclude list\n", target.Identifier, s.Identifier)
		return false
	}

	// is target included from segment via the include list
	if s.Included.ContainsSensitive(target.Identifier) {
		log.Debugf("target %s included in segment %s via include list\n", target.Identifier, s.Identifier)
		return true
	}

	// is target included in the segment via the clauses
	if s.Rules.Evaluate(target, nil) {
		log.Debugf("target %s included in segment %s via rules\n", target.Identifier, s.Identifier)
		return true
	}

	log.Debugf("No rules to include target %s in segment %s\n", target.Identifier, s.Identifier)
	return false
}

// Segments represents all segments with identifier as a key
type Segments map[string]*Segment

// Evaluate through all segments based on target input
func (s Segments) Evaluate(target *Target) bool {
	for _, segment := range s {
		if segment.Evaluate(target) {
			return true
		}
	}
	return false
}
