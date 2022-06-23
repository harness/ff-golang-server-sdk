package evaluation

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/harness/ff-golang-server-sdk/logger"

	"github.com/harness/ff-golang-server-sdk/rest"
)

const (
	identifier        = "identifier"
	harness           = "harness"
	beta              = "beta"
	alpha             = "alpha"
	excluded          = "excluded"
	offVariation      = "false"
	simple            = "simple"
	simpleWithPrereq  = "simplePrereq"
	notValidFlag      = "notValidFlag"
	theme             = "theme"
	size              = "size"
	weight            = "weight"
	org               = "org"
	invalidInt        = "invalidInt"
	invalidNumber     = "invalidNumber"
	invalidJSON       = "invalidJSON"
	prereqNotFound    = "prereqNotFound"
	prereqVarNotFound = "prereqVarNotFound"
)

var (
	empty              = ""
	darktheme          = "darktheme"
	lighttheme         = "lighttheme"
	smallSize          = "50"
	mediumSize         = "100"
	normalWeight       = "50.0"
	heavyWeight        = "100"
	invalidIntValue    = "1a0"
	invalidNumberValue = "1.a0"
	identifierTrue     = "true"
	identifierFalse    = "false"
	targetIdentifier   = "harness"
	json1              = "json1"
	json2              = "json2"
	harness1           = "harness1"
	harness2           = "harness2"
	json1Value         = fmt.Sprintf("{\"org\": \"%s\"}", harness1)
	json2Value         = fmt.Sprintf("{\"org\": \"%s\"}", harness2)
	boolVariations     = []rest.Variation{
		{
			Identifier: identifierTrue,
			Value:      identifierTrue,
		},
		{
			Identifier: identifierFalse,
			Value:      identifierFalse,
		},
	}
	stringVariations = []rest.Variation{
		{
			Identifier: lighttheme,
			Value:      lighttheme,
		},
		{
			Identifier: darktheme,
			Value:      darktheme,
		},
	}
	intVariations = []rest.Variation{
		{
			Identifier: smallSize,
			Value:      smallSize,
		},
		{
			Identifier: mediumSize,
			Value:      mediumSize,
		},
	}
	numberVariations = []rest.Variation{
		{
			Identifier: normalWeight,
			Value:      normalWeight,
		},
		{
			Identifier: heavyWeight,
			Value:      heavyWeight,
		},
	}
	jsonVariations = []rest.Variation{
		{
			Identifier: json1,
			Value:      json1Value,
		},
		{
			Identifier: json2,
			Value:      json2Value,
		},
	}
	testRepo = NewTestRepository(
		map[string]rest.FeatureConfig{
			simple: {
				Feature: simple,
				State:   rest.FeatureStateOn,
				DefaultServe: rest.Serve{
					Variation: &identifierTrue,
				},
				Variations: boolVariations,
				Kind:       "boolean",
			},
			theme: {
				Feature: theme,
				State:   rest.FeatureStateOn,
				DefaultServe: rest.Serve{
					Variation: &lighttheme,
				},
				Variations: stringVariations,
				Kind:       "string",
			},
			size: {
				Feature: size,
				State:   rest.FeatureStateOn,
				DefaultServe: rest.Serve{
					Variation: &mediumSize,
				},
				Variations: intVariations,
				Kind:       "int",
			},
			weight: {
				Feature: weight,
				State:   rest.FeatureStateOn,
				DefaultServe: rest.Serve{
					Variation: &heavyWeight,
				},
				Variations: numberVariations,
				Kind:       "number",
			},
			org: {
				Feature: org,
				State:   rest.FeatureStateOn,
				DefaultServe: rest.Serve{
					Variation: &json2,
				},
				Variations: jsonVariations,
				Kind:       "json",
			},
			invalidInt: {
				Feature: invalidInt,
				State:   rest.FeatureStateOn,
				DefaultServe: rest.Serve{
					Variation: &invalidIntValue,
				},
				Variations: []rest.Variation{
					{
						Identifier: invalidIntValue,
						Value:      invalidIntValue,
					},
				},
				Kind: "int",
			},
			invalidNumber: {
				Feature: invalidNumber,
				State:   rest.FeatureStateOn,
				DefaultServe: rest.Serve{
					Variation: &invalidNumberValue,
				},
				Variations: []rest.Variation{
					{
						Identifier: invalidNumberValue,
						Value:      invalidNumberValue,
					},
				},
				Kind: "number",
			},
			invalidJSON: {
				Feature: invalidJSON,
				State:   rest.FeatureStateOn,
				DefaultServe: rest.Serve{
					Variation: &invalidNumberValue,
				},
				Variations: []rest.Variation{
					{
						Identifier: invalidNumberValue,
						Value:      invalidNumberValue,
					},
				},
				Kind: "json",
			},
			simpleWithPrereq: {
				Feature: simpleWithPrereq,
				State:   rest.FeatureStateOn,
				DefaultServe: rest.Serve{
					Variation: &identifierTrue,
				},
				Variations: boolVariations,
				Prerequisites: &[]rest.Prerequisite{
					{
						Feature:    simple,
						Variations: []string{identifierTrue, identifierFalse},
					},
				},
				Kind: "boolean",
			},
			prereqNotFound: {
				Feature: prereqNotFound,
				State:   rest.FeatureStateOn,
				DefaultServe: rest.Serve{
					Variation: &identifierTrue,
				},
				Variations: boolVariations,
				Prerequisites: &[]rest.Prerequisite{
					{
						Feature:    "prereqNotFound",
						Variations: []string{identifierTrue, identifierFalse},
					},
				},
				Kind: "boolean",
			},
			prereqVarNotFound: {
				Feature:      prereqVarNotFound,
				OffVariation: offVariation,
				State:        rest.FeatureStateOn,
				DefaultServe: rest.Serve{
					Variation: &identifierTrue,
				},
				Variations: boolVariations,
				Prerequisites: &[]rest.Prerequisite{
					{
						Feature:    simple,
						Variations: []string{normalWeight, heavyWeight},
					},
				},
				Kind: "boolean",
			},
			notValidFlag: {
				Feature: notValidFlag,
				State:   rest.FeatureStateOn,
				DefaultServe: rest.Serve{
					Variation: &empty,
				},
				Variations: boolVariations,
				Kind:       "boolean",
			},
		},
		map[string]rest.Segment{
			excluded: {
				Identifier: excluded,
				Excluded: &[]rest.Target{
					{
						Identifier: harness,
					},
				},
			},
			beta: {
				Identifier: beta,
				Included: &[]rest.Target{
					{
						Identifier: harness,
					},
				},
			},
			alpha: {
				Identifier: alpha,
				Rules: &[]rest.Clause{
					{
						Attribute: identifier,
						Op:        equalOperator,
						Values:    []string{harness},
					},
				},
			},
		},
	)
)

type TestRepository struct {
	flags    map[string]rest.FeatureConfig
	segments map[string]rest.Segment
}

func NewTestRepository(flags map[string]rest.FeatureConfig, segments map[string]rest.Segment) TestRepository {
	return TestRepository{
		flags:    flags,
		segments: segments,
	}
}

func (m TestRepository) GetSegment(identifier string) (rest.Segment, error) {
	segment, ok := m.segments[identifier]
	if !ok {
		return rest.Segment{}, fmt.Errorf("segment not found %s", identifier)
	}
	return segment, nil
}

func (m TestRepository) GetFlag(identifier string) (rest.FeatureConfig, error) {
	flag, ok := m.flags[identifier]
	if !ok {
		return rest.FeatureConfig{}, fmt.Errorf("flag not found %s", identifier)
	}
	return flag, nil
}

func TestNewEvaluator(t *testing.T) {
	noOpLogger := logger.NewNoOpLogger()
	eval, _ := NewEvaluator(testRepo, nil, noOpLogger)
	type args struct {
		query  Query
		logger logger.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    *Evaluator
		wantErr bool
	}{
		{
			name:    "nil query should return error",
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "should return test repo",
			args: args{
				query:  testRepo,
				logger: noOpLogger,
			},
			want:    eval,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEvaluator(tt.args.query, nil, noOpLogger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEvaluator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEvaluator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_evaluateClause(t *testing.T) {
	type fields struct {
		query Query
	}
	type args struct {
		clause *rest.Clause
		target *Target
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "clause is nil return false",
			fields: fields{},
			args: args{
				clause: nil,
				target: &Target{
					Identifier: harness,
				},
			},
			want: false,
		},
		{
			name:   "empty operator should return false",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Op:     "",
					Values: []string{harness},
				},
				target: nil,
			},
			want: false,
		},
		{
			name:   "nil values should return false",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Values: nil,
				},
				target: nil,
			},
			want: false,
		},
		{
			name:   "empty values should return false",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Values: []string{},
				},
				target: nil,
			},
			want: false,
		},
		{
			name:   "wrong operator should return false",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Attribute: identifier,
					Op:        "greaterthan",
					Values:    []string{harness},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want: false,
		},
		{
			name:   "empty attribute should return false",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Attribute: empty,
					Op:        equalOperator,
					Values:    []string{harness},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want: false,
		},
		{
			name:   "check match operator",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Attribute: identifier,
					Op:        matchOperator,
					Values:    []string{"^harness$"},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want: true,
		},
		{
			name:   "check match operator (error)",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Attribute: identifier,
					Op:        matchOperator,
					Values:    []string{"^harness(wings$"},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want: false,
		},
		{
			name:   "check in operator",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Attribute: identifier,
					Op:        inOperator,
					Values:    []string{"harness", "wings-software"},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want: true,
		},
		{
			name:   "check in operator (not found) should return false",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Attribute: identifier,
					Op:        inOperator,
					Values:    []string{"harness1", "wings-software"},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want: false,
		},
		{
			name:   "check equal operator",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Attribute: identifier,
					Op:        equalOperator,
					Values:    []string{harness},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want: true,
		},
		{
			name:   "check equal sensitive operator",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Attribute: identifier,
					Op:        equalSensitiveOperator,
					Values:    []string{harness},
				},
				target: &Target{
					Identifier: "Harness",
				},
			},
			want: false,
		},
		{
			name:   "check gt operator",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Attribute: identifier,
					Op:        gtOperator,
					Values:    []string{"A"},
				},
				target: &Target{
					Identifier: "B",
				},
			},
			want: true,
		},
		{
			name:   "check gt operator - negative path",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Attribute: identifier,
					Op:        gtOperator,
					Values:    []string{"B"},
				},
				target: &Target{
					Identifier: "A",
				},
			},
			want: false,
		},
		{
			name:   "check starts with operator",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Attribute: identifier,
					Op:        startsWithOperator,
					Values:    []string{harness},
				},
				target: &Target{
					Identifier: harness + " - wings software",
				},
			},
			want: true,
		},
		{
			name:   "check ends with operator",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Attribute: identifier,
					Op:        endsWithOperator,
					Values:    []string{harness},
				},
				target: &Target{
					Identifier: "wings software - " + harness,
				},
			},
			want: true,
		},
		{
			name:   "check contains operator",
			fields: fields{},
			args: args{
				clause: &rest.Clause{
					Attribute: identifier,
					Op:        containsOperator,
					Values:    []string{harness},
				},
				target: &Target{
					Identifier: "wings " + harness + " software",
				},
			},
			want: true,
		},
		{
			name: "check segments operator",
			fields: fields{
				query: testRepo,
			},
			args: args{
				clause: &rest.Clause{
					Op:     segmentMatchOperator,
					Values: []string{beta},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Evaluator{
				query:  tt.fields.query,
				logger: logger.NewNoOpLogger(),
			}
			if got := e.evaluateClause(tt.args.clause, tt.args.target); got != tt.want {
				t.Errorf("Evaluator.evaluateClause() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_evaluateRules(t *testing.T) {
	type fields struct {
		query Query
	}
	type args struct {
		servingRules []rest.ServingRule
		target       *Target
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "when rules is nil return \"\"",
			args: args{
				servingRules: nil,
			},
			want: "",
		},
		{
			name: "when target is nil return \"\"",
			args: args{
				target: nil,
			},
			want: "",
		},
		{
			name: "evaluate rule",
			args: args{
				// both rule clauses are true so it will serve false and true
				// priority is on second one and should return true
				servingRules: []rest.ServingRule{
					{
						Priority: 2,
						Clauses: []rest.Clause{
							{
								Attribute: identifier,
								Op:        equalOperator,
								Values:    []string{harness},
							},
						},
						Serve: rest.Serve{
							Variation: &identifierFalse,
						},
					},
					{
						Priority: 1,
						Clauses: []rest.Clause{
							{
								Attribute: identifier,
								Op:        equalOperator,
								Values:    []string{harness},
							},
						},
						Serve: rest.Serve{
							Variation: &identifierTrue,
						},
					},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want: identifierTrue,
		},
		{
			name: "evaluate rule continue in for loop should occur",
			args: args{
				// both rule clauses are true so it will serve false and true
				// priority is on second one and should return true
				servingRules: []rest.ServingRule{
					{
						Priority: 1,
						Clauses: []rest.Clause{
							{
								Attribute: identifier,
								Op:        equalOperator,
								Values:    []string{"harnesss"},
							},
						},
						Serve: rest.Serve{
							Variation: &identifierTrue,
						},
					},
					{
						Priority: 2,
						Clauses: []rest.Clause{
							{
								Attribute: identifier,
								Op:        equalOperator,
								Values:    []string{harness},
							},
						},
						Serve: rest.Serve{
							Variation: &identifierTrue,
						},
					},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want: identifierTrue,
		},
		{
			name: "evaluate rule by distribution",
			args: args{
				servingRules: []rest.ServingRule{
					{
						Priority: 1,
						Clauses: []rest.Clause{
							{
								Attribute: identifier,
								Op:        equalOperator,
								Values:    []string{harness},
							},
						},
						Serve: rest.Serve{
							Distribution: &rest.Distribution{
								BucketBy: identifier,
								Variations: []rest.WeightedVariation{
									{Variation: identifierTrue, Weight: 5},
									{Variation: identifierFalse, Weight: 95},
								},
							},
						},
					},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want: identifierFalse,
		},
		{
			name: "evaluate rule (target is nil) return variation identifier empty",
			args: args{
				servingRules: []rest.ServingRule{
					{
						Priority: 1,
						Clauses: []rest.Clause{
							{
								Attribute: identifier,
								Op:        equalOperator,
								Values:    []string{harness},
							},
						},
						Serve: rest.Serve{
							Variation: &identifierFalse,
						},
					},
				},
				target: nil,
			},
			want: "",
		},
		{
			name: "when rules is empty return \"\"",
			args: args{
				servingRules: []rest.ServingRule{},
				target: &Target{
					Identifier: harness,
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Evaluator{
				query:  tt.fields.query,
				logger: logger.NewNoOpLogger(),
			}
			if got := e.evaluateRules(tt.args.servingRules, tt.args.target); got != tt.want {
				t.Errorf("Evaluator.evaluateRules() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_evaluateVariationMap(t *testing.T) {
	type fields struct {
		query Query
	}
	type args struct {
		variationsMap []rest.VariationMap
		target        *Target
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "when variations map is nil return \"\"",
			args: args{
				variationsMap: nil,
			},
			want: "",
		},
		{
			name: "when target is nil return \"\"",
			args: args{
				target: nil,
			},
			want: "",
		},
		{
			name: "when target identifier in targets serve true",
			args: args{
				variationsMap: []rest.VariationMap{
					{
						Variation: identifierTrue,
						Targets: &[]rest.TargetMap{
							{
								Identifier: &targetIdentifier,
							},
						},
					},
				},
				target: &Target{
					Identifier: targetIdentifier,
				},
			},
			want: identifierTrue,
		},
		{
			name: "when all targets in all variation maps is nil then serve \"\"",
			fields: fields{
				query: testRepo,
			},
			args: args{
				variationsMap: []rest.VariationMap{
					{
						Variation:      identifierTrue,
						TargetSegments: &[]string{beta},
					},
				},
				target: &Target{
					Identifier: targetIdentifier,
				},
			},
			want: identifierTrue,
		},
		{
			name: "when all targets and segments in all variation maps is nil then serve \"\"",
			fields: fields{
				query: testRepo,
			},
			args: args{
				target: &Target{
					Identifier: targetIdentifier,
				},
			},
			want: "",
		},
		{
			name: "target identifier in segments serve true",
			fields: fields{
				query: testRepo,
			},
			args: args{
				variationsMap: []rest.VariationMap{
					{
						Variation:      identifierTrue,
						TargetSegments: &[]string{beta},
					},
				},
				target: &Target{
					Identifier: targetIdentifier,
				},
			},
			want: identifierTrue,
		},
		{
			name: "when variations map is empty return \"\"",
			args: args{
				variationsMap: []rest.VariationMap{},
				target: &Target{
					Identifier: targetIdentifier,
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Evaluator{
				query:  tt.fields.query,
				logger: logger.NewNoOpLogger(),
			}
			if got := e.evaluateVariationMap(tt.args.variationsMap, tt.args.target); got != tt.want {
				t.Errorf("Evaluator.evaluateVariationMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_evaluateFlag(t *testing.T) {
	type fields struct {
		query Query
	}
	type args struct {
		fc     rest.FeatureConfig
		target *Target
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    rest.Variation
		wantErr bool
	}{
		{
			name: "evaluation of flag when is off state serve off variation",
			args: args{
				fc: rest.FeatureConfig{
					OffVariation: offVariation,
					State:        rest.FeatureStateOff,
					Variations:   boolVariations,
				},
			},
			want:    boolVariations[1],
			wantErr: false,
		},
		{
			name: "evaluation with target when flag is off serve off variation",
			args: args{
				fc: rest.FeatureConfig{
					OffVariation: offVariation,
					State:        rest.FeatureStateOff,
					Variations:   boolVariations,
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want:    boolVariations[1],
			wantErr: false,
		},
		{
			name: "evaluate flag should return default serve variation",
			args: args{
				fc: rest.FeatureConfig{
					State:      rest.FeatureStateOn,
					Variations: boolVariations,
					DefaultServe: rest.Serve{
						Variation: &boolVariations[0].Value,
					},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want:    boolVariations[0],
			wantErr: false,
		},
		{
			name: "evaluate flag should return default serve distribution",
			args: args{
				fc: rest.FeatureConfig{
					State:      rest.FeatureStateOn,
					Variations: boolVariations,
					DefaultServe: rest.Serve{
						Distribution: &rest.Distribution{
							Variations: []rest.WeightedVariation{
								{
									Variation: identifierTrue,
									Weight:    5,
								},
								{
									Variation: identifierFalse,
									Weight:    95,
								},
							},
						},
					},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want:    boolVariations[1],
			wantErr: false,
		},
		{
			name: "evaluate flag should return rule serve",
			args: args{
				fc: rest.FeatureConfig{
					State:      rest.FeatureStateOn,
					Variations: boolVariations,
					Rules: &[]rest.ServingRule{
						{
							Clauses: []rest.Clause{
								{
									Attribute: identifier,
									Op:        equalOperator,
									Values:    []string{harness},
								},
							},
							Serve: rest.Serve{
								Variation: &boolVariations[0].Value,
							},
						},
					},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want:    boolVariations[0],
			wantErr: false,
		},
		{
			name: "evaluate flag using variationMap and target should return 'true'",
			args: args{
				fc: rest.FeatureConfig{
					State:      rest.FeatureStateOn,
					Variations: boolVariations,
					VariationToTargetMap: &[]rest.VariationMap{
						{
							Variation: identifierTrue,
							Targets: &[]rest.TargetMap{
								{
									Identifier: &targetIdentifier,
								},
							},
						},
					},
				},
				target: &Target{
					Identifier: targetIdentifier,
				},
			},
			want:    boolVariations[0],
			wantErr: false,
		},
		{
			name: "evaluate flag variation returns an error",
			args: args{
				fc: rest.FeatureConfig{
					State: rest.FeatureStateOn,
				},
				target: &Target{
					Identifier: targetIdentifier,
				},
			},
			want:    rest.Variation{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Evaluator{
				query:  tt.fields.query,
				logger: logger.NewNoOpLogger(),
			}
			got, err := e.evaluateFlag(tt.args.fc, tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluator.evaluateFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Evaluator.evaluateFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_isTargetIncludedOrExcludedInSegment(t *testing.T) {
	type fields struct {
		query Query
	}
	type args struct {
		segmentList []string
		target      *Target
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "segment list is empty return false",
			args: args{
				segmentList: nil,
			},
			want: false,
		},
		{
			name: "segment not found should return false",
			fields: fields{
				query: testRepo,
			},
			args: args{
				segmentList: []string{"segmentNotFound1000"},
			},
			want: false,
		},
		{
			name: "segment in excluded should return false",
			fields: fields{
				query: testRepo,
			},
			args: args{
				segmentList: []string{excluded},
				target: &Target{
					Identifier: harness,
				},
			},
			want: false,
		},
		{
			name: "segment with target identifier should return true",
			fields: fields{
				query: testRepo,
			},
			args: args{
				segmentList: []string{beta},
				target: &Target{
					Identifier: harness,
				},
			},
			want: true,
		},
		{
			name: "evaluate rule in segment rules should return true",
			fields: fields{
				query: testRepo,
			},
			args: args{
				segmentList: []string{alpha},
				target: &Target{
					Identifier: harness,
				},
			},
			want: true,
		},
		{
			name: "segment rule clause with false result should return false",
			fields: fields{
				query: testRepo,
			},
			args: args{
				segmentList: []string{alpha},
				target: &Target{
					Identifier: "no_identifier",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Evaluator{
				query:  tt.fields.query,
				logger: logger.NewNoOpLogger(),
			}
			if got := e.isTargetIncludedOrExcludedInSegment(tt.args.segmentList, tt.args.target); got != tt.want {
				t.Errorf("Evaluator.isTargetIncludedOrExcludedInSegment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_checkPreRequisite(t *testing.T) {
	type fields struct {
		query Query
	}
	type args struct {
		parent *rest.FeatureConfig
		target *Target
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "data provider missing, returns error",
			args: args{
				parent: &rest.FeatureConfig{},
			},
			want:    true,
			wantErr: true,
		},
		{
			name: "no prerequities should return true",
			fields: fields{
				query: testRepo,
			},
			args: args{
				parent: &rest.FeatureConfig{},
			},
			want: true,
		},
		{
			name: "prereq simple should return true",
			fields: fields{
				query: testRepo,
			},
			args: args{
				parent: &rest.FeatureConfig{
					State: rest.FeatureStateOn,
					Prerequisites: &[]rest.Prerequisite{
						{
							Feature:    simple,
							Variations: []string{identifierTrue, identifierFalse},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "prereq flag doesn't exists it should return false",
			fields: fields{
				query: testRepo,
			},
			args: args{
				parent: &rest.FeatureConfig{
					State: rest.FeatureStateOn,
					Prerequisites: &[]rest.Prerequisite{
						{
							Feature:    "prereq not found",
							Variations: []string{identifierTrue, identifierFalse},
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Evaluator{
				query:  tt.fields.query,
				logger: logger.NewNoOpLogger(),
			}
			got, err := e.checkPreRequisite(tt.args.parent, tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluator.checkPreRequisite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Evaluator.checkPreRequisite() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_evaluate(t *testing.T) {
	type fields struct {
		query Query
	}
	type args struct {
		identifier string
		target     *Target
		kind       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    rest.Variation
		wantErr bool
	}{
		{
			name:   "data provider missing return error",
			fields: fields{},
			args: args{
				identifier: simple,
				target: &Target{
					Identifier: harness,
				},
				kind: "boolean",
			},
			want:    rest.Variation{},
			wantErr: true,
		},
		{
			name: "flag doesn't exist",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier: "some_test_flag",
				kind:       "boolean",
			},
			want:    rest.Variation{},
			wantErr: true,
		},
		{
			name: "flag kind mismatch",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier: simple,
				kind:       "string",
			},
			want:    rest.Variation{},
			wantErr: true,
		},
		{
			name: "prereq flag simple should return true",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier: simpleWithPrereq,
				kind:       "boolean",
			},
			want: boolVariations[0],
		},
		{
			name: "error evaluating flag",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier: notValidFlag,
				kind:       "boolean",
			},
			want:    rest.Variation{},
			wantErr: true,
		},
		{
			name: "error evaluating prereq",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier: prereqVarNotFound,
				kind:       "boolean",
			},
			want:    boolVariations[1], // returns off variation
			wantErr: false,
		},
		{
			name: "happy path",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier: simple,
				kind:       "boolean",
			},
			want: boolVariations[0],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Evaluator{
				query:  tt.fields.query,
				logger: logger.NewNoOpLogger(),
			}
			got, err := e.evaluate(tt.args.identifier, tt.args.target, tt.args.kind)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluator.evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Evaluator.evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_BoolVariation(t *testing.T) {
	type fields struct {
		query Query
	}
	type args struct {
		identifier   string
		target       *Target
		defaultValue bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "bool flag not found return default value",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier:   "flagNotFound1000",
				target:       nil,
				defaultValue: false,
			},
			want: false,
		},
		{
			name: "bool evaluation of flag 'simple' should return true",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier:   simple,
				target:       nil,
				defaultValue: false,
			},
			want: true,
		},
		{
			name: "bool evaluation of flag 'simple' with target 'harness' should return true",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier: simple,
				target: &Target{
					Identifier: harness,
				},
				defaultValue: false,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Evaluator{
				query:  tt.fields.query,
				logger: logger.NewNoOpLogger(),
			}
			if got := e.BoolVariation(tt.args.identifier, tt.args.target, tt.args.defaultValue); got != tt.want {
				t.Errorf("Evaluator.BoolVariation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_StringVariation(t *testing.T) {
	type fields struct {
		query Query
	}
	type args struct {
		identifier   string
		target       *Target
		defaultValue string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "string flag not found return default value",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier:   "flagNotFound1000",
				target:       nil,
				defaultValue: darktheme,
			},
			want: darktheme,
		},
		{
			name: "string evaluation of flag 'theme' should return lightheme",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier:   theme,
				target:       nil,
				defaultValue: darktheme,
			},
			want: lighttheme,
		},
		{
			name: "string evaluation of flag 'theme' with target 'harness' should return lighttheme",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier: theme,
				target: &Target{
					Identifier: harness,
				},
				defaultValue: darktheme,
			},
			want: lighttheme,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Evaluator{
				query:  tt.fields.query,
				logger: logger.NewNoOpLogger(),
			}
			if got := e.StringVariation(tt.args.identifier, tt.args.target, tt.args.defaultValue); got != tt.want {
				t.Errorf("Evaluator.StringVariation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_IntVariation(t *testing.T) {
	type fields struct {
		query Query
	}
	type args struct {
		identifier   string
		target       *Target
		defaultValue int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "int flag not found return default value",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier:   "flagNotFound1000",
				target:       nil,
				defaultValue: 50,
			},
			want: 50,
		},
		{
			name: "int evaluation of flag 'size' should return medium",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier:   size,
				target:       nil,
				defaultValue: 50,
			},
			want: 100,
		},
		{
			name: "not valid int evaluation of flag 'size' should return default value",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier:   invalidInt,
				target:       nil,
				defaultValue: 50,
			},
			want: 50,
		},
		{
			name: "int evaluation of flag 'size' with target 'harness' should return medium",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier: size,
				target: &Target{
					Identifier: harness,
				},
				defaultValue: 50,
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Evaluator{
				query:  tt.fields.query,
				logger: logger.NewNoOpLogger(),
			}
			if got := e.IntVariation(tt.args.identifier, tt.args.target, tt.args.defaultValue); got != tt.want {
				t.Errorf("Evaluator.IntVariation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_NumberVariation(t *testing.T) {
	type fields struct {
		query Query
	}
	type args struct {
		identifier   string
		target       *Target
		defaultValue float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{
			name: "number flag not found return default value",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier:   "flagNotFound1000",
				target:       nil,
				defaultValue: 50.0,
			},
			want: 50.0,
		},
		{
			name: "number evaluation of flag 'weight' should return heavyWeight",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier:   weight,
				target:       nil,
				defaultValue: 50.0,
			},
			want: 100.0,
		},
		{
			name: "number evaluation of flag 'weight' should return default value",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier:   invalidNumber,
				target:       nil,
				defaultValue: 50.0,
			},
			want: 50.0,
		},
		{
			name: "number evaluation of flag 'weight' with target 'harness' should return heavyWeight",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier: weight,
				target: &Target{
					Identifier: harness,
				},
				defaultValue: 50.0,
			},
			want: 100.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Evaluator{
				query:  tt.fields.query,
				logger: logger.NewNoOpLogger(),
			}
			if got := e.NumberVariation(tt.args.identifier, tt.args.target, tt.args.defaultValue); got != tt.want {
				t.Errorf("Evaluator.NumberVariation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_JSONVariation(t *testing.T) {
	defaultValue := map[string]interface{}{
		"email": "harness@harness.io",
	}
	type fields struct {
		query Query
	}
	type args struct {
		identifier   string
		target       *Target
		defaultValue map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]interface{}
	}{
		{
			name: "json flag not found return default value",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier:   "flagNotFound1000",
				target:       nil,
				defaultValue: defaultValue,
			},
			want: defaultValue,
		},
		{
			name: "json evaluation of flag 'org' should return json2Value",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier:   org,
				target:       nil,
				defaultValue: defaultValue,
			},
			want: map[string]interface{}{
				org: harness2,
			},
		},
		{
			name: "json evaluation of flag 'org' should return default value",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier:   invalidJSON,
				target:       nil,
				defaultValue: defaultValue,
			},
			want: defaultValue,
		},
		{
			name: "json evaluation of flag 'org' with target 'harness' should return json2",
			fields: fields{
				query: testRepo,
			},
			args: args{
				identifier: org,
				target: &Target{
					Identifier: harness,
				},
				defaultValue: defaultValue,
			},
			want: map[string]interface{}{
				org: harness2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Evaluator{
				query:  tt.fields.query,
				logger: logger.NewNoOpLogger(),
			}
			if got := e.JSONVariation(tt.args.identifier, tt.args.target, tt.args.defaultValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Evaluator.JSONVariation() = %v, want %v", got, tt.want)
			}
		})
	}
}
