package evaluation

import (
	"reflect"
	"testing"
)

func TestGetStructFieldValue(t *testing.T) {
	identifier := "john"
	name := "John"
	target := Target{
		Identifier: identifier,
		Name:       &name,
		Anonymous:  false,
		Attributes: map[string]interface{}{
			"email": "john@doe.com",
		},
	}
	type args struct {
		target interface{}
		attr   string
	}
	tests := []struct {
		name string
		args args
		want reflect.Value
	}{
		{name: "check identifier", args: struct {
			target interface{}
			attr   string
		}{target: target, attr: "identifier"}, want: reflect.ValueOf(identifier)},
		{name: "check name", args: struct {
			target interface{}
			attr   string
		}{target: target, attr: "name"}, want: reflect.ValueOf(name)},
		{name: "check anonymous", args: struct {
			target interface{}
			attr   string
		}{target: target, attr: "anonymous"}, want: reflect.ValueOf(false)},
		{name: "check attributes", args: struct {
			target interface{}
			attr   string
		}{target: target, attr: "attributes"}, want: reflect.ValueOf(target.Attributes)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStructFieldValue(tt.args.target, tt.args.attr); got == tt.want {
				t.Errorf("GetStructFieldValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_caseInsensitiveFieldByName(t *testing.T) {
	identifier := "john"
	name := "John"
	target := Target{
		Identifier: identifier,
		Name:       &name,
		Anonymous:  false,
		Attributes: map[string]interface{}{
			"email": "john@doe.com",
		},
	}
	type args struct {
		v    reflect.Value
		name string
	}
	tests := []struct {
		name string
		args args
		want reflect.Value
	}{
		{name: "check with struct", args: struct {
			v    reflect.Value
			name string
		}{v: reflect.ValueOf(target), name: "identifier"}, want: reflect.ValueOf("john")},
		{name: "check with other than struct", args: struct {
			v    reflect.Value
			name string
		}{v: reflect.ValueOf("Identifier"), name: "identifier"}, want: reflect.ValueOf("Identifier")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := caseInsensitiveFieldByName(tt.args.v, tt.args.name); got == tt.want {
				t.Errorf("caseInsensitiveFieldByName() = %v, want %v", got, tt.want)
			}
		})
	}
}
