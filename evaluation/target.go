package evaluation

import (
	"github.com/drone/ff-golang-server-sdk.v1/types"

	"reflect"
)

// Target object
type Target struct {
	Identifier string
	Name       *string
	Anonymous  bool
	Attributes map[string]interface{}
}

// GetAttrValue returns value from target with specified attribute
func (t Target) GetAttrValue(attr string) reflect.Value {
	var value reflect.Value
	attrVal, ok := t.Attributes[attr] // first check custom attributes
	if ok {
		value = reflect.ValueOf(attrVal)
	} else {
		value = GetStructFieldValue(t, attr)
	}
	return value
}

// GetOperator returns interface based on attribute value
func (t Target) GetOperator(attr string) types.ValueType {
	value := t.GetAttrValue(attr)
	switch value.Kind() {
	case reflect.Bool:
		return types.Boolean(value.Bool())
	case reflect.String:
		return types.String(value.String())
	case reflect.Float64, reflect.Float32:
		return types.Number(value.Float())
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Uint, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uint8:
		return types.Integer(value.Int())
	case reflect.Array, reflect.Chan, reflect.Complex128, reflect.Complex64, reflect.Func, reflect.Interface,
		reflect.Invalid, reflect.Map, reflect.Ptr, reflect.Slice, reflect.Struct, reflect.Uintptr, reflect.UnsafePointer:
		return nil
	default:
		return nil
	}
}
