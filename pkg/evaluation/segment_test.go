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

	target := Target{
		Identifier: "john",
		Name:       nil,
		Anonymous:  false,
		Attributes: map[string]interface{}{
			"email": "john@doe.com",
		},
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
				{Attribute: "email", Id: uuid.New().String(), Negate: false, Op: equalOperator, Value: []string{"john@doe.com"}},
			}, Tags: nil, Version: 1}, args: struct{ target *Target }{target: &target}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Segment{
				Identifier:  tt.fields.Identifier,
				Name:        tt.fields.Name,
				CreatedAt:   tt.fields.CreatedAt,
				ModifiedAt:  tt.fields.ModifiedAt,
				Environment: tt.fields.Environment,
				Excluded:    tt.fields.Excluded,
				Included:    tt.fields.Included,
				Rules:       tt.fields.Rules,
				Tags:        tt.fields.Tags,
				Version:     tt.fields.Version,
			}
			if got := s.Evaluate(tt.args.target); got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}
