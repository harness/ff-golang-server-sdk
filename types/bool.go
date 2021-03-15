package types

import (
	"fmt"
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

// StartsWith always return false
func (b Boolean) StartsWith(value interface{}) bool {
	return false
}

// EndsWith always return false
func (b Boolean) EndsWith(value interface{}) bool {
	return false
}

// Match always return false
func (b Boolean) Match(value interface{}) bool {
	return false
}

// Contains always return false
func (b Boolean) Contains(value interface{}) bool {
	return false
}

// EqualSensitive always return false
func (b Boolean) EqualSensitive(value interface{}) bool {
	return false
}

// Equal check if the boolean and value are equal
func (b Boolean) Equal(value interface{}) bool {
	val, ok := value.(bool)
	if ok {
		return Boolean(val) == b
	}
	return false
}

// GreaterThan always return false
func (b Boolean) GreaterThan(value interface{}) bool {
	return false
}

// GreaterThanEqual always return false
func (b Boolean) GreaterThanEqual(value interface{}) bool {
	return false
}

// LessThan always return false
func (b Boolean) LessThan(value interface{}) bool {
	return false
}

// LessThanEqual always return false
func (b Boolean) LessThanEqual(value interface{}) bool {
	return false
}

// In always return false
func (b Boolean) In(value interface{}) bool {
	return false
}
