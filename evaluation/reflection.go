package evaluation

import (
	"reflect"
	"strings"
)

func caseInsensitiveFieldByName(v reflect.Value, name string) reflect.Value {
	if v.Kind() != reflect.Struct {
		return v
	}
	name = strings.ToLower(name)
	return v.FieldByNameFunc(func(n string) bool { return strings.ToLower(n) == name })
}

// GetStructFieldValue returns struct field value or simple value using attr field
func GetStructFieldValue(target interface{}, attr string) reflect.Value {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() == reflect.Struct {
		return caseInsensitiveFieldByName(targetValue, attr)
	}
	return targetValue
}
