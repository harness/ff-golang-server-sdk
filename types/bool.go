package types

import (
	"fmt"
	"strconv"

	"github.com/harness/ff-golang-server-sdk/log"
)

// Boolean type for clause attribute evaluation
type Boolean bool

// NewBoolean creates a Boolean instance with the object value
func NewBoolean(value interface{}) (*Boolean, error) {
	num, ok := value.(bool)
	if ok {
		newBool := Boolean(num)
		return &newBool, nil
	}
	return nil, fmt.Errorf("%v: cant cast to a booelan", ErrWrongTypeAssertion)
}

// boolOperator iterates over value, converts it to a float64 and calls fn on each element
func boolOperator(value []string, fn func(bool) bool) bool {
	if len(value) > 0 {
		if i, err := strconv.ParseBool(value[0]); err == nil {
			return fn(i)
		}
		log.Warnf("input contains invalid value for bool comparisons: %s\n", value)
	}
	return false
}

// StartsWith always return false
func (b Boolean) StartsWith(value []string) bool {
	return false
}

// EndsWith always return false
func (b Boolean) EndsWith(value []string) bool {
	return false
}

// Match always return false
func (b Boolean) Match(value []string) bool {
	return false
}

// Contains always return false
func (b Boolean) Contains(value []string) bool {
	return false
}

// EqualSensitive always return false
func (b Boolean) EqualSensitive(value []string) bool {
	return false
}

// Equal check if the boolean and value are equal
func (b Boolean) Equal(value []string) bool {
	return boolOperator(value, func(f bool) bool {
		return bool(b) == f
	})
}

// GreaterThan always return false
func (b Boolean) GreaterThan(value []string) bool {
	return false
}

// GreaterThanEqual always return false
func (b Boolean) GreaterThanEqual(value []string) bool {
	return false
}

// LessThan always return false
func (b Boolean) LessThan(value []string) bool {
	return false
}

// LessThanEqual always return false
func (b Boolean) LessThanEqual(value []string) bool {
	return false
}

// In always return false
func (b Boolean) In(value []string) bool {
	return false
}
