package evaluation

import (
	"encoding/json"

	"github.com/stretchr/testify/assert"

	"github.com/drone/ff-golang-server-sdk/types"

	"strconv"
	"testing"

	"github.com/google/uuid"
)

const (
	boolKind   = "boolean"
	stringKind = "string"
	numberKind = "number"
	intKind    = "int"
	jsonKind   = "json"
)

func TestFeatureConfig_JsonVariation(t *testing.T) {

	v1Value, err := json.Marshal(map[string]interface{}{
		"name":    "sdk",
		"version": "1.0",
	})
	if err != nil {
		t.Fail()
	}

	v2Value, err := json.Marshal(map[string]interface{}{
		"name":    "sdk",
		"version": "2.0",
	})
	if err != nil {
		t.Fail()
	}

	jsonflagName := "SimpleJSON"

	on := Variation{
		Name:       stringPtr("On"),
		Value:      string(v1Value),
		Identifier: "on",
	}

	off := Variation{
		Name:       stringPtr("Off"),
		Value:      string(v2Value),
		Identifier: "off",
	}

	other := Variation{
		Name:       stringPtr("Other"),
		Value:      "",
		Identifier: "other",
	}

	tests := []struct {
		name          string
		featureConfig FeatureConfig
		want          Variation
		wantErr       bool
	}{
		{
			name:          "Test on variation returned when flag is on",
			featureConfig: makeFeatureConfig(jsonflagName, jsonKind, on, off, on, FeatureStateOn, nil),
			want:          on,
			wantErr:       false,
		},
		{
			name:          "Test off variation returned when flag is off",
			featureConfig: makeFeatureConfig(jsonflagName, jsonKind, on, off, on, FeatureStateOff, nil),
			want:          off,
			wantErr:       false,
		},
		{
			name:          "Test error returned when invalid default serve is provided",
			featureConfig: makeFeatureConfig(jsonflagName, jsonKind, on, off, other, FeatureStateOn, nil),
			want:          Variation{},
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		val := tt
		t.Run(tt.name, func(t *testing.T) {
			fc := val.featureConfig
			got, err := fc.JSONVariation(nil)
			if (err != nil) != val.wantErr {
				t.Errorf("JSONVariation() error = %v, wantErr %v", err, val.wantErr)
				return
			}
			assert.Equal(t, val.want, got, "JSONVariation() = %v, want %v", got, val.want)
		})
	}
}

func TestFeatureConfig_StringVariation(t *testing.T) {
	stringflagName := "SimpleString"

	on := Variation{
		Name:       stringPtr("On"),
		Value:      "v1",
		Identifier: "on",
	}

	off := Variation{
		Name:       stringPtr("Off"),
		Value:      "v2",
		Identifier: "off",
	}

	other := Variation{
		Name:       stringPtr("Other"),
		Value:      "v3",
		Identifier: "other",
	}

	tests := []struct {
		name          string
		featureConfig FeatureConfig
		want          Variation
		wantErr       bool
	}{
		{
			name:          "Test on variation returned when flag is on",
			featureConfig: makeFeatureConfig(stringflagName, stringKind, on, off, on, FeatureStateOn, nil),
			want:          on,
			wantErr:       false,
		},
		{
			name:          "Test off variation returned when flag is off",
			featureConfig: makeFeatureConfig(stringflagName, stringKind, on, off, on, FeatureStateOff, nil),
			want:          off,
			wantErr:       false,
		},
		{
			name:          "Test error returned when invalid default serve is provided",
			featureConfig: makeFeatureConfig(stringflagName, stringKind, on, off, other, FeatureStateOn, nil),
			want:          Variation{},
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		val := tt
		t.Run(tt.name, func(t *testing.T) {
			fc := val.featureConfig
			got, err := fc.StringVariation(nil)
			if (err != nil) != val.wantErr {
				t.Errorf("StringVariation() error = %v, wantErr %v", err, val.wantErr)
				return
			}
			assert.Equal(t, val.want, got, "StringVariation() = %v, want %v", got, val.want)
		})
	}
}

func TestFeatureConfig_NumberVariation(t *testing.T) {
	numberflagName := "SimpleNumber"

	on := Variation{
		Name:       stringPtr("On"),
		Value:      strconv.FormatFloat(1.0, 'f', -1, 64),
		Identifier: "on",
	}

	off := Variation{
		Name:       stringPtr("Off"),
		Value:      strconv.FormatFloat(2.0, 'f', -1, 64),
		Identifier: "off",
	}

	other := Variation{
		Name:       stringPtr("Other"),
		Value:      strconv.FormatFloat(3.0, 'f', -1, 64),
		Identifier: "other",
	}

	tests := []struct {
		name          string
		featureConfig FeatureConfig
		want          Variation
		wantErr       bool
	}{
		{
			name:          "Test on variation returned when flag is on",
			featureConfig: makeFeatureConfig(numberflagName, numberKind, on, off, on, FeatureStateOn, nil),
			want:          on,
			wantErr:       false,
		},
		{
			name:          "Test off variation returned when flag is off",
			featureConfig: makeFeatureConfig(numberflagName, numberKind, on, off, on, FeatureStateOff, nil),
			want:          off,
			wantErr:       false,
		},
		{
			name:          "Test error returned when invalid default serve is provided",
			featureConfig: makeFeatureConfig(numberflagName, numberKind, on, off, other, FeatureStateOn, nil),
			want:          Variation{},
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		val := tt
		t.Run(tt.name, func(t *testing.T) {
			fc := val.featureConfig
			got, err := fc.NumberVariation(nil)
			if (err != nil) != val.wantErr {
				t.Errorf("NumberVariation() error = %v, wantErr %v", err, val.wantErr)
				return
			}
			assert.Equal(t, val.want, got, "NumberVariation() = %v, want %v", got, val.want)
		})
	}
}

func TestFeatureConfig_IntVariation(t *testing.T) {
	intflagName := "SimpleInt"

	on := Variation{
		Name:       stringPtr("On"),
		Value:      strconv.FormatInt(4, 10),
		Identifier: "on",
	}

	off := Variation{
		Name:       stringPtr("Off"),
		Value:      strconv.FormatInt(9, 10),
		Identifier: "off",
	}

	other := Variation{
		Name:       stringPtr("Other"),
		Value:      strconv.FormatInt(19, 10),
		Identifier: "other",
	}

	tests := []struct {
		name          string
		featureConfig FeatureConfig
		want          Variation
		wantErr       bool
	}{
		{
			name:          "Test on variation returned when flag is on",
			featureConfig: makeFeatureConfig(intflagName, intKind, on, off, on, FeatureStateOn, nil),
			want:          on,
			wantErr:       false,
		},
		{
			name:          "Test off variation returned when flag is off",
			featureConfig: makeFeatureConfig(intflagName, intKind, on, off, on, FeatureStateOff, nil),
			want:          off,
			wantErr:       false,
		},
		{
			name:          "Test error returned when invalid default serve is provided",
			featureConfig: makeFeatureConfig(intflagName, intKind, on, off, other, FeatureStateOn, nil),
			want:          Variation{},
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		val := tt
		t.Run(tt.name, func(t *testing.T) {
			fc := val.featureConfig
			got, err := fc.IntVariation(nil)
			if (err != nil) != val.wantErr {
				t.Errorf("IntVariation() error = %v, wantErr %v", err, val.wantErr)
				return
			}
			assert.Equal(t, val.want, got, "IntVariation() = %v, want %v", got, val.want)
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
		{
			name: "happy path, no rules",
			sr:   []ServingRule{},
			args: struct {
				target       *Target
				segments     Segments
				defaultServe Serve
			}{
				target: target,
				defaultServe: struct {
					Distribution *Distribution
					Variation    *string
				}{
					Distribution: nil,
					Variation:    &onVariationIdentifier,
				},
			},
			want: onVariationIdentifier,
		},
		{
			name: "equalOperator",
			sr: []ServingRule{
				{
					Clauses: []Clause{
						{
							Attribute: "identifier",
							ID:        "id",
							Negate:    false,
							Op:        equalOperator,
							Value: []string{
								harness,
							},
						},
					},
					Priority: 0,
					RuleID:   uuid.New().String(),
					Serve: struct {
						Distribution *Distribution
						Variation    *string
					}{
						Distribution: nil,
						Variation:    &onVariationIdentifier,
					},
				},
			},
			args: struct {
				target       *Target
				segments     Segments
				defaultServe Serve
			}{
				target: target, defaultServe: struct {
					Distribution *Distribution
					Variation    *string
				}{
					Distribution: nil,
					Variation:    &onVariationIdentifier,
				},
			},
			want: onVariationIdentifier,
		},
		{
			name: "equal with rules",
			sr: []ServingRule{
				{
					Clauses: []Clause{
						{
							Attribute: "identifier",
							ID:        "id",
							Negate:    false,
							Op:        equalOperator,
							Value: []string{
								harness,
							},
						},
					},
					Priority: 0,
					RuleID:   uuid.NewString(),
					Serve: struct {
						Distribution *Distribution
						Variation    *string
					}{
						Distribution: nil,
						Variation:    &offVariationIdentifier,
					},
				},
			},
			args: struct {
				target       *Target
				segments     Segments
				defaultServe Serve
			}{
				target: target,
				defaultServe: struct {
					Distribution *Distribution
					Variation    *string
				}{
					Distribution: nil,
					Variation:    &onVariationIdentifier,
				},
			},
			want: offVariationIdentifier,
		},
		{
			name: "segment match",
			sr: []ServingRule{
				{
					Clauses: []Clause{
						{
							Op: segmentMatchOperator,
							Value: []string{
								segment.Identifier,
							},
						},
					},
					Priority: 0,
					RuleID:   uuid.NewString(),
					Serve: struct {
						Distribution *Distribution
						Variation    *string
					}{
						Distribution: nil,
						Variation:    &offVariationIdentifier,
					},
				},
			},
			args: struct {
				target       *Target
				segments     Segments
				defaultServe Serve
			}{
				target:   target,
				segments: Segments{segment.Identifier: segment},
				defaultServe: struct {
					Distribution *Distribution
					Variation    *string
				}{
					Distribution: nil,
					Variation:    &onVariationIdentifier,
				},
			},
			want: offVariationIdentifier,
		},
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

// Creates a slice of ServingRule. This slice will contain one clause, that
// states the 'identifier' attribute must be evaluated using the operator parameter against
// the list of values, and if successful return the varation specified in the variationToServe parameter.
func makeIdentifierRule(values []string, operator, variationToServe string) []ServingRule {
	rule := ServingRule{
		Priority: 0,
		RuleID:   uuid.NewString(),
		Clauses: []Clause{
			{Attribute: "identifier", ID: "id", Negate: false, Op: operator, Value: values},
		},
		Serve: Serve{
			Variation: &variationToServe,
		},
	}
	return []ServingRule{rule}
}

func makeFeatureConfig(name, kind string, variation1, variation2, defaultServe Variation, state FeatureState, rules []ServingRule) FeatureConfig {

	return FeatureConfig{
		DefaultServe: Serve{
			Variation: &defaultServe.Identifier,
		},
		Environment:          "dev",
		Feature:              name,
		Kind:                 kind,
		OffVariation:         variation2.Identifier,
		Rules:                rules,
		Prerequisites:        nil,
		Project:              "default",
		State:                state,
		VariationToTargetMap: nil,
		Variations: []Variation{
			{Description: nil, Identifier: variation1.Identifier, Name: variation1.Name, Value: variation1.Value},
			{Description: nil, Identifier: variation2.Identifier, Name: variation2.Name, Value: variation2.Value},
		},
	}
}

func stringPtr(value string) *string {
	return &value
}

func TestFeatureConfig_Evaluate(t *testing.T) {
	f := false
	harness := "Harness"
	m := make(map[string]interface{})
	m["email"] = "john@doe.com"

	boolFlagName := "SimpleBool"

	onBool := Variation{
		Name:       stringPtr("On"),
		Value:      "true",
		Identifier: "on",
	}

	offBool := Variation{
		Name:       stringPtr("Off"),
		Value:      "false",
		Identifier: "off",
	}

	target := Target{
		Identifier: harness,
		Name:       nil,
		Anonymous:  &f,
		Attributes: &m,
	}

	type args struct {
		target *Target
	}
	tests := []struct {
		name          string
		featureConfig FeatureConfig
		args          args
		want          Evaluation
		wantErr       bool
	}{
		{
			name:          "Test Bool FeatureConfig with no rules serves variation onBool when on",
			featureConfig: makeFeatureConfig(boolFlagName, boolKind, onBool, offBool, onBool, FeatureStateOn, nil),
			args:          args{&target},
			want:          Evaluation{Flag: boolFlagName, Variation: onBool},
			wantErr:       false},
		{
			name:          "Test Bool FeatureConfig Evaluate with no rules serves variation offBool when off",
			featureConfig: makeFeatureConfig(boolFlagName, boolKind, onBool, offBool, onBool, FeatureStateOff, nil),
			args:          args{&target},
			want:          Evaluation{Flag: boolFlagName, Variation: offBool},
			wantErr:       false},
		{
			name:          "Test Bool FeatureConfig Evaluate with 'attribute equals rule' serves offBool on match when flag on",
			featureConfig: makeFeatureConfig(boolFlagName, boolKind, onBool, offBool, onBool, FeatureStateOn, makeIdentifierRule([]string{harness}, equalOperator, offBool.Identifier)),
			args:          args{&target},
			want:          Evaluation{Flag: boolFlagName, Variation: offBool},
			wantErr:       false},
		{
			name:          "Test Bool FeatureConfig Evaluate with 'attribute equals rule' serves onBool on non-match when flag on",
			featureConfig: makeFeatureConfig(boolFlagName, boolKind, onBool, offBool, onBool, FeatureStateOn, makeIdentifierRule([]string{"foobar"}, equalOperator, offBool.Identifier)),
			args:          args{&target},
			want:          Evaluation{Flag: boolFlagName, Variation: onBool},
			wantErr:       false},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			fc := tc.featureConfig
			got, err := fc.Evaluate(tc.args.target)
			if (err != nil) != tc.wantErr {
				t.Errorf("BoolVariation() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			assert.Equal(t, tc.want, got)
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
