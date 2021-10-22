package types

import (
	"fmt"
	"strconv"

	"github.com/harness/ff-golang-server-sdk/log"
)

// Integer type for clause attribute evaluation
type Integer int64

// NewInteger creates a Integer instance with the object value
func NewInteger(value interface{}) (*Integer, error) {
	num, ok := value.(int64)
	if ok {
		newStr := Integer(num)
		return &newStr, nil
	}
	return nil, fmt.Errorf("%v: cant cast to a integer", ErrWrongTypeAssertion)
}

// intOperator takes the first element from the slice, converts it to a int64 and passes to fn for processing.
// we ignore any additional elements if they exist.
func intOperator(value []string, fn func(int64) bool) bool {
	if len(value) > 0 {
		if i, err := strconv.ParseInt(value[0], 10, 64); err == nil {
			if fn(i) {
				return true
			}
		}
		log.Warnf("input contains invalid value for integer comparisons: %s\n", value)
	}

	return false
}

// StartsWith always return false
func (n Integer) StartsWith([]string) bool {
	return false
}

// EndsWith always return false
func (n Integer) EndsWith([]string) bool {
	return false
}

// Match always return false
func (n Integer) Match([]string) bool {
	return false
}

// Contains always return false
func (n Integer) Contains([]string) bool {
	return false
}

// EqualSensitive always return false
func (n Integer) EqualSensitive([]string) bool {
	return false
}

// Equal check if the number and value are equal
func (n Integer) Equal(value []string) bool {
	return intOperator(value, func(f int64) bool {
		return int64(n) == f
	})
}

// GreaterThan checks if the number is greater than the value
func (n Integer) GreaterThan(value []string) bool {
	return intOperator(value, func(f int64) bool {
		return int64(n) > f
	})
}

// GreaterThanEqual checks if the number is greater or equal than the value
func (n Integer) GreaterThanEqual(value []string) bool {
	return intOperator(value, func(f int64) bool {
		return int64(n) >= f
	})
}

// LessThan checks if the number is less than the value
func (n Integer) LessThan(value []string) bool {
	return intOperator(value, func(f int64) bool {
		return int64(n) < f
	})
}

// LessThanEqual checks if the number is less or equal than the value
func (n Integer) LessThanEqual(value []string) bool {
	return intOperator(value, func(f int64) bool {
		return int64(n) <= f
	})
}

// In checks if the number exist in slice of numbers (value)
func (n Integer) In(value []string) bool {
	for _, x := range value {
		if n.Equal([]string{x}) {
			return true
		}
	}
	return false
}
