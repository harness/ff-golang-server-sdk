package evaluation

import (
	"encoding/json"
	"github.com/wings-software/ff-client-sdk-go/types"
	"reflect"
	"strconv"
)

const (
	segmentMatchOperator   = "segmentMatch"
	inOperator             = "in"
	equalOperator          = "equal"
	gtOperator             = "gt"
	startsWithOperator     = "starts_with"
	endsWithOperator       = "ends_with"
	containsOperator       = "contains"
	equalSensitiveOperator = "equal_sensitive"
)

type Evaluation struct {
	Flag  string
	Value interface{}
}

type Clause struct {
	Attribute string
	Id        string
	Negate    bool
	Op        string
	Value     []string
}

func (c *Clause) Evaluate(target *Target, segments Segments, operator types.ValueType) bool {
	switch c.Op {
	case segmentMatchOperator:
		return c.segmentMatch(target, segments)
	case inOperator:
		return operator.In(c.Value)
	case equalOperator:
		return operator.Equal(c.Value)
	case gtOperator:
		return operator.GreaterThan(c.Value)
	case startsWithOperator:
		return operator.StartsWith(c.Value)
	case endsWithOperator:
		return operator.EndWith(c.Value)
	case containsOperator:
		return operator.Contains(c.Value)
	case equalSensitiveOperator:
		return operator.EqualSensitive(c.Value)
	}
	return false
}

// segmentMatch should expect array of string (targetSegment key)
// segment match should process Segment rules
func (c *Clause) segmentMatch(target *Target, segments Segments) bool {
	if c.Op != segmentMatchOperator || segments == nil {
		return false // should we return error ?
	}

	return segments.Evaluate(target)
}

type Clauses []Clause

func (c Clauses) Evaluate(target *Target, segments Segments) bool {
	for _, clause := range c {
		// AND operation
		op := target.GetOperator(clause.Attribute)
		if op == nil || !clause.Evaluate(target, segments, op) {
			return false
		}
		// continue on next clause
	}
	// it means that all clauses passed
	return true
}

type Distribution struct {
	BucketBy   string
	Variations []WeightedVariation
}

func (d *Distribution) GetKeyName(target *Target) string {
	variation := ""
	for _, tdVariation := range d.Variations {
		variation = tdVariation.Variation
		if d.isEnabled(target, tdVariation.Weight) {
			return variation
		}
	}
	// distance between last variation and total percentage
	if d.isEnabled(target, OneHundred) {
		return variation
	}
	return ""
}

func (d *Distribution) isEnabled(target *Target, percentage int) bool {
	value := target.GetAttrValue(d.BucketBy)
	identifier := value.String()
	if identifier == "" {
		return false
	}

	bucketId := GetNormalizedNumber(identifier, d.BucketBy)
	return percentage > 0 && bucketId <= percentage
}

type FeatureConfig struct {
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
	Segments             map[string]*Segment `json:"-"`
}

func (fc FeatureConfig) GetSegmentIdentifiers() StrSlice {
	slice := make(StrSlice, 0)
	for _, rule := range fc.Rules {
		for _, clause := range rule.Clauses {
			if clause.Op == segmentMatchOperator {
				for _, val := range clause.Value {
					slice = append(slice, val)
				}
			}
		}
	}
	return slice
}

func (fc *FeatureConfig) GetKind() reflect.Kind {
	switch fc.Kind {
	case "boolean":
		return reflect.Bool
	case "string":
		return reflect.String
	case "int", "integer":
		return reflect.Int
	case "number":
		return reflect.Float64
	case "json":
		return reflect.Map
	default:
		return reflect.Invalid
	}
}

func (fc FeatureConfig) Evaluate(target *Target) *Evaluation {
	var value interface{}
	switch fc.GetKind() {
	case reflect.Bool:
		value = fc.BoolVariation(target, false) // need more info
	case reflect.String:
		value = fc.StringVariation(target, "") // need more info
	case reflect.Int:
		value = fc.IntVariation(target, 0)
	case reflect.Float64:
		value = fc.NumberVariation(target, 0)
	case reflect.Map:
		value = fc.JsonVariation(target, map[string]interface{}{})
	}
	return &Evaluation{
		Flag:  fc.Feature,
		Value: value,
	}
}

func (fc FeatureConfig) GetVariationName(target *Target) string {
	if fc.State == FeatureStateOff {
		return fc.OffVariation
	} else {
		// TODO: variation to target
		if fc.VariationToTargetMap != nil && len(fc.VariationToTargetMap) > 0 {
			for _, variationMap := range fc.VariationToTargetMap {
				if variationMap.Targets != nil {
					for _, t := range variationMap.Targets {
						if target.Identifier == t {
							return variationMap.Variation
						}
					}
				}
			}
		}

		return fc.Rules.GetVariationName(target, fc.Segments, fc.DefaultServe)
	}
}

func (fc *FeatureConfig) BoolVariation(target *Target, defaultValue bool) bool {
	if fc.GetKind() != reflect.Bool {
		return defaultValue
	}
	return fc.Variations.FindByIdentifier(fc.GetVariationName(target)).Bool(defaultValue)
}

func (fc *FeatureConfig) StringVariation(target *Target, defaultValue string) string {
	if fc.GetKind() != reflect.String {
		return defaultValue
	}
	return fc.Variations.FindByIdentifier(fc.GetVariationName(target)).String(defaultValue)
}

func (fc *FeatureConfig) IntVariation(target *Target, defaultValue int64) int64 {
	if fc.GetKind() != reflect.Int {
		return defaultValue
	}
	return fc.Variations.FindByIdentifier(fc.GetVariationName(target)).Int(defaultValue)
}

func (fc *FeatureConfig) NumberVariation(target *Target, defaultValue float64) float64 {
	if fc.GetKind() != reflect.Float64 {
		return defaultValue
	}
	return fc.Variations.FindByIdentifier(fc.GetVariationName(target)).Number(defaultValue)
}

func (fc *FeatureConfig) JsonVariation(target *Target, defaultValue types.JSON) types.JSON {
	if fc.GetKind() != reflect.Map {
		return defaultValue
	}
	return fc.Variations.FindByIdentifier(fc.GetVariationName(target)).JSON(defaultValue)
}

type FeatureState string

const (
	FeatureStateOff FeatureState = "off"
	FeatureStateOn  FeatureState = "on"
)

type Prerequisite struct {
	Feature    string
	Variations []string
}

type Serve struct {
	Distribution *Distribution
	Variation    *string
}

type ServingRule struct {
	Clauses  Clauses
	Priority int
	RuleId   string
	Serve    Serve
}

type ServingRules []ServingRule

func (sr ServingRules) GetVariationName(target *Target, segments Segments, defaultServe Serve) string {
RULES:
	for _, rule := range sr {
		// rules are OR operation
		if !rule.Clauses.Evaluate(target, segments) {
			continue RULES
		}
		var name string
		if rule.Serve.Variation != nil {
			name = *rule.Serve.Variation
		} else {
			name = rule.Serve.Distribution.GetKeyName(target)
		}
		return name
	}

	// not found then return defaultServe
	if defaultServe.Variation != nil {
		return *defaultServe.Variation
	}

	if defaultServe.Distribution != nil {
		defaultServe.Distribution.isEnabled(target, OneHundred)
	}
	return "" // need defaultServe
}

type Tag struct {
	Name  string
	Value *string
}

type Variation struct {
	Description *string
	Identifier  string
	Name        *string
	Value       string
}

func (v *Variation) Bool(defaultValue bool) bool {
	if v == nil {
		return defaultValue
	}
	boolValue, _ := strconv.ParseBool(v.Value)
	return boolValue
}

func (v *Variation) String(defaultValue string) string {
	if v == nil {
		return defaultValue
	}
	return v.Value
}

func (v *Variation) Number(defaultValue float64) float64 {
	if v == nil {
		return defaultValue
	}
	number, _ := strconv.ParseFloat(v.Value, 64)
	return number
}

func (v *Variation) Int(defaultValue int64) int64 {
	if v == nil {
		return defaultValue
	}
	intVal, _ := strconv.ParseInt(v.Value, 10, 64)
	return intVal
}

func (v *Variation) JSON(defaultValue types.JSON) types.JSON {
	if v == nil {
		return defaultValue
	}
	result := make(types.JSON)
	if err := json.Unmarshal([]byte(v.Value), &result); err != nil {
		return defaultValue
	}
	return result
}

type Variations []Variation

func (v Variations) FindByIdentifier(identifier string) *Variation {
	for _, val := range v {
		if val.Identifier == identifier {
			return &val
		}
	}
	return nil
}

type VariationMap struct {
	TargetSegments []string
	Targets        []string
	Variation      string
}

type WeightedVariation struct {
	Variation string
	Weight    int
}
