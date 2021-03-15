package types

import (
	"fmt"
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

func (n Number) operator(value interface{}, fn func(float64) bool) bool {
	num, ok := value.([]float64)
	if ok {
		return fn(num[0])
	}
	return false
}

// StartsWith always return false
func (n Number) StartsWith(value interface{}) bool {
	return false
}

// EndsWith always return false
func (n Number) EndsWith(value interface{}) bool {
	return false
}

// Match always return false
func (n Number) Match(value interface{}) bool {
	return false
}

// Contains always return false
func (n Number) Contains(value interface{}) bool {
	return false
}

// EqualSensitive always return false
func (n Number) EqualSensitive(value interface{}) bool {
	return false
}

// Equal check if the number and value are equal
func (n Number) Equal(value interface{}) bool {
	return n.operator(value, func(f float64) bool {
		return float64(n) == f
	})
}

// GreaterThan checks if the number is greater than the value
func (n Number) GreaterThan(value interface{}) bool {
	return n.operator(value, func(f float64) bool {
		return float64(n) > f
	})
}

// GreaterThanEqual checks if the number is greater or equal than the value
func (n Number) GreaterThanEqual(value interface{}) bool {
	return n.operator(value, func(f float64) bool {
		return float64(n) >= f
	})
}

// LessThan checks if the number is less than the value
func (n Number) LessThan(value interface{}) bool {
	return n.operator(value, func(f float64) bool {
		return float64(n) < f
	})
}

// LessThanEqual checks if the number is less or equal than the value
func (n Number) LessThanEqual(value interface{}) bool {
	return n.operator(value, func(f float64) bool {
		return float64(n) <= f
	})
}

// In checks if the number exist in slice of numbers (value)
func (n Number) In(value interface{}) bool {
	array, ok := value.([]interface{})
	if ok {
		for _, val := range array {
			if n.Equal(val) {
				return true
			}
		}
	}
	return false
}
