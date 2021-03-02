package types

import (
	"fmt"
	"github.com/wings-software/ff-client-sdk-go"
)

type Integer int64

func NewInteger(value interface{}) (*Integer, error) {
	num, ok := value.(int64)
	if ok {
		newStr := Integer(num)
		return &newStr, nil
	}
	return nil, fmt.Errorf("%v: cant cast to a integer", ff_golang_server_sdk.ErrWrongTypeAssertion)
}

func (n Integer) operator(value interface{}, fn func(int64) bool) bool {
	num, ok := value.([]int64)
	if ok {
		return fn(num[0])
	}
	return false
}

func (n Integer) StartsWith(interface{}) bool {
	return false
}

func (n Integer) EndWith(interface{}) bool {
	return false
}

func (n Integer) Match(interface{}) bool {
	return false
}

func (n Integer) Contains(interface{}) bool {
	return false
}

func (n Integer) EqualSensitive(interface{}) bool {
	return false
}

func (n Integer) Equal(value interface{}) bool {
	return n.operator(value, func(f int64) bool {
		return int64(n) == f
	})
}

func (n Integer) GreaterThan(value interface{}) bool {
	return n.operator(value, func(f int64) bool {
		return int64(n) > f
	})
}

func (n Integer) GreaterThanEqual(value interface{}) bool {
	return n.operator(value, func(f int64) bool {
		return int64(n) >= f
	})
}

func (n Integer) LessThan(value interface{}) bool {
	return n.operator(value, func(f int64) bool {
		return int64(n) < f
	})
}

func (n Integer) LessThanEqual(value interface{}) bool {
	return n.operator(value, func(f int64) bool {
		return int64(n) <= f
	})
}

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
