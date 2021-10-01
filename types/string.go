package types

import (
	"fmt"
	"regexp"
	"strings"
)

// String type for clause attribute evaluation
type String string

// NewString creates a string with the object value
func NewString(value interface{}) (*String, error) {
	str, ok := value.(string)
	if ok {
		newStr := String(str)
		return &newStr, nil
	}
	return nil, fmt.Errorf("%v: cant cast to a string", ErrWrongTypeAssertion)
}

// String implement Stringer interface
func (s String) String() string {
	return string(s)
}

// stringOperator takes the first element from the slice and passes to fn for processing.
// we ignore any additional elements if they exist.
func stringOperator(value []string, fn func(string) bool) bool {
	if len(value) > 0 {
		return fn(value[0])
	}
	return false
}

// StartsWith check if the string starts with the value
func (s String) StartsWith(value []string) bool {
	return stringOperator(value, func(c string) bool {
		return strings.HasPrefix(string(s), c)
	})
}

// EndsWith check if the string ends with the value
func (s String) EndsWith(value []string) bool {
	return stringOperator(value, func(c string) bool {
		return strings.HasSuffix(string(s), c)
	})
}

// Match check if the string match the regex value
func (s String) Match(value []string) bool {
	return stringOperator(value, func(c string) bool {
		if matched, err := regexp.MatchString(string(s), c); err == nil {
			return matched
		}
		return false
	})
}

// Contains check if the string contains the value
func (s String) Contains(value []string) bool {
	return stringOperator(value, func(c string) bool {
		return strings.Contains(string(s), c)
	})
}

// EqualSensitive check if the string and value are equal (case sensitive)
func (s String) EqualSensitive(value []string) bool {
	return stringOperator(value, func(c string) bool {
		return string(s) == c
	})
}

// Equal check if the string and value are equal
func (s String) Equal(value []string) bool {
	return stringOperator(value, func(c string) bool {
		return strings.EqualFold(string(s), c)
	})
}

// GreaterThan checks if the string is greater than the value
func (s String) GreaterThan(value []string) bool {
	return stringOperator(value, func(c string) bool {
		return strings.ToLower(string(s)) > strings.ToLower(c)
	})
}

// GreaterThanEqual checks if the string is greater or equal than the value
func (s String) GreaterThanEqual(value []string) bool {
	return stringOperator(value, func(c string) bool {
		return strings.ToLower(string(s)) >= strings.ToLower(c)
	})
}

// LessThan checks if the string is less than the value
func (s String) LessThan(value []string) bool {
	return stringOperator(value, func(c string) bool {
		return strings.ToLower(string(s)) < strings.ToLower(c)
	})
}

// LessThanEqual checks if the string is less or equal than the value
func (s String) LessThanEqual(value []string) bool {
	return stringOperator(value, func(c string) bool {
		return strings.ToLower(string(s)) <= strings.ToLower(c)
	})
}

// In checks if the string exist in slice of strings (value)
func (s String) In(value []string) bool {
	for _, x := range value {
		if strings.EqualFold(string(s), x) {
			return true
		}
	}
	return false
}
