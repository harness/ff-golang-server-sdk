package types

import "reflect"

// Array type for clause attribute evaluation
type Array struct {
	reflect.Kind
	value interface{}
}

// NewArray creates a Array instance with the object value
// TBD
func NewArray(kind reflect.Kind, value interface{}) *Array {
	return &Array{Kind: kind, value: value}
}
