package types

import (
	"fmt"
	"github.com/drone/ff-golang-server-sdk"
)

type Boolean bool

func NewBoolean(value interface{}) (*Boolean, error) {
	num, ok := value.(bool)
	if ok {
		newBool := Boolean(num)
		return &newBool, nil
	}
	return nil, fmt.Errorf("%v: cant cast to a integer", ff_golang_server_sdk.ErrWrongTypeAssertion)
}

func (b Boolean) StartsWith(value interface{}) bool {
	return false
}

func (b Boolean) EndWith(value interface{}) bool {
	return false
}

func (b Boolean) Match(value interface{}) bool {
	return false
}

func (b Boolean) Contains(value interface{}) bool {
	return false
}

func (b Boolean) EqualSensitive(value interface{}) bool {
	return false
}

func (b Boolean) Equal(value interface{}) bool {
	val, ok := value.(bool)
	if ok {
		return Boolean(val) == b
	}
	return false
}

func (b Boolean) GreaterThan(value interface{}) bool {
	return false
}

func (b Boolean) GreaterThanEqual(value interface{}) bool {
	return false
}

func (b Boolean) LessThan(value interface{}) bool {
	return false
}

func (b Boolean) LessThanEqual(value interface{}) bool {
	return false
}

func (b Boolean) In(value interface{}) bool {
	return false
}
