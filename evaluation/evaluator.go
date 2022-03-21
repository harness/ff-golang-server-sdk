package evaluation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/harness/ff-golang-server-sdk/log"
	"github.com/harness/ff-golang-server-sdk/rest"
)

const (
	oneHundred = 100

	segmentMatchOperator   = "segmentMatch"
	matchOperator          = "match"
	inOperator             = "in"
	equalOperator          = "equal"
	gtOperator             = "gt"
	startsWithOperator     = "starts_with"
	endsWithOperator       = "ends_with"
	containsOperator       = "contains"
	equalSensitiveOperator = "equal_sensitive"
)

type ProcessEvaluation func(fc *rest.FeatureConfig, target *Target, variation *rest.Variation)

// Query provides methods for segment and flag retrieval
type Query interface {
	GetSegment(identifier string) (rest.Segment, error)
	GetFlag(identifier string) (rest.FeatureConfig, error)
}

// Evaluator engine evaluates flag from provided query
type Evaluator struct {
	query Query
}

// NewEvaluator constructs evaluator with query instance
func NewEvaluator(query Query) (*Evaluator, error) {
	if query == nil {
		return nil, ErrQueryProviderMissing
	}
	return &Evaluator{
		query: query,
	}, nil
}

func (e Evaluator) evaluateClause(clause *rest.Clause, target *Target) bool {
	if clause == nil {
		return false
	}

	values := clause.Values
	if len(values) == 0 {
		return false
	}
	value := values[0]

	operator := clause.Op
	if operator == "" {
		return false
	}

	attrValue := getAttrValue(target, clause.Attribute)
	if operator != segmentMatchOperator && !attrValue.IsValid() {
		return false
	}

	object := attrValue.String()

	switch operator {
	case startsWithOperator:
		return strings.HasPrefix(object, value)
	case endsWithOperator:
		return strings.HasSuffix(object, value)
	case matchOperator:
		found, err := regexp.MatchString(value, object)
		if err != nil || !found {
			return false
		}
		return true
	case containsOperator:
		return strings.Contains(object, value)
	case equalOperator:
		return strings.EqualFold(object, value)
	case equalSensitiveOperator:
		return object == value
	case inOperator:
		for _, val := range values {
			if val == object {
				return true
			}
		}
		return false
	case gtOperator:
		return object > value
	case segmentMatchOperator:
		return e.isTargetIncludedOrExcludedInSegment(values, target)
	default:
		return false
	}
}

func (e Evaluator) evaluateClauses(clauses []rest.Clause, target *Target) bool {
	for i := range clauses {
		if !e.evaluateClause(&clauses[i], target) {
			return false
		}
	}
	return true
}

func (e Evaluator) evaluateRule(servingRule *rest.ServingRule, target *Target) bool {
	return e.evaluateClauses(servingRule.Clauses, target)
}

func (e Evaluator) evaluateRules(servingRules []rest.ServingRule, target *Target) string {
	if target == nil || servingRules == nil {
		return ""
	}

	sort.SliceStable(servingRules, func(i, j int) bool {
		return servingRules[i].Priority < servingRules[j].Priority
	})
	for i := range servingRules {
		rule := servingRules[i]
		// if evaluation is false just continue to next rule
		if !e.evaluateRule(&rule, target) {
			continue
		}

		// rule matched, check if there is distribution
		if rule.Serve.Distribution != nil {
			return evaluateDistribution(rule.Serve.Distribution, target)
		}

		// rule matched, here must be variation if distribution is undefined or null
		if rule.Serve.Variation != nil {
			return *rule.Serve.Variation
		}
	}
	return ""
}

func (e Evaluator) evaluateVariationMap(variationsMap []rest.VariationMap, target *Target) string {
	if variationsMap == nil || target == nil {
		return ""
	}

	for _, variationMap := range variationsMap {
		if variationMap.Targets != nil {
			for _, t := range *variationMap.Targets {
				if *t.Identifier != "" && *t.Identifier == target.Identifier {
					return variationMap.Variation
				}
			}
		}

		segmentIdentifiers := variationMap.TargetSegments
		if segmentIdentifiers != nil && e.isTargetIncludedOrExcludedInSegment(*segmentIdentifiers, target) {
			return variationMap.Variation
		}
	}
	return ""
}

func (e Evaluator) evaluateFlag(fc rest.FeatureConfig, target *Target) (rest.Variation, error) {
	var variation = fc.OffVariation
	if fc.State == rest.FeatureState_on {
		variation = ""
		if fc.VariationToTargetMap != nil {
			variation = e.evaluateVariationMap(*fc.VariationToTargetMap, target)
		}
		if variation == "" && fc.Rules != nil {
			variation = e.evaluateRules(*fc.Rules, target)
		}
		if variation == "" {
			variation = evaluateDistribution(fc.DefaultServe.Distribution, target)
		}
		if variation == "" && fc.DefaultServe.Variation != nil {
			variation = *fc.DefaultServe.Variation
		}
	}

	if variation != "" {
		return findVariation(fc.Variations, variation)
	}
	return rest.Variation{}, fmt.Errorf("%w: %s", ErrEvaluationFlag, fc.Feature)
}

func (e Evaluator) isTargetIncludedOrExcludedInSegment(segmentList []string, target *Target) bool {
	if segmentList == nil {
		return false
	}
	for _, segmentIdentifier := range segmentList {
		segment, err := e.query.GetSegment(segmentIdentifier)
		if err != nil {
			return false
		}
		// Should Target be excluded - if in excluded list we return false
		if segment.Excluded != nil && isTargetInList(target, *segment.Excluded) {
			log.Debugf("Target %s excluded from segment %s via exclude list", target.Name, segment.Name)
			return false
		}

		// Should Target be included - if in included list we return true
		if segment.Included != nil && isTargetInList(target, *segment.Included) {
			log.Debugf(
				"Target %s included in segment %s via include list",
				target.Name,
				segment.Name)
			return true
		}

		// Should Target be included via segment rules
		rules := segment.Rules
		if rules != nil && e.evaluateClauses(*rules, target) {
			log.Debugf(
				"Target %s included in segment %s via rules", target.Name, segment.Name)
			return true
		}
	}
	return false
}

func (e Evaluator) checkPreRequisite(parent *rest.FeatureConfig, target *Target) (bool, error) {
	if e.query == nil {
		log.Errorf(ErrQueryProviderMissing.Error())
		return true, ErrQueryProviderMissing
	}
	prerequisites := parent.Prerequisites
	if prerequisites != nil {
		log.Infof(
			"Checking pre requisites %v of parent feature %v",
			prerequisites,
			parent.Feature)
		for _, pre := range *prerequisites {
			prereqFeature := pre.Feature
			prereqFeatureConfig, err := e.query.GetFlag(prereqFeature)
			if err != nil {
				log.Errorf(
					"Could not retrieve the pre requisite details of feature flag : %v", prereqFeature)
				return true, nil
			}

			prereqEvaluatedVariation, err := e.evaluateFlag(prereqFeatureConfig, target)
			if err != nil {
				log.Errorf(
					"Could not evaluate the prerequisite details of feature flag : %v", prereqFeature)
				return true, nil
			}

			log.Infof(
				"Pre requisite flag %v has variation %v for target %v",
				prereqFeatureConfig.Feature,
				prereqEvaluatedVariation,
				target)

			// Compare if the pre requisite variation is a possible valid value of
			// the pre requisite FF
			validPrereqVariations := pre.Variations
			log.Infof(
				"Pre requisite flag %v should have the variations %v",
				prereqFeatureConfig.Feature,
				validPrereqVariations)
			if !contains(validPrereqVariations, prereqEvaluatedVariation.Identifier) {
				return false, nil
			}
			// no need to check this error because it is recursion so err should never happen
			if r, _ := e.checkPreRequisite(&prereqFeatureConfig, target); !r {
				return false, nil
			}
		}
	}
	return true, nil
}

func (e Evaluator) evaluate(identifier string, target *Target, kind string,
	processEvaluation ProcessEvaluation) (rest.Variation, error) {

	if e.query == nil {
		log.Errorf(ErrQueryProviderMissing.Error())
		return rest.Variation{}, ErrQueryProviderMissing
	}
	flag, err := e.query.GetFlag(identifier)
	if err != nil {
		return rest.Variation{}, err
	}

	if flag.Kind != kind {
		return rest.Variation{}, fmt.Errorf("%w, expected: %s, got: %s", ErrFlagKindMismatch, kind, flag.Kind)
	}

	if flag.Prerequisites != nil {
		prereq, err := e.checkPreRequisite(&flag, target)
		if err != nil || !prereq {
			return findVariation(flag.Variations, flag.OffVariation)
		}
	}
	variation, err := e.evaluateFlag(flag, target)
	if err != nil {
		return rest.Variation{}, err
	}
	if processEvaluation != nil {
		processEvaluation(&flag, target, &variation)
	}
	return variation, nil
}

// BoolVariation returns boolean evaluation for target
func (e Evaluator) BoolVariation(identifier string, target *Target, defaultValue bool, processEvaluation ProcessEvaluation) bool {
	variation, err := e.evaluate(identifier, target, "boolean", processEvaluation)
	if err != nil {
		log.Errorf("Error while evaluating boolean flag '%s', err: %v", identifier, err)
		return defaultValue
	}
	return strings.ToLower(variation.Value) == "true"
}

// StringVariation returns string evaluation for target
func (e Evaluator) StringVariation(identifier string, target *Target, defaultValue string,
	processEvaluation ProcessEvaluation) string {

	variation, err := e.evaluate(identifier, target, "string", processEvaluation)
	if err != nil {
		log.Errorf("Error while evaluating string flag '%s', err: %v", identifier, err)
		return defaultValue
	}
	return variation.Value
}

// IntVariation returns int evaluation for target
func (e Evaluator) IntVariation(identifier string, target *Target, defaultValue int,
	processEvaluation ProcessEvaluation) int {

	variation, err := e.evaluate(identifier, target, "int", processEvaluation)
	if err != nil {
		log.Errorf("Error while evaluating int flag '%s', err: %v", identifier, err)
		return defaultValue
	}
	val, err := strconv.Atoi(variation.Value)
	if err != nil {
		return defaultValue
	}
	return val
}

// NumberVariation returns number evaluation for target
func (e Evaluator) NumberVariation(identifier string, target *Target, defaultValue float64,
	processEvaluation ProcessEvaluation) float64 {

	variation, err := e.evaluate(identifier, target, "number", processEvaluation)
	if err != nil {
		log.Errorf("Error while evaluating number flag '%s', err: %v", identifier, err)
		return defaultValue
	}
	val, err := strconv.ParseFloat(variation.Value, 64)
	if err != nil {
		return defaultValue
	}
	return val
}

// JSONVariation returns json evaluation for target
func (e Evaluator) JSONVariation(identifier string, target *Target,
	defaultValue map[string]interface{}, processEvaluation ProcessEvaluation) map[string]interface{} {

	variation, err := e.evaluate(identifier, target, "json", processEvaluation)
	if err != nil {
		log.Errorf("Error while evaluating json flag '%s', err: %v", identifier, err)
		return defaultValue
	}
	val := make(map[string]interface{})
	err = json.Unmarshal([]byte(variation.Value), &val)
	if err != nil {
		return defaultValue
	}
	return val
}
