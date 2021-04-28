package evaluation

import (
	"github.com/google/uuid"

	"testing"
)

func TestSegment_Evaluate(t *testing.T) {
	type fields struct {
		Identifier  string
		Name        string
		CreatedAt   *int64
		ModifiedAt  *int64
		Environment *string
		Excluded    StrSlice
		Included    StrSlice
		Rules       Clauses
		Tags        []Tag
		Version     int64
	}
	type args struct {
		target *Target
	}
	f := false
	m := make(map[string]interface{}, 0)
	m["email"] = "john@doe.com"
	target := Target{
		Identifier: "john",
		Name:       nil,
		Anonymous:  &f,
		Attributes: &m,
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{name: "test included", fields: struct {
			Identifier  string
			Name        string
			CreatedAt   *int64
			ModifiedAt  *int64
			Environment *string
			Excluded    StrSlice
			Included    StrSlice
			Rules       Clauses
			Tags        []Tag
			Version     int64
		}{Identifier: "beta", Name: "Beta users", CreatedAt: nil, ModifiedAt: nil, Environment: nil, Excluded: nil,
			Included: []string{"john"}, Rules: nil, Tags: nil, Version: 1}, args: struct{ target *Target }{target: &target}, want: true},
		{name: "test rules", fields: struct {
			Identifier  string
			Name        string
			CreatedAt   *int64
			ModifiedAt  *int64
			Environment *string
			Excluded    StrSlice
			Included    StrSlice
			Rules       Clauses
			Tags        []Tag
			Version     int64
		}{Identifier: "beta", Name: "Beta users", CreatedAt: nil, ModifiedAt: nil, Environment: nil, Excluded: nil,
			Included: nil, Rules: []Clause{
				{Attribute: "email", ID: uuid.New().String(), Negate: false, Op: equalOperator, Value: []string{"john@doe.com"}},
			}, Tags: nil, Version: 1}, args: struct{ target *Target }{target: &target}, want: true},
	}
	for _, tt := range tests {
		val := tt
		t.Run(val.name, func(t *testing.T) {
			s := Segment{
				Identifier:  val.fields.Identifier,
				Name:        val.fields.Name,
				CreatedAt:   val.fields.CreatedAt,
				ModifiedAt:  val.fields.ModifiedAt,
				Environment: val.fields.Environment,
				Excluded:    val.fields.Excluded,
				Included:    val.fields.Included,
				Rules:       val.fields.Rules,
				Tags:        val.fields.Tags,
				Version:     val.fields.Version,
			}
			if got := s.Evaluate(val.args.target); got != val.want {
				t.Errorf("Evaluate() = %v, want %v", got, val.want)
			}
		})
	}
}
