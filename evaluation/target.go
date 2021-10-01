package evaluation

import (
	"fmt"
	"strings"

	"github.com/drone/ff-golang-server-sdk/types"

	"reflect"
)

// Target object
type Target struct {
	Identifier string
	Name       string
	Anonymous  *bool
	Attributes *map[string]interface{}
}

// GetAttrValue returns value from target with specified attribute
func (t Target) GetAttrValue(attr string) reflect.Value {
	var value reflect.Value
	attrs := make(map[string]interface{})
	if t.Attributes != nil {
		attrs = *t.Attributes
	}

	attrVal, ok := attrs[attr] // first check custom attributes
	if ok {
		value = reflect.ValueOf(attrVal)
	} else {
		// We only have two fields here, so we will access the fields directly, and use reflection if we start adding
		// more in the future
		switch strings.ToLower(attr) {
		case "identifier":
			value = reflect.ValueOf(t.Identifier)
		case "name":
			value = reflect.ValueOf(t.Name)
		}
	}
	return value
}

// GetOperator returns interface based on attribute value
func (t Target) GetOperator(attr string) (types.ValueType, error) {
	if attr == "" {
		return nil, nil
	}
	value := t.GetAttrValue(attr)
	switch value.Kind() {
	case reflect.Bool:
		return types.Boolean(value.Bool()), nil
	case reflect.String:
		return types.String(value.String()), nil
	case reflect.Float64, reflect.Float32:
		return types.Number(value.Float()), nil
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Uint, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uint8:
		return types.Integer(value.Int()), nil
	case reflect.Slice:
		return types.NewSlice(value.Interface()), nil
	case reflect.Array, reflect.Chan, reflect.Complex128, reflect.Complex64, reflect.Func, reflect.Interface,
		reflect.Invalid, reflect.Map, reflect.Ptr, reflect.Struct, reflect.Uintptr, reflect.UnsafePointer:
		fallthrough
	default:
		return nil, fmt.Errorf("unexpected type: %s", value.Kind().String())
	}
}
