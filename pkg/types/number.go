package types

import (
	"fmt"
	"github.com/wings-software/ff-client-sdk-go/pkg"
)

type Number float64

func NewNumber(value interface{}) (*Number, error) {
	num, ok := value.(float64)
	if ok {
		newStr := Number(num)
		return &newStr, nil
	}
	return nil, fmt.Errorf("%v: cant cast to a number", pkg.ErrWrongTypeAssertion)
}

func (n Number) operator(value interface{}, fn func(float64) bool) bool {
	num, ok := value.([]float64)
	if ok {
		return fn(num[0])
	}
	return false
}

func (n Number) StartsWith(value interface{}) bool {
	return false
}

func (n Number) EndWith(value interface{}) bool {
	return false
}

func (n Number) Match(value interface{}) bool {
	return false
}

func (n Number) Contains(value interface{}) bool {
	return false
}

func (n Number) EqualSensitive(value interface{}) bool {
	return false
}

func (n Number) Equal(value interface{}) bool {
	return n.operator(value, func(f float64) bool {
		return float64(n) == f
	})
}

func (n Number) GreaterThan(value interface{}) bool {
	return n.operator(value, func(f float64) bool {
		return float64(n) > f
	})
}

func (n Number) GreaterThanEqual(value interface{}) bool {
	return n.operator(value, func(f float64) bool {
		return float64(n) >= f
	})
}

func (n Number) LessThan(value interface{}) bool {
	return n.operator(value, func(f float64) bool {
		return float64(n) < f
	})
}

func (n Number) LessThanEqual(value interface{}) bool {
	return n.operator(value, func(f float64) bool {
		return float64(n) <= f
	})
}

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
