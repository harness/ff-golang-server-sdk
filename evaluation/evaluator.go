package evaluation

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/harness/ff-golang-server-sdk/sdk_codes"

	"github.com/harness/ff-golang-server-sdk/logger"

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

var (
	ErrNilFlag = errors.New("flag is nil")
)

// Query provides methods for segment and flag retrieval
type Query interface {
	GetSegment(identifier string) (rest.Segment, error)
	GetFlag(identifier string) (rest.FeatureConfig, error)
	GetFlags() ([]rest.FeatureConfig, error)
	GetFlagMap() (map[string]*rest.FeatureConfig, error)
}

// FlagVariations list of FlagVariations
type FlagVariations []FlagVariation

// FlagVariation contains all required for ff-server to evaluate.
type FlagVariation struct {
	FlagIdentifier string
	Kind           rest.FeatureConfigKind
	Variation      rest.Variation
}

// PostEvalData holds information for post evaluation processing
type PostEvalData struct {
	FeatureConfig *rest.FeatureConfig
	Target        *Target
	Variation     *rest.Variation
}

// PostEvaluateCallback interface can be used for advanced processing
// of evaluated data
type PostEvaluateCallback interface {
	PostEvaluateProcessor(data *PostEvalData)
}

// Evaluator engine evaluates flag from provided query
type Evaluator struct {
	query            Query
	postEvalCallback PostEvaluateCallback
	logger           logger.Logger
}

// NewEvaluator constructs evaluator with query instance
func NewEvaluator(query Query, postEvalCallback PostEvaluateCallback, logger logger.Logger) (*Evaluator, error) {
	if query == nil {
		return nil, ErrQueryProviderMissing
	}
	return &Evaluator{
		logger:           logger,
		query:            query,
		postEvalCallback: postEvalCallback,
	}, nil
}

func (e Evaluator) evaluateClause(clause *rest.Clause, target *Target) bool {
	if clause == nil || len(clause.Values) == 0 || clause.Op == "" {
		e.logger.Debugf("Clause cannot be evaluated because operator is either nil, has no values or operation: Clause (%v)", clause)
		return false
	}

	value := clause.Values[0]
	attrValue := getAttrValue(target, clause.Attribute)

	if clause.Op != segmentMatchOperator && attrValue == "" {
		e.logger.Debugf("Operator is not a segment match and attribute value is not valid: Operator (%s), attributeVal (%s)", clause.Op, attrValue)
		return false
	}

	switch clause.Op {
	case startsWithOperator:
		return strings.HasPrefix(attrValue, value)
	case endsWithOperator:
		return strings.HasSuffix(attrValue, value)
	case matchOperator:
		found, err := regexp.MatchString(value, attrValue)
		return err == nil && found
	case containsOperator:
		return strings.Contains(attrValue, value)
	case equalOperator:
		return strings.EqualFold(attrValue, value)
	case equalSensitiveOperator:
		return attrValue == value
	case inOperator:
		for _, val := range clause.Values {
			if val == attrValue {
				return true
			}
		}
		return false
	case gtOperator:
		return attrValue > value
	case segmentMatchOperator:
		return e.isTargetIncludedOrExcludedInSegment(clause.Values, target)
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
		e.logger.Debugf("Serving Rules or Target are Nil")
		return ""
	}

	sort.SliceStable(servingRules, func(i, j int) bool {
		return servingRules[i].Priority < servingRules[j].Priority
	})
	for i := range servingRules {
		rule := servingRules[i]
		if len(rule.Clauses) == 0 {
			e.logger.Warnf("Serving Rule has no Clauses: Rule (%v)", rule)
		}
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
			e.logger.Debugf("Rule Matched for Target(%v), Variation returned (%v)", target, *rule.Serve.Variation)
			return *rule.Serve.Variation
		} else {
			e.logger.Warnf("No Variation on Serve for Rule (%v), Target (%v)", rule, target)
		}
	}
	return ""
}

// evaluateGroupRules evaluates the groups rules.  Note Group rule are represented by a rest.Clause, instead
// of a rest.Rule.   Unlike feature clauses which are AND'd, in a case of  a group these must be OR'd.
func (e Evaluator) evaluateGroupRules(rules []rest.Clause, target *Target) (bool, rest.Clause) {
	for _, r := range rules {
		rule := r
		if e.evaluateClause(&rule, target) {
			return true, r
		}
	}
	return false, rest.Clause{}
}

func (e Evaluator) evaluateVariationMap(variationsMap []rest.VariationMap, target *Target) string {
	if variationsMap == nil || target == nil {
		return ""
	}

	for _, variationMap := range variationsMap {
		if variationMap.Targets != nil {
			for _, t := range *variationMap.Targets {
				if t.Identifier != "" && t.Identifier == target.Identifier {
					e.logger.Debugf("Specific targeting matched in Variation Map: Variation Map (%v) Target(%v), Variation returned (%s)", t.Identifier, target, variationMap.Variation)
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
	if fc.State == rest.FeatureStateOn {
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
	} else {
		e.logger.Debugf("Flag is off: Flag(%s)", fc.Feature)
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
			e.logger.Errorf("Error on GetSegment returning false. Target (%v), Segment (%s)", target, segmentIdentifier)
			return false
		}
		// Should Target be excluded - if in excluded list we return false
		if segment.Excluded != nil && isTargetInList(target, *segment.Excluded) {
			e.logger.Debugf("Target (%v) excluded from segment %s via exclude list", target, segmentIdentifier)
			return false
		}

		// Should Target be included - if in included list we return true
		if segment.Included != nil && isTargetInList(target, *segment.Included) {
			e.logger.Debugf(
				"Target %s included in segment %s via include list",
				target.Name,
				segment.Name)
			return true
		}

		// Should Target be included via segment rules
		rules := segment.Rules
		// if rules is nil pointer or points to the empty slice
		if rules != nil && len(*rules) > 0 {
			if included, clause := e.evaluateGroupRules(*rules, target); included {
				e.logger.Debugf(
					"Target [%s] included in group [%s] via rule %+v", target.Name, segment.Name, clause)
				return true
			}
		}

	}
	return false
}

func (e Evaluator) checkPreRequisite(fc *rest.FeatureConfig, target *Target) (bool, error) {
	if e.query == nil {
		e.logger.Errorf(ErrQueryProviderMissing.Error())
		return true, ErrQueryProviderMissing
	}
	prerequisites := fc.Prerequisites
	if prerequisites != nil {
		e.logger.Debugf(
			"Checking pre requisites %v of parent feature %v",
			prerequisites,
			fc.Feature)
		for _, pre := range *prerequisites {
			prereqFeature := pre.Feature
			prereqFeatureConfig, err := e.query.GetFlag(prereqFeature)
			if err != nil {
				e.logger.Errorf(
					"Could not retrieve the pre requisite details of feature flag : %v", prereqFeature)
				return true, nil
			}

			prereqEvaluatedVariation, err := e.evaluateFlag(prereqFeatureConfig, target)
			if err != nil {
				e.logger.Errorf(
					"Could not evaluate the prerequisite details of feature flag : %v", prereqFeature)
				return true, nil
			}

			e.logger.Debugf(
				"Pre requisite flag %v has variation %v for target %v",
				prereqFeatureConfig.Feature,
				prereqEvaluatedVariation,
				target)

			// Compare if the pre requisite variation is a possible valid value of
			// the pre requisite FF
			validPrereqVariations := pre.Variations
			e.logger.Debugf(
				"Pre requisite flag %v should have the variations %v",
				prereqFeatureConfig.Feature,
				validPrereqVariations)
			if !contains(validPrereqVariations, prereqEvaluatedVariation.Identifier) {
				return false, nil
			}
			if r, _ := e.checkPreRequisite(&prereqFeatureConfig, target); !r {
				return false, nil
			}
		}
	}
	return true, nil
}

// EvaluateAll evaluates all the flags
func (e Evaluator) EvaluateAll(target *Target) (FlagVariations, error) {
	return e.evaluateAll(target)
}

// takes uses feature store.List function to get all the flags.
func (e Evaluator) evaluateAll(target *Target) ([]FlagVariation, error) {
	var variations []FlagVariation
	flags, err := e.query.GetFlagMap()
	if err != nil {
		return variations, err
	}
	for _, f := range flags {
		v, err := e.getVariationForTheFlag(f, target)
		if err != nil {
			e.logger.Warnf("Error Getting Variation for Flag: Flag (%s), Target (%v), Err: %s", f.Feature, target, err)
		}
		variations = append(variations, FlagVariation{f.Feature, f.Kind, v})
	}

	return variations, nil
}

// Evaluate exposes evaluate to the caller.
func (e Evaluator) Evaluate(identifier string, target *Target) (FlagVariation, error) {
	return e.evaluate(identifier, target)
}

// this is evaluating flag.
func (e Evaluator) evaluate(identifier string, target *Target) (FlagVariation, error) {
	e.logger.Debugf("Evaluating: Flag(%s) Target(%v)", identifier, target)
	if e.query == nil {
		e.logger.Errorf(ErrQueryProviderMissing.Error())
		return FlagVariation{}, ErrQueryProviderMissing
	}
	flag, err := e.query.GetFlag(identifier)
	if err != nil {
		e.logger.Warnf("Error Getting Flag: Flag (%s), Target(%v), Err: %s", identifier, target, err)
		return FlagVariation{}, err
	}

	variation, err := e.getVariationForTheFlag(&flag, target)
	if err != nil {
		e.logger.Warnf("Error Getting Variation for Flag: Flag (%s), Target(%v), Err: %s", identifier, target, err)
		return FlagVariation{}, err
	}
	return FlagVariation{flag.Feature, flag.Kind, variation}, nil
}

// evaluates the flag and returns a proper variation.
func (e Evaluator) getVariationForTheFlag(flag *rest.FeatureConfig, target *Target) (rest.Variation, error) {
	if flag == nil {
		return rest.Variation{}, ErrNilFlag
	}

	if flag.Prerequisites != nil {
		prereq, err := e.checkPreRequisite(flag, target)
		if err != nil || !prereq {
			return findVariation(flag.Variations, flag.OffVariation)
		}
	}
	variation, err := e.evaluateFlag(*flag, target)
	if err != nil {
		return rest.Variation{}, err
	}
	if e.postEvalCallback != nil {
		data := PostEvalData{
			FeatureConfig: flag,
			Target:        target,
			Variation:     &variation,
		}

		e.postEvalCallback.PostEvaluateProcessor(&data)
	}
	return variation, nil
}

// BoolVariation returns boolean evaluation for target
func (e Evaluator) BoolVariation(identifier string, target *Target, defaultValue bool) (bool, error) {
	//flagVariation, err := e.evaluate(identifier, target, "boolean")
	flagVariation, err := e.evaluate(identifier, target)
	if err != nil {
		return defaultValue, err
	}
	return strings.ToLower(flagVariation.Variation.Value) == "true", nil
}

// StringVariation returns string evaluation for target
func (e Evaluator) StringVariation(identifier string, target *Target, defaultValue string) (string, error) {
	flagVariation, err := e.evaluate(identifier, target)
	if err != nil {
		return defaultValue, err
	}
	return flagVariation.Variation.Value, nil
}

// IntVariation returns int evaluation for target
func (e Evaluator) IntVariation(identifier string, target *Target, defaultValue int) (int, error) {
	flagVariation, err := e.evaluate(identifier, target)
	if err != nil {
		return defaultValue, err
	}
	val, err := strconv.Atoi(flagVariation.Variation.Value)
	if err != nil {
		return defaultValue, err
	}
	return val, nil
}

// NumberVariation returns number evaluation for target
func (e Evaluator) NumberVariation(identifier string, target *Target, defaultValue float64) (float64, error) {
	//all numbers are stored as ints in the database
	flagVariation, err := e.evaluate(identifier, target)
	if err != nil {
		return defaultValue, err
	}
	val, err := strconv.ParseFloat(flagVariation.Variation.Value, 64)
	if err != nil {
		return defaultValue, err
	}
	return val, nil
}

// JSONVariation returns json evaluation for target
func (e Evaluator) JSONVariation(identifier string, target *Target,
	defaultValue map[string]interface{}) (map[string]interface{}, error) {
	flagVariation, err := e.evaluate(identifier, target)
	if err != nil {
		return defaultValue, err
	}
	val := make(map[string]interface{})
	err = json.Unmarshal([]byte(flagVariation.Variation.Value), &val)
	if err != nil {
		return defaultValue, err
	}
	e.logger.Debugf("%s Evaluated json flag successfully: '%s'", sdk_codes.EvaluationSuccess, identifier)
	return val, nil
}
