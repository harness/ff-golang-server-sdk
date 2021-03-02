package evaluation

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/wings-software/ff-client-sdk-go/types"
	"reflect"
	"strconv"
	"testing"
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
		t.Run(tt.name, func(t *testing.T) {
			fc := &FeatureConfig{
				DefaultServe:         tt.fields.DefaultServe,
				Environment:          tt.fields.Environment,
				Feature:              tt.fields.Feature,
				Kind:                 tt.fields.Kind,
				OffVariation:         tt.fields.OffVariation,
				Prerequisites:        tt.fields.Prerequisites,
				Project:              tt.fields.Project,
				Rules:                tt.fields.Rules,
				State:                tt.fields.State,
				VariationToTargetMap: tt.fields.VariationToTargetMap,
				Variations:           tt.fields.Variations,
			}
			if got := fc.JsonVariation(tt.args.target, tt.args.defaultValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonVariation() = %v, want %v", got, tt.want)
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
		t.Run(tt.name, func(t *testing.T) {
			fc := &FeatureConfig{
				DefaultServe:         tt.fields.DefaultServe,
				Environment:          tt.fields.Environment,
				Feature:              tt.fields.Feature,
				Kind:                 tt.fields.Kind,
				OffVariation:         tt.fields.OffVariation,
				Prerequisites:        tt.fields.Prerequisites,
				Project:              tt.fields.Project,
				Rules:                tt.fields.Rules,
				State:                tt.fields.State,
				VariationToTargetMap: tt.fields.VariationToTargetMap,
				Variations:           tt.fields.Variations,
			}
			if got := fc.StringVariation(tt.args.target, tt.args.defaultValue); got != tt.want {
				t.Errorf("StringVariation() = %v, want %v", got, tt.want)
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
		t.Run(tt.name, func(t *testing.T) {
			fc := &FeatureConfig{
				DefaultServe:         tt.fields.DefaultServe,
				Environment:          tt.fields.Environment,
				Feature:              tt.fields.Feature,
				Kind:                 tt.fields.Kind,
				OffVariation:         tt.fields.OffVariation,
				Prerequisites:        tt.fields.Prerequisites,
				Project:              tt.fields.Project,
				Rules:                tt.fields.Rules,
				State:                tt.fields.State,
				VariationToTargetMap: tt.fields.VariationToTargetMap,
				Variations:           tt.fields.Variations,
			}
			if got := fc.NumberVariation(tt.args.target, tt.args.defaultValue); got != tt.want {
				t.Errorf("NumberVariation() = %v, want %v", got, tt.want)
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
		t.Run(tt.name, func(t *testing.T) {
			fc := &FeatureConfig{
				DefaultServe:         tt.fields.DefaultServe,
				Environment:          tt.fields.Environment,
				Feature:              tt.fields.Feature,
				Kind:                 tt.fields.Kind,
				OffVariation:         tt.fields.OffVariation,
				Prerequisites:        tt.fields.Prerequisites,
				Project:              tt.fields.Project,
				Rules:                tt.fields.Rules,
				State:                tt.fields.State,
				VariationToTargetMap: tt.fields.VariationToTargetMap,
				Variations:           tt.fields.Variations,
			}
			if got := fc.IntVariation(tt.args.target, tt.args.defaultValue); got != tt.want {
				t.Errorf("IntVariation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServingRules_GetVariationName(t *testing.T) {

	harness := "Harness"
	onVariationIdentifier := "v1"
	offVariationIdentifier := "v2"

	target := &Target{
		Identifier: harness,
		Name:       &harness,
		Anonymous:  false,
		Attributes: nil,
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
				{Attribute: "identifier", Id: "id", Negate: false, Op: equalOperator, Value: []string{
					harness,
				}},
			}, Priority: 0, RuleId: uuid.New().String(), Serve: struct {
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
				{Attribute: "identifier", Id: "id", Negate: false, Op: equalOperator, Value: []string{
					harness,
				}},
			}, Priority: 0, RuleId: uuid.NewString(), Serve: struct {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.GetVariationName(tt.args.target, tt.args.segments, tt.args.defaultServe); got != tt.want {
				t.Errorf("GetVariationName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeatureConfig_Evaluate(t *testing.T) {
	harness := "Harness"
	v1 := "v1"
	v2 := "v2"
	target := Target{
		Identifier: harness,
		Name:       nil,
		Anonymous:  false,
		Attributes: nil,
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
					{Attribute: "identifier", Id: "id", Negate: false, Op: equalOperator, Value: []string{
						harness,
					}},
				}, Priority: 0, RuleId: uuid.NewString(), Serve: struct {
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
		t.Run(tt.name, func(t *testing.T) {
			fc := FeatureConfig{
				DefaultServe:         tt.fields.DefaultServe,
				Environment:          tt.fields.Environment,
				Feature:              tt.fields.Feature,
				Kind:                 tt.fields.Kind,
				OffVariation:         tt.fields.OffVariation,
				Prerequisites:        tt.fields.Prerequisites,
				Project:              tt.fields.Project,
				Rules:                tt.fields.Rules,
				State:                tt.fields.State,
				VariationToTargetMap: tt.fields.VariationToTargetMap,
				Variations:           tt.fields.Variations,
			}
			if got := fc.Evaluate(tt.args.target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClause_Evaluate(t *testing.T) {
	type fields struct {
		Attribute string
		Id        string
		Negate    bool
		Op        string
		Value     []string
	}
	type args struct {
		target   *Target
		segments Segments
		operator types.ValueType
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
		{name: "segment match operator", fields: struct {
			Attribute string
			Id        string
			Negate    bool
			Op        string
			Value     []string
		}{Attribute: "identifier", Id: uuid.New().String(), Negate: false, Op: segmentMatchOperator, Value: []string{"beta"}},
			args: struct {
				target   *Target
				segments Segments
				operator types.ValueType
			}{target: &target, segments: map[string]*Segment{
				"beta": {
					Identifier:  "beta",
					Name:        "Beta users",
					CreatedAt:   nil,
					ModifiedAt:  nil,
					Environment: nil,
					Excluded:    nil,
					Included:    []string{target.Identifier},
					Rules:       nil,
					Tags:        nil,
					Version:     0,
				},
			}, operator: types.String("john@doe.com")}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Clause{
				Attribute: tt.fields.Attribute,
				Id:        tt.fields.Id,
				Negate:    tt.fields.Negate,
				Op:        tt.fields.Op,
				Value:     tt.fields.Value,
			}
			if got := c.Evaluate(tt.args.target, tt.args.segments, tt.args.operator); got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}
