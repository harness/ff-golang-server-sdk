package evaluation

import (
	"encoding/json"

	"github.com/drone/ff-golang-server-sdk/types"

	"reflect"
	"strconv"
	"testing"

	"github.com/google/uuid"
)

func TestFeatureConfig_JsonVariation(t *testing.T) {
	v1 := "v1"
	v1Value := map[string]interface{}{
		"name":    "sdk",
		"version": "1.0",
	}

	body, err := json.Marshal(v1Value)
	if err != nil {
		t.Fail()
	}
	v1Str := string(body)

	v2 := "v2"
	v2Value := map[string]interface{}{}
	body, err = json.Marshal(v2Value)
	if err != nil {
		t.Fail()
	}
	v2Str := string(body)

	type fields struct {
		DefaultServe         Serve
		Environment          string
		Feature              string
		Kind                 string
		OffVariation         string
		Prerequisites        []Prerequisite
		Project              string
		Rules                ServingRules
		State                FeatureState
		VariationToTargetMap []VariationMap
		Variations           Variations
	}
	type args struct {
		target       *Target
		defaultValue types.JSON
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   types.JSON
	}{
		{name: "on state", fields: struct {
			DefaultServe         Serve
			Environment          string
			Feature              string
			Kind                 string
			OffVariation         string
			Prerequisites        []Prerequisite
			Project              string
			Rules                ServingRules
			State                FeatureState
			VariationToTargetMap []VariationMap
			Variations           Variations
		}{DefaultServe: struct {
			Distribution *Distribution
			Variation    *string
		}{Distribution: nil, Variation: &v1}, Environment: "dev", Feature: "flag", Kind: "json",
			OffVariation: v2, Prerequisites: nil, Project: "default", Rules: []ServingRule{}, State: "on",
			VariationToTargetMap: nil, Variations: []Variation{
				{Description: nil, Identifier: v1, Name: &v1, Value: v1Str},
				{Description: nil, Identifier: v2, Name: &v2, Value: v2Str},
			}}, args: struct {
			target       *Target
			defaultValue types.JSON
		}{target: nil, defaultValue: map[string]interface{}{
			"name":    "cf server",
			"version": "1.0",
		}}, want: v1Value},
		{name: "off state", fields: struct {
			DefaultServe         Serve
			Environment          string
			Feature              string
			Kind                 string
			OffVariation         string
			Prerequisites        []Prerequisite
			Project              string
			Rules                ServingRules
			State                FeatureState
			VariationToTargetMap []VariationMap
			Variations           Variations
		}{DefaultServe: struct {
			Distribution *Distribution
			Variation    *string
		}{Distribution: nil, Variation: &v2}, Environment: "dev", Feature: "flag", Kind: "json",
			OffVariation: v2, Prerequisites: nil, Project: "default", Rules: []ServingRule{}, State: "off",
			VariationToTargetMap: nil, Variations: []Variation{
				{Description: nil, Identifier: v1, Name: &v1, Value: v1Str},
				{Description: nil, Identifier: v2, Name: &v2, Value: v2Str},
			}}, args: struct {
			target       *Target
			defaultValue types.JSON
		}{target: nil, defaultValue: map[string]interface{}{
			"name":    "cf server",
			"version": "1.0",
		}}, want: v2Value},
	}
	for _, tt := range tests {
		val := tt
		t.Run(tt.name, func(t *testing.T) {
			fc := &FeatureConfig{
				DefaultServe:         val.fields.DefaultServe,
				Environment:          val.fields.Environment,
				Feature:              val.fields.Feature,
				Kind:                 val.fields.Kind,
				OffVariation:         val.fields.OffVariation,
				Prerequisites:        val.fields.Prerequisites,
				Project:              val.fields.Project,
				Rules:                val.fields.Rules,
				State:                val.fields.State,
				VariationToTargetMap: val.fields.VariationToTargetMap,
				Variations:           val.fields.Variations,
			}
			if got := fc.JSONVariation(val.args.target, val.args.defaultValue); !reflect.DeepEqual(got, val.want) {
				t.Errorf("JSONVariation() = %v, want %v", got, val.want)
			}
		})
	}
}

func TestFeatureConfig_StringVariation(t *testing.T) {
	v1Id := "v1"
	v1Value := "v1"
	v2Id := "v2"
	v2Value := "v2"
	type fields struct {
		DefaultServe         Serve
		Environment          string
		Feature              string
		Kind                 string
		OffVariation         string
		Prerequisites        []Prerequisite
		Project              string
		Rules                ServingRules
		State                FeatureState
		VariationToTargetMap []VariationMap
		Variations           Variations
	}
	type args struct {
		target       *Target
		defaultValue string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{name: "on variation", fields: struct {
			DefaultServe         Serve
			Environment          string
			Feature              string
			Kind                 string
			OffVariation         string
			Prerequisites        []Prerequisite
			Project              string
			Rules                ServingRules
			State                FeatureState
			VariationToTargetMap []VariationMap
			Variations           Variations
		}{DefaultServe: struct {
			Distribution *Distribution
			Variation    *string
		}{Distribution: nil, Variation: &v1Id}, Environment: "dev", Feature: "flag", Kind: "string",
			OffVariation: "v2", Prerequisites: nil, Project: "default", Rules: []ServingRule{}, State: "on",
			VariationToTargetMap: nil, Variations: []Variation{
				{Description: nil, Identifier: v1Id, Name: &v1Id, Value: v1Value},
				{Description: nil, Identifier: v2Id, Name: &v2Id, Value: v2Value},
			}}, args: struct {
			target       *Target
			defaultValue string
		}{target: nil, defaultValue: "v1"}, want: v1Value},
		{name: "off variation", fields: struct {
			DefaultServe         Serve
			Environment          string
			Feature              string
			Kind                 string
			OffVariation         string
			Prerequisites        []Prerequisite
			Project              string
			Rules                ServingRules
			State                FeatureState
			VariationToTargetMap []VariationMap
			Variations           Variations
		}{DefaultServe: struct {
			Distribution *Distribution
			Variation    *string
		}{Distribution: nil, Variation: &v1Id}, Environment: "dev", Feature: "flag", Kind: "string",
			OffVariation: "v2", Prerequisites: nil, Project: "default", Rules: []ServingRule{}, State: "off",
			VariationToTargetMap: nil, Variations: []Variation{
				{Description: nil, Identifier: v1Id, Name: &v1Id, Value: v1Value},
				{Description: nil, Identifier: v2Id, Name: &v2Id, Value: v2Value},
			}}, args: struct {
			target       *Target
			defaultValue string
		}{target: nil, defaultValue: "v1"}, want: v2Value},
	}
	for _, tt := range tests {
		val := tt
		t.Run(tt.name, func(t *testing.T) {
			fc := &FeatureConfig{
				DefaultServe:         val.fields.DefaultServe,
				Environment:          val.fields.Environment,
				Feature:              val.fields.Feature,
				Kind:                 val.fields.Kind,
				OffVariation:         val.fields.OffVariation,
				Prerequisites:        val.fields.Prerequisites,
				Project:              val.fields.Project,
				Rules:                val.fields.Rules,
				State:                val.fields.State,
				VariationToTargetMap: val.fields.VariationToTargetMap,
				Variations:           val.fields.Variations,
			}
			if got := fc.StringVariation(val.args.target, val.args.defaultValue); got != val.want {
				t.Errorf("StringVariation() = %v, want %v", got, val.want)
			}
		})
	}
}

func TestFeatureConfig_NumberVariation(t *testing.T) {
	v1Id := "v1"
	v1Value := 1.0
	v2Id := "v2"
	v2Value := 2.0
	type fields struct {
		DefaultServe         Serve
		Environment          string
		Feature              string
		Kind                 string
		OffVariation         string
		Prerequisites        []Prerequisite
		Project              string
		Rules                ServingRules
		State                FeatureState
		VariationToTargetMap []VariationMap
		Variations           Variations
	}
	type args struct {
		target       *Target
		defaultValue float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{name: "on variation", fields: struct {
			DefaultServe         Serve
			Environment          string
			Feature              string
			Kind                 string
			OffVariation         string
			Prerequisites        []Prerequisite
			Project              string
			Rules                ServingRules
			State                FeatureState
			VariationToTargetMap []VariationMap
			Variations           Variations
		}{DefaultServe: struct {
			Distribution *Distribution
			Variation    *string
		}{Distribution: nil, Variation: &v1Id}, Environment: "dev", Feature: "flag", Kind: "number",
			OffVariation: "v2", Prerequisites: nil, Project: "default", Rules: []ServingRule{}, State: "on",
			VariationToTargetMap: nil, Variations: []Variation{
				{Description: nil, Identifier: v1Id, Name: &v1Id, Value: strconv.FormatFloat(v1Value, 'f', -1, 64)},
				{Description: nil, Identifier: v2Id, Name: &v2Id, Value: strconv.FormatFloat(v2Value, 'f', -1, 64)},
			}}, args: struct {
			target       *Target
			defaultValue float64
		}{target: nil, defaultValue: 1.0}, want: v1Value},
		{name: "off variation", fields: struct {
			DefaultServe         Serve
			Environment          string
			Feature              string
			Kind                 string
			OffVariation         string
			Prerequisites        []Prerequisite
			Project              string
			Rules                ServingRules
			State                FeatureState
			VariationToTargetMap []VariationMap
			Variations           Variations
		}{DefaultServe: struct {
			Distribution *Distribution
			Variation    *string
		}{Distribution: nil, Variation: &v1Id}, Environment: "dev", Feature: "flag", Kind: "number",
			OffVariation: "v2", Prerequisites: nil, Project: "default", Rules: []ServingRule{}, State: "off",
			VariationToTargetMap: nil, Variations: []Variation{
				{Description: nil, Identifier: v1Id, Name: &v1Id, Value: strconv.FormatFloat(v1Value, 'f', -1, 64)},
				{Description: nil, Identifier: v2Id, Name: &v2Id, Value: strconv.FormatFloat(v2Value, 'f', -1, 64)},
			}}, args: struct {
			target       *Target
			defaultValue float64
		}{target: nil, defaultValue: 1.0}, want: v2Value},
	}
	for _, tt := range tests {
		val := tt
		t.Run(tt.name, func(t *testing.T) {
			fc := &FeatureConfig{
				DefaultServe:         val.fields.DefaultServe,
				Environment:          val.fields.Environment,
				Feature:              val.fields.Feature,
				Kind:                 val.fields.Kind,
				OffVariation:         val.fields.OffVariation,
				Prerequisites:        val.fields.Prerequisites,
				Project:              val.fields.Project,
				Rules:                val.fields.Rules,
				State:                val.fields.State,
				VariationToTargetMap: val.fields.VariationToTargetMap,
				Variations:           val.fields.Variations,
			}
			if got := fc.NumberVariation(val.args.target, val.args.defaultValue); got != val.want {
				t.Errorf("NumberVariation() = %v, want %v", got, val.want)
			}
		})
	}
}

func TestFeatureConfig_IntVariation(t *testing.T) {
	v1Id := "v1"
	v1Value := 5
	v2Id := "v2"
	v2Value := 9
	type fields struct {
		DefaultServe         Serve
		Environment          string
		Feature              string
		Kind                 string
		OffVariation         string
		Prerequisites        []Prerequisite
		Project              string
		Rules                ServingRules
		State                FeatureState
		VariationToTargetMap []VariationMap
		Variations           Variations
	}
	type args struct {
		target       *Target
		defaultValue int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{name: "on variation", fields: struct {
			DefaultServe         Serve
			Environment          string
			Feature              string
			Kind                 string
			OffVariation         string
			Prerequisites        []Prerequisite
			Project              string
			Rules                ServingRules
			State                FeatureState
			VariationToTargetMap []VariationMap
			Variations           Variations
		}{DefaultServe: struct {
			Distribution *Distribution
			Variation    *string
		}{Distribution: nil, Variation: &v1Id}, Environment: "dev", Feature: "flag", Kind: "int",
			OffVariation: "v2", Prerequisites: nil, Project: "default", Rules: []ServingRule{}, State: "on",
			VariationToTargetMap: nil, Variations: []Variation{
				{Description: nil, Identifier: v1Id, Name: &v1Id, Value: strconv.Itoa(v1Value)},
				{Description: nil, Identifier: v2Id, Name: &v2Id, Value: strconv.Itoa(v2Value)},
			}}, args: struct {
			target       *Target
			defaultValue int64
		}{target: nil, defaultValue: 1.0}, want: int64(v1Value)},
		{name: "off variation", fields: struct {
			DefaultServe         Serve
			Environment          string
			Feature              string
			Kind                 string
			OffVariation         string
			Prerequisites        []Prerequisite
			Project              string
			Rules                ServingRules
			State                FeatureState
			VariationToTargetMap []VariationMap
			Variations           Variations
		}{DefaultServe: struct {
			Distribution *Distribution
			Variation    *string
		}{Distribution: nil, Variation: &v1Id}, Environment: "dev", Feature: "flag", Kind: "int",
			OffVariation: "v2", Prerequisites: nil, Project: "default", Rules: []ServingRule{}, State: "off",
			VariationToTargetMap: nil, Variations: []Variation{
				{Description: nil, Identifier: v1Id, Name: &v1Id, Value: strconv.Itoa(v1Value)},
				{Description: nil, Identifier: v2Id, Name: &v2Id, Value: strconv.Itoa(v2Value)},
			}}, args: struct {
			target       *Target
			defaultValue int64
		}{target: nil, defaultValue: 1.0}, want: int64(v2Value)},
	}
	for _, tt := range tests {
		val := tt
		t.Run(tt.name, func(t *testing.T) {
			fc := &FeatureConfig{
				DefaultServe:         val.fields.DefaultServe,
				Environment:          val.fields.Environment,
				Feature:              val.fields.Feature,
				Kind:                 val.fields.Kind,
				OffVariation:         val.fields.OffVariation,
				Prerequisites:        val.fields.Prerequisites,
				Project:              val.fields.Project,
				Rules:                val.fields.Rules,
				State:                val.fields.State,
				VariationToTargetMap: val.fields.VariationToTargetMap,
				Variations:           val.fields.Variations,
			}
			if got := fc.IntVariation(val.args.target, val.args.defaultValue); got != val.want {
				t.Errorf("IntVariation() = %v, want %v", got, val.want)
			}
		})
	}
}

func TestServingRules_GetVariationName(t *testing.T) {
	f := false
	dev := "dev"
	harness := "Harness"
	onVariationIdentifier := "v1"
	offVariationIdentifier := "v2"
	m := make(map[string]interface{})
	m["email"] = "john@doe.com"
	segment := &Segment{
		Identifier:  "beta",
		Name:        "beta",
		Environment: &dev,
		Included:    []string{harness},
		Rules:       nil,
		Tags:        nil,
		Version:     1,
	}
	target := &Target{
		Identifier: harness,
		Name:       &harness,
		Anonymous:  &f,
		Attributes: &m,
	}
	type args struct {
		target       *Target
		segments     Segments
		defaultServe Serve
	}
	tests := []struct {
		name string
		sr   ServingRules
		args args
		want string
	}{
		{name: "equalOperator", sr: []ServingRule{
			{Clauses: []Clause{
				{Attribute: "identifier", ID: "id", Negate: false, Op: equalOperator, Value: []string{
					harness,
				}},
			}, Priority: 0, RuleID: uuid.New().String(), Serve: struct {
				Distribution *Distribution
				Variation    *string
			}{Distribution: nil, Variation: &onVariationIdentifier}},
		}, args: struct {
			target       *Target
			segments     Segments
			defaultServe Serve
		}{target: target, defaultServe: struct {
			Distribution *Distribution
			Variation    *string
		}{Distribution: nil, Variation: &onVariationIdentifier}}, want: onVariationIdentifier},
		{name: "equal with rules", sr: []ServingRule{
			{Clauses: []Clause{
				{Attribute: "identifier", ID: "id", Negate: false, Op: equalOperator, Value: []string{
					harness,
				}},
			}, Priority: 0, RuleID: uuid.NewString(), Serve: struct {
				Distribution *Distribution
				Variation    *string
			}{Distribution: nil, Variation: &offVariationIdentifier}},
		}, args: struct {
			target       *Target
			segments     Segments
			defaultServe Serve
		}{target: target, defaultServe: struct {
			Distribution *Distribution
			Variation    *string
		}{Distribution: nil, Variation: &onVariationIdentifier}}, want: offVariationIdentifier},
		//
		{name: "segment match", sr: []ServingRule{
			{Clauses: []Clause{
				{Op: segmentMatchOperator, Value: []string{
					segment.Identifier,
				}},
			}, Priority: 0, RuleID: uuid.NewString(), Serve: struct {
				Distribution *Distribution
				Variation    *string
			}{Distribution: nil, Variation: &offVariationIdentifier}},
		}, args: struct {
			target       *Target
			segments     Segments
			defaultServe Serve
		}{target: target, segments: Segments{segment.Identifier: segment}, defaultServe: struct {
			Distribution *Distribution
			Variation    *string
		}{Distribution: nil, Variation: &onVariationIdentifier}}, want: offVariationIdentifier},
	}
	for _, tt := range tests {
		val := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := val.sr.GetVariationName(val.args.target, val.args.segments, val.args.defaultServe); got != val.want {
				t.Errorf("GetVariationName() = %v, want %v", got, val.want)
			}
		})
	}
}

func TestFeatureConfig_Evaluate(t *testing.T) {
	f := false
	harness := "Harness"
	v1 := "v1"
	v2 := "v2"
	m := make(map[string]interface{})
	m["email"] = "john@doe.com"
	target := Target{
		Identifier: harness,
		Name:       nil,
		Anonymous:  &f,
		Attributes: &m,
	}
	type fields struct {
		DefaultServe         Serve
		Environment          string
		Feature              string
		Kind                 string
		OffVariation         string
		Prerequisites        []Prerequisite
		Project              string
		Rules                ServingRules
		State                FeatureState
		VariationToTargetMap []VariationMap
		Variations           Variations
	}
	type args struct {
		target *Target
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Evaluation
	}{
		{name: "happy path", fields: struct {
			DefaultServe         Serve
			Environment          string
			Feature              string
			Kind                 string
			OffVariation         string
			Prerequisites        []Prerequisite
			Project              string
			Rules                ServingRules
			State                FeatureState
			VariationToTargetMap []VariationMap
			Variations           Variations
		}{DefaultServe: struct {
			Distribution *Distribution
			Variation    *string
		}{Distribution: nil, Variation: &v1}, Environment: "dev", Feature: "flag", Kind: "boolean",
			OffVariation: v2, Prerequisites: nil, Project: "default", Rules: []ServingRule{
				{Clauses: []Clause{
					{Attribute: "identifier", ID: "id", Negate: false, Op: equalOperator, Value: []string{
						harness,
					}},
				}, Priority: 0, RuleID: uuid.NewString(), Serve: struct {
					Distribution *Distribution
					Variation    *string
				}{Distribution: nil, Variation: &v2}},
			}, State: "on", VariationToTargetMap: nil, Variations: []Variation{
				{Description: nil, Identifier: v1, Name: &v1, Value: "true"},
				{Description: nil, Identifier: v2, Name: &v2, Value: "false"},
			}}, args: struct{ target *Target }{target: &target}, want: &Evaluation{
			Flag:  "flag",
			Value: false,
		}},
	}
	for _, tt := range tests {
		val := tt
		t.Run(tt.name, func(t *testing.T) {
			fc := FeatureConfig{
				DefaultServe:         val.fields.DefaultServe,
				Environment:          val.fields.Environment,
				Feature:              val.fields.Feature,
				Kind:                 val.fields.Kind,
				OffVariation:         val.fields.OffVariation,
				Prerequisites:        val.fields.Prerequisites,
				Project:              val.fields.Project,
				Rules:                val.fields.Rules,
				State:                val.fields.State,
				VariationToTargetMap: val.fields.VariationToTargetMap,
				Variations:           val.fields.Variations,
			}
			if got := fc.Evaluate(val.args.target); !reflect.DeepEqual(got, val.want) {
				t.Errorf("Evaluate() = %v, want %v", got, val.want)
			}
		})
	}
}

func TestClause_Evaluate(t *testing.T) {
	f := false
	type fields struct {
		Attribute string
		ID        string
		Negate    bool
		Op        string
		Value     []string
	}
	type args struct {
		target   *Target
		segments Segments
		operator types.ValueType
	}
	m := make(map[string]interface{})
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
		{name: "segment match operator (include)", fields: struct {
			Attribute string
			ID        string
			Negate    bool
			Op        string
			Value     []string
		}{Op: segmentMatchOperator, Value: []string{"beta"}},
			args: struct {
				target   *Target
				segments Segments
				operator types.ValueType
			}{target: &target, segments: map[string]*Segment{
				"beta": {
					Identifier: "beta",
					Name:       "Beta users",
					Included:   []string{target.Identifier},
				},
			}, operator: nil}, want: true},
	}
	for _, tt := range tests {
		val := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &Clause{
				Attribute: val.fields.Attribute,
				ID:        val.fields.ID,
				Negate:    val.fields.Negate,
				Op:        val.fields.Op,
				Value:     val.fields.Value,
			}
			if got := c.Evaluate(val.args.target, val.args.segments, val.args.operator); got != val.want {
				t.Errorf("Evaluate() = %v, want %v", got, val.want)
			}
		})
	}
}
