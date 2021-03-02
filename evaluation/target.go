package evaluation

import (
	"github.com/drone/ff-golang-server-sdk/types"
	"reflect"
)

type Target struct {
	Identifier string
	Name       *string
	Anonymous  bool
	Attributes map[string]interface{}
}

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

func (t Target) GetOperator(attr string) types.ValueType {
	value := t.GetAttrValue(attr)
	switch value.Kind() {
	case reflect.Bool:
		return types.Boolean(value.Bool())
	case reflect.String:
		return types.String(value.String())
	case reflect.Float64:
		return types.Number(value.Float())
	case reflect.Int:
		return types.Integer(value.Int())
		// TODO: map support, JSON Path is solution, need clarification from the team
		//case reflect.Map:
		//	mapIntf := value.Interface()
		//	mapValue := mapIntf.(map[string]interface{})
		//	_, key := parseMapField()
		//	val := mapValue[key]
		//	parseAttr(val.(string))
		//	return t.GetOperator(key)
	}

	return nil
}
