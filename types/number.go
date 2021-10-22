package types

import (
	"fmt"
	"strconv"

	"github.com/harness/ff-golang-server-sdk/log"
)

// Number type for clause attribute evaluation
type Number float64

// NewNumber creates a Number instance with the object value
func NewNumber(value interface{}) (*Number, error) {
	num, ok := value.(float64)
	if ok {
		newStr := Number(num)
		return &newStr, nil
	}
	return nil, fmt.Errorf("%v: cant cast to a number", ErrWrongTypeAssertion)
}

// numberOperator takes the first element from the slice, converts it to a float64 and passes to fn for processing.
// we ignore any additional elements if they exist.
func numberOperator(value []string, fn func(float64) bool) bool {
	if len(value) > 0 {
		if i, err := strconv.ParseFloat(value[0], 64); err == nil {
			if fn(i) {
				return true
			}
		}
		log.Warnf("input contains invalid value for number comparisons: %s\n", value)
	}
	return false
}

// StartsWith always return false
func (n Number) StartsWith(value []string) bool {
	return false
}

// EndsWith always return false
func (n Number) EndsWith(value []string) bool {
	return false
}

// Match always return false
func (n Number) Match(value []string) bool {
	return false
}

// Contains always return false
func (n Number) Contains(value []string) bool {
	return false
}

// EqualSensitive always return false
func (n Number) EqualSensitive(value []string) bool {
	return false
}

// Equal check if the number and value are equal
func (n Number) Equal(value []string) bool {
	return numberOperator(value, func(f float64) bool {
		return float64(n) == f
	})
}

// GreaterThan checks if the number is greater than the value
func (n Number) GreaterThan(value []string) bool {
	return numberOperator(value, func(f float64) bool {
		return float64(n) > f
	})
}

// GreaterThanEqual checks if the number is greater or equal than the value
func (n Number) GreaterThanEqual(value []string) bool {
	return numberOperator(value, func(f float64) bool {
		return float64(n) >= f
	})
}

// LessThan checks if the number is less than the value
func (n Number) LessThan(value []string) bool {
	return numberOperator(value, func(f float64) bool {
		return float64(n) < f
	})
}

// LessThanEqual checks if the number is less or equal than the value
func (n Number) LessThanEqual(value []string) bool {
	return numberOperator(value, func(f float64) bool {
		return float64(n) <= f
	})
}

// In checks if the number exist in slice of numbers (value)
func (n Number) In(value []string) bool {
	for _, x := range value {
		if n.Equal([]string{x}) {
			return true
		}
	}
	return false
}
