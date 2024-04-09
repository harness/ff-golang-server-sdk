package evaluation

import (
	"reflect"
	"testing"

	"github.com/harness/ff-golang-server-sdk/rest"
)

func Test_getAttrValueIsNil(t *testing.T) {
	type args struct {
		target *Target
		attr   string
	}
	tests := []struct {
		name    string
		args    args
		wantStr string
	}{
		{
			name: "when target is nil should return empty string",
			args: args{
				attr: "identifier",
			},
			wantStr: "",
		},
		{
			name: "wrong attribute should return empty string",
			args: args{
				target: &Target{
					Identifier: "harness",
					Attributes: &map[string]interface{}{
						"email": "enver.bisevac@harness.io",
					},
				},
				attr: "no_identifier",
			},
			wantStr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAttrValue(tt.args.target, tt.args.attr); got != tt.wantStr {
				t.Errorf("getAttrValue() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

func Test_getAttrValue(t *testing.T) {
	email := "john@doe.com"
	type args struct {
		target *Target
		attr   string
	}
	tests := []struct {
		name    string
		args    args
		wantStr string
	}{
		{
			name: "check identifier",
			args: args{
				target: &Target{
					Identifier: "harness",
				},
				attr: "identifier",
			},
			wantStr: "harness",
		},
		{
			name: "check name",
			args: args{
				target: &Target{
					Name: "harness",
				},
				attr: "name",
			},
			wantStr: "harness",
		},
		{
			name: "check attributes",
			args: args{
				target: &Target{
					Identifier: "identifier",
					Attributes: &map[string]interface{}{
						"email": email,
					},
				},
				attr: "email",
			},
			wantStr: email,
		},
		{
			name: "check integer attributes",
			args: args{
				target: &Target{
					Identifier: "identifier",
					Attributes: &map[string]interface{}{
						"age": 123,
					},
				},
				attr: "age",
			},
			wantStr: "123",
		},
		{
			name: "check int64 attributes",
			args: args{
				target: &Target{
					Identifier: "identifier",
					Attributes: &map[string]interface{}{
						"age": int64(123),
					},
				},
				attr: "age",
			},
			wantStr: "123",
		},
		{
			name: "check boolean attributes",
			args: args{
				target: &Target{
					Identifier: "identifier",
					Attributes: &map[string]interface{}{
						"active": true,
					},
				},
				attr: "active",
			},
			wantStr: "true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr := getAttrValue(tt.args.target, tt.args.attr)
			if gotStr != tt.wantStr {
				t.Errorf("getAttrValue() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func Test_findVariation(t *testing.T) {
	trueVariation := rest.Variation{
		Identifier: identifierTrue,
		Value:      identifierTrue,
	}
	falseVariation := rest.Variation{

		Identifier: identifierFalse,
		Value:      identifierFalse,
	}
	type args struct {
		variations []rest.Variation
		identifier string
	}
	tests := []struct {
		name    string
		args    args
		want    rest.Variation
		wantErr bool
	}{
		{
			name: "true variation",
			args: args{
				variations: []rest.Variation{trueVariation, falseVariation},
				identifier: identifierTrue,
			},
			want:    trueVariation,
			wantErr: false,
		},
		{
			name: "not found variation",
			args: args{
				variations: []rest.Variation{trueVariation, falseVariation},
				identifier: "hundred",
			},
			want:    rest.Variation{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findVariation(tt.args.variations, tt.args.identifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("findVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findVariation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getNormalizedNumber(t *testing.T) {
	type args struct {
		identifier string
		bucketBy   string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "happy path",
			args: args{
				identifier: "enver.bisevac@harness",
				bucketBy:   "email",
			},
			want: 61,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNormalizedNumber(tt.args.identifier, tt.args.bucketBy); got != tt.want {
				t.Errorf("getNormalizedNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isEnabled(t *testing.T) {
	type args struct {
		target     *Target
		bucketBy   string
		percentage int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "target identifier is empty should return false",
			args: args{
				target: &Target{
					Identifier: "",
				},
				bucketBy:   identifier,
				percentage: 0,
			},
			want: false,
		},
		{
			name: "rollout to 40",
			args: args{
				target: &Target{
					Identifier: "enver",
					Attributes: &map[string]interface{}{
						"email": "enver.bisevac@harness.io",
					},
				},
				bucketBy:   "email",
				percentage: 40,
			},
			want: true,
		},
		{
			name: "rollout from 50",
			args: args{
				target: &Target{
					Identifier: harness,
				},
				bucketBy:   identifier,
				percentage: 50,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEnabled(tt.args.target, tt.args.bucketBy, tt.args.percentage); got != tt.want {
				t.Errorf("isEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_evaluateDistribution(t *testing.T) {
	type args struct {
		distribution *rest.Distribution
		target       *Target
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "distribution is nil",
			args: args{
				distribution: nil,
			},
			want: empty,
		},
		{
			name: "distribution is nil and target is nil",
			args: args{
				distribution: nil,
				target:       nil,
			},
			want: empty,
		},
		{
			name: "serve empty",
			args: args{
				distribution: &rest.Distribution{
					BucketBy: identifier,
					Variations: []rest.WeightedVariation{
						{Variation: identifierTrue, Weight: 1},
						{Variation: identifierFalse, Weight: 2},
					},
				},
				target: &Target{
					Identifier: harness,
				},
			},
			want: identifierFalse,
		},
		{
			name: "serve false",
			args: args{
				distribution: &rest.Distribution{
					BucketBy: identifier,
					Variations: []rest.WeightedVariation{
						{Variation: identifierTrue, Weight: 50},
						{Variation: identifierFalse, Weight: 100},
					},
				},
				target: &Target{
					Identifier: "enver",
				},
			},
			want: identifierFalse,
		},
		{
			name: "bucket value is 67 it should serve B",
			args: args{
				distribution: &rest.Distribution{
					BucketBy: identifier,
					Variations: []rest.WeightedVariation{
						{Variation: "A", Weight: 10},
						{Variation: "B", Weight: 60},
						{Variation: "C", Weight: 30},
					},
				},
				target: &Target{
					Identifier: "enver",
				},
			},
			want: "B",
		},
		{
			name: "bucket value is 32 it should serve A",
			args: args{
				distribution: &rest.Distribution{
					BucketBy: "email",
					Variations: []rest.WeightedVariation{
						{Variation: "A", Weight: 40},
						{Variation: "B", Weight: 40},
						{Variation: "C", Weight: 20},
					},
				},
				target: &Target{
					Identifier: "enver",
					Attributes: &map[string]interface{}{
						"email": "enver.bisevac@harness.io",
					},
				},
			},
			want: "A",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := evaluateDistribution(tt.args.distribution, tt.args.target); got != tt.want {
				t.Errorf("evaluateDistribution() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isTargetInList(t *testing.T) {
	identifier := harness
	type args struct {
		target  *Target
		targets []rest.Target
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "target in a list",
			args: args{
				target: &Target{
					Identifier: identifier,
				},
				targets: []rest.Target{
					{Identifier: identifier},
				},
			},
			want: true,
		},
		{
			name: "target not in a list",
			args: args{
				target: &Target{
					Identifier: identifier,
				},
				targets: []rest.Target{
					{Identifier: "enver"},
				},
			},
			want: false,
		},
		{
			name: "targets is nil should return false",
			args: args{
				target: &Target{
					Identifier: identifier,
				},
				targets: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTargetInList(tt.args.target, tt.args.targets); got != tt.want {
				t.Errorf("isTargetInList() = %v, want %v", got, tt.want)
			}
		})
	}
}
