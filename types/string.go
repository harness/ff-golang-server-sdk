package types

import (
	"fmt"
	"github.com/drone/ff-golang-server-sdk"
	"regexp"
	"strings"
)

type String string

func NewString(value interface{}) (*String, error) {
	str, ok := value.(string)
	if ok {
		newStr := String(str)
		return &newStr, nil
	}
	return nil, fmt.Errorf("%v: cant cast to a string", ff_golang_server_sdk.ErrWrongTypeAssertion)
}

func (s String) String() string {
	return string(s)
}

func (s String) operator(value interface{}, fn func(string) bool) bool {
	slice, ok := value.([]string)
	if ok {
		return fn(slice[0]) // need confirmation
	}
	return false
}

func (s String) StartsWith(value interface{}) bool {
	return s.operator(value, func(c string) bool {
		return strings.HasPrefix(string(s), c)
	})
}

func (s String) EndWith(value interface{}) bool {
	return s.operator(value, func(c string) bool {
		return strings.HasSuffix(string(s), c)
	})
}

func (s String) Match(value interface{}) bool {
	return s.operator(value, func(c string) bool {
		if matched, err := regexp.MatchString(string(s), c); err == nil {
			return matched
		}
		return false
	})
}

func (s String) Contains(value interface{}) bool {
	return s.operator(value, func(c string) bool {
		return strings.Contains(string(s), c)
	})
}

func (s String) EqualSensitive(value interface{}) bool {
	return s.operator(value, func(c string) bool {
		return string(s) == c
	})
}

func (s String) Equal(value interface{}) bool {
	return s.operator(value, func(c string) bool {
		return strings.ToLower(string(s)) == strings.ToLower(c)
	})
}

func (s String) GreaterThan(value interface{}) bool {
	return s.operator(value, func(c string) bool {
		return strings.ToLower(string(s)) > strings.ToLower(c)
	})
}

func (s String) GreaterThanEqual(value interface{}) bool {
	return s.operator(value, func(c string) bool {
		return strings.ToLower(string(s)) >= strings.ToLower(c)
	})
}

func (s String) LessThan(value interface{}) bool {
	return s.operator(value, func(c string) bool {
		return strings.ToLower(string(s)) < strings.ToLower(c)
	})
}

func (s String) LessThanEqual(value interface{}) bool {
	return s.operator(value, func(c string) bool {
		return strings.ToLower(string(s)) <= strings.ToLower(c)
	})
}

func (s String) In(value interface{}) bool {
	array, ok := value.([]string)
	if ok {
		for _, val := range array {
			if s.Equal(val) {
				return true
			}
		}
	}
	return false
}
