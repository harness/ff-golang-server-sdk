package evaluation

import (
	"reflect"
	"testing"

	"github.com/drone/ff-golang-server-sdk.v0/types"
)

func TestTarget_GetOperator(t1 *testing.T) {
	harness := "Harness"
	type fields struct {
		Identifier string
		Name       *string
		Anonymous  bool
		Attributes map[string]interface{}
	}
	type args struct {
		attr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   types.ValueType
	}{
		{name: "boolean operator", fields: struct {
			Identifier string
			Name       *string
			Anonymous  bool
			Attributes map[string]interface{}
		}{Identifier: "harness", Name: &harness, Anonymous: false, Attributes: nil},
			args: struct{ attr string }{attr: "anonymous"}, want: types.Boolean(false)},
		{name: "string operator", fields: struct {
			Identifier string
			Name       *string
			Anonymous  bool
			Attributes map[string]interface{}
		}{Identifier: "harness", Name: &harness, Anonymous: false, Attributes: nil},
			args: struct{ attr string }{attr: "identifier"}, want: types.String("harness")},
		{name: "int operator", fields: struct {
			Identifier string
			Name       *string
			Anonymous  bool
			Attributes map[string]interface{}
		}{Identifier: "harness", Name: &harness, Anonymous: false, Attributes: map[string]interface{}{
			"order": 1,
		}},
			args: struct{ attr string }{attr: "order"}, want: types.Integer(1)},
		{name: "number operator", fields: struct {
			Identifier string
			Name       *string
			Anonymous  bool
			Attributes map[string]interface{}
		}{Identifier: "harness", Name: &harness, Anonymous: false, Attributes: map[string]interface{}{
			"weight": 99.99,
		}},
			args: struct{ attr string }{attr: "weight"}, want: types.Number(99.99)},
	}
	for _, tt := range tests {
		val := tt
		t1.Run(val.name, func(t1 *testing.T) {
			t := Target{
				Identifier: val.fields.Identifier,
				Name:       val.fields.Name,
				Anonymous:  val.fields.Anonymous,
				Attributes: val.fields.Attributes,
			}
			if got := t.GetOperator(val.args.attr); !reflect.DeepEqual(got, val.want) {
				t1.Errorf("GetOperator() = %v, want %v", got, val.want)
			}
		})
	}
}

func TestTarget_GetAttrValue(t1 *testing.T) {
	name := "John"
	identifier := "john"
	email := "john@doe.com"
	type fields struct {
		Identifier string
		Name       *string
		Anonymous  bool
		Attributes map[string]interface{}
	}
	type args struct {
		attr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   reflect.Value
	}{
		{name: "check identifier", fields: struct {
			Identifier string
			Name       *string
			Anonymous  bool
			Attributes map[string]interface{}
		}{Identifier: identifier, Name: &name, Anonymous: false, Attributes: types.JSON{}},
			args: struct{ attr string }{attr: "identifier"}, want: reflect.ValueOf(identifier)},
		{name: "check attributes", fields: struct {
			Identifier string
			Name       *string
			Anonymous  bool
			Attributes map[string]interface{}
		}{Identifier: "john", Name: &name, Anonymous: false, Attributes: types.JSON{
			"email": email,
		}},
			args: struct{ attr string }{attr: "email"}, want: reflect.ValueOf(email)},
	}
	for _, tt := range tests {
		val := tt
		t1.Run(val.name, func(t1 *testing.T) {
			t := Target{
				Identifier: val.fields.Identifier,
				Name:       val.fields.Name,
				Anonymous:  val.fields.Anonymous,
				Attributes: val.fields.Attributes,
			}
			if got := t.GetAttrValue(val.args.attr); !reflect.DeepEqual(got.Interface(), val.want.Interface()) {
				t1.Errorf("GetAttrValue() = %v, want %v", got, val.want)
			}
		})
	}
}

func TestTarget_GetOperator1(t1 *testing.T) {
	type fields struct {
		Identifier string
		Name       *string
		Anonymous  bool
		Attributes map[string]interface{}
	}
	type args struct {
		attr string
	}

	name := "John"
	tests := []struct {
		name   string
		fields fields
		args   args
		want   types.ValueType
	}{
		{name: "bool operator", fields: struct {
			Identifier string
			Name       *string
			Anonymous  bool
			Attributes map[string]interface{}
		}{Identifier: "john", Name: &name, Anonymous: false, Attributes: types.JSON{}},
			args: struct{ attr string }{attr: "anonymous"}, want: types.Boolean(false)},
		{name: "string operator", fields: struct {
			Identifier string
			Name       *string
			Anonymous  bool
			Attributes map[string]interface{}
		}{Identifier: "john", Name: &name, Anonymous: false, Attributes: types.JSON{}},
			args: struct{ attr string }{attr: "identifier"}, want: types.String("john")},
		{name: "number operator", fields: struct {
			Identifier string
			Name       *string
			Anonymous  bool
			Attributes map[string]interface{}
		}{Identifier: "john", Name: &name, Anonymous: false, Attributes: types.JSON{
			"height": 186.5,
		}},
			args: struct{ attr string }{attr: "height"}, want: types.Number(186.5)},
		{name: "integer operator", fields: struct {
			Identifier string
			Name       *string
			Anonymous  bool
			Attributes map[string]interface{}
		}{Identifier: "john", Name: &name, Anonymous: false, Attributes: types.JSON{
			"zip": 90210,
		}},
			args: struct{ attr string }{attr: "zip"}, want: types.Integer(90210)},
	}
	for _, tt := range tests {
		val := tt
		t1.Run(val.name, func(t1 *testing.T) {
			t := Target{
				Identifier: val.fields.Identifier,
				Name:       val.fields.Name,
				Anonymous:  val.fields.Anonymous,
				Attributes: val.fields.Attributes,
			}
			if got := t.GetOperator(val.args.attr); !reflect.DeepEqual(got, val.want) {
				t1.Errorf("GetOperator() = %v, want %v", got, val.want)
			}
		})
	}
}
