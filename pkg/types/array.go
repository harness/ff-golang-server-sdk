package types

import "reflect"

type Array struct {
	reflect.Kind
	value interface{}
}

func NewArray(kind reflect.Kind, value interface{}) *Array {
	return &Array{Kind: kind, value: value}
}
