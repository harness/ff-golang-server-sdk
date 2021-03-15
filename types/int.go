package types

import (
	"fmt"
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

func (n Integer) operator(value interface{}, fn func(int64) bool) bool {
	num, ok := value.([]int64)
	if ok {
		return fn(num[0])
	}
	return false
}

// StartsWith always return false
func (n Integer) StartsWith(interface{}) bool {
	return false
}

// EndsWith always return false
func (n Integer) EndsWith(interface{}) bool {
	return false
}

// Match always return false
func (n Integer) Match(interface{}) bool {
	return false
}

// Contains always return false
func (n Integer) Contains(interface{}) bool {
	return false
}

// EqualSensitive always return false
func (n Integer) EqualSensitive(interface{}) bool {
	return false
}

// Equal check if the number and value are equal
func (n Integer) Equal(value interface{}) bool {
	return n.operator(value, func(f int64) bool {
		return int64(n) == f
	})
}

// GreaterThan checks if the number is greater than the value
func (n Integer) GreaterThan(value interface{}) bool {
	return n.operator(value, func(f int64) bool {
		return int64(n) > f
	})
}

// GreaterThanEqual checks if the number is greater or equal than the value
func (n Integer) GreaterThanEqual(value interface{}) bool {
	return n.operator(value, func(f int64) bool {
		return int64(n) >= f
	})
}

// LessThan checks if the number is less than the value
func (n Integer) LessThan(value interface{}) bool {
	return n.operator(value, func(f int64) bool {
		return int64(n) < f
	})
}

// LessThanEqual checks if the number is less or equal than the value
func (n Integer) LessThanEqual(value interface{}) bool {
	return n.operator(value, func(f int64) bool {
		return int64(n) <= f
	})
}

// In checks if the number exist in slice of numbers (value)
func (n Integer) In(value interface{}) bool {
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
