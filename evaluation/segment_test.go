package evaluation

import (
	"testing"
)

func TestSegment_Evaluate(t *testing.T) {
	type fields struct {
		Identifier string
		Excluded   StrSlice
		Included   StrSlice
		Rules      SegmentRules
	}

	f := false
	m := make(map[string]interface{})
	m["email"] = "john@doe.com"
	target := Target{
		Identifier: "john",
		Anonymous:  &f,
		Attributes: &m,
	}

	tests := []struct {
		name   string
		fields fields
		args   Target
		want   bool
	}{
		{name: "test target included by list", fields: fields{Identifier: "beta", Included: []string{"john"}}, args: target, want: true},
		{name: "test target excluded by list", fields: fields{Identifier: "beta", Included: []string{"john"}, Excluded: []string{"john"}}, args: target, want: false},
		{name: "test target included by rules", fields: fields{Identifier: "beta", Rules: []Clause{{Attribute: "email", ID: "1", Op: equalOperator, Value: []string{"john@doe.com"}}}}, args: target, want: true},
		{name: "test target not included by rules", fields: fields{Identifier: "beta", Rules: []Clause{{Attribute: "email", ID: "2", Op: equalOperator, Value: []string{"foo@doe.com"}}}}, args: target, want: false},
		{name: "test target rules evaluating with OR", fields: fields{Identifier: "beta", Rules: []Clause{
			{Attribute: "email", ID: "1", Op: equalOperator, Value: []string{"john@doe.com"}},
			{Attribute: "email", ID: "2", Op: equalOperator, Value: []string{"foo@doe.com"}},
		}}, args: target, want: true},
	}
	for _, tt := range tests {
		val := tt
		t.Run(val.name, func(t *testing.T) {
			s := Segment{
				Identifier: val.fields.Identifier,
				Excluded:   val.fields.Excluded,
				Included:   val.fields.Included,
				Rules:      val.fields.Rules,
			}
			if got := s.Evaluate(&val.args); got != val.want {
				t.Errorf("Evaluate() = %v, want %v", got, val.want)
			}
		})
	}
}

func TestSegments_Evaluate(t *testing.T) {
	f := false
	m := make(map[string]interface{})
	m["email"] = "john@doe.com"
	target := Target{
		Identifier: "john",
		Anonymous:  &f,
		Attributes: &m,
	}

	tests := map[string]struct {
		segments Segments
		target   Target
		want     bool
	}{
		"test target included by segment alpha returns true": {
			segments: Segments{"alpha": {Identifier: "alpha", Included: []string{target.Identifier}}},
			target:   target,
			want:     true,
		},
		"test target not included segment alpha, but included in beta returns true": {
			segments: Segments{
				"alpha": {Identifier: "alpha", Included: []string{}},
				"beta":  {Identifier: "beta", Included: []string{target.Identifier}},
			},
			target: target,
			want:   true,
		},
		"test target not included segment alpha, and not included in beta returns false": {
			segments: Segments{
				"alpha": {Identifier: "alpha", Included: []string{}},
				"beta":  {Identifier: "beta", Included: []string{}},
			},
			target: target,
			want:   false,
		},
		"test target included segment alpha, and excluded in beta returns true": {
			segments: Segments{
				"alpha": {Identifier: "alpha", Included: []string{target.Identifier}},
				"beta":  {Identifier: "beta", Excluded: []string{target.Identifier}},
			},
			target: target,
			want:   true,
		},
		"test target excluded from segment alpha, and included in beta returns true": {
			segments: Segments{
				"alpha": {Identifier: "alpha", Excluded: []string{target.Identifier}},
				"beta":  {Identifier: "beta", Included: []string{target.Identifier}},
			},
			target: target,
			want:   true,
		},
	}
	for name, tt := range tests {
		val := tt
		t.Run(name, func(t *testing.T) {
			s := val.segments
			if got := s.Evaluate(&val.target); got != val.want {
				t.Errorf("Evaluate() = %v, want %v", got, val.want)
			}
		})
	}
}
