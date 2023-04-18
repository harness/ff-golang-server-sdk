package evaluation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

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

// Query provides methods for segment and flag retrieval
type Query interface {
	GetSegment(identifier string) (rest.Segment, error)
	GetFlag(identifier string) (rest.FeatureConfig, error)
	GetFlags() ([]rest.FeatureConfig, error)
}

// FlagVariations list of FlagVariations
type FlagVariations []FlagVariation

// FlagVariation contains all required for ff-server To evaluate.
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

// Evaluator engine evaluates flag From provided query
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

	object := reflectValueToString(attrValue)

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
		// if evaluation is false just continue To next rule
		if !e.evaluateRule(&rule, target) {
			continue
		}

		// rule matched, Check if there is distribution
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

type EvaluationSummary struct {
	Checks []Check
}

type Check struct {
	stage  Stage
	result bool
}

type registerCheck func(name Stage, result bool)

type Stage string

const (
	flagExistsCheck             = "flag_exists"
	flagEnabledCheck            = "flag_enabled"
	prereqExistsCheck           = "prereq_exists"
	prereqPassesCheck           = "prereq_passes"
	targetRulesExistsCheck      = "target_rules_exists"
	targetRulesPassesCheck      = "target_rules_passes"
	targetGroupRulesExistsCheck = "target_group_rules_exists"
	targetGroupRulesPassesCheck = "target_group_rules_passes"
	distributionCheck           = "evaluate_distribution"
	returnVariation             = "return_variation"
)

func (e Evaluator) evaluateFlag(fc rest.FeatureConfig, target *Target, register registerCheck) (rest.Variation, error) {
	if register == nil {
		register = noopRegisterFunc
	}
	var variation = fc.OffVariation
	if fc.State == rest.FeatureStateOn {
		register(flagEnabledCheck, true)
		variation = ""
		if fc.VariationToTargetMap != nil && len(*fc.VariationToTargetMap) > 0 {
			register(targetRulesExistsCheck, true)
			variation = e.evaluateVariationMap(*fc.VariationToTargetMap, target)
			// register if a variation was selected or not
			if variation == "" {
				register(targetRulesPassesCheck, true)
			} else {
				register(targetRulesPassesCheck, false)
			}
		} else {
			register(targetRulesExistsCheck, false)
		}
		if variation == "" && fc.Rules != nil && len(*fc.Rules) > 0 {
			register(targetGroupRulesExistsCheck, true)
			variation = e.evaluateRules(*fc.Rules, target)

			// register if a variation was selected or not
			if variation == "" {
				register(targetGroupRulesPassesCheck, true)
			} else {
				register(targetGroupRulesPassesCheck, false)
			}
		} else {
			register(targetGroupRulesExistsCheck, false)
		}
		if variation == "" && fc.DefaultServe.Distribution != nil {
			register(distributionCheck, true)
			variation = evaluateDistribution(fc.DefaultServe.Distribution, target)
		}
		if variation == "" && fc.DefaultServe.Variation != nil {
			variation = *fc.DefaultServe.Variation
		}
	} else {
		register(flagEnabledCheck, false)
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
			e.logger.Debugf("Target %s excluded From segment %s via exclude list", target.Name, segment.Name)
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
		// if rules is nil pointer or points To the empty slice
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

			prereqEvaluatedVariation, err := e.evaluateFlag(prereqFeatureConfig, target, nil)
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

// takes uses feature store.List function To get all the flags.
func (e Evaluator) evaluateAll(target *Target) ([]FlagVariation, error) {
	var variations []FlagVariation
	flags, err := e.query.GetFlags()
	if err != nil {
		return variations, err
	}
	for _, f := range flags {
		v, _ := e.getVariationForTheFlag(f, target, nil)
		variations = append(variations, FlagVariation{f.Feature, f.Kind, v})
	}

	return variations, nil
}

// Evaluate exposes evaluate To the caller.
func (e Evaluator) Evaluate(identifier string, target *Target) (FlagVariation, error) {
	return e.evaluate(identifier, target, nil)
}

var noopRegisterFunc = func(name Stage, result bool) {}

// this is evaluating flag.
func (e Evaluator) evaluate(identifier string, target *Target, register registerCheck) (FlagVariation, error) {
	if register == nil {
		register = noopRegisterFunc
	}
	if e.query == nil {
		e.logger.Errorf(ErrQueryProviderMissing.Error())
		return FlagVariation{}, ErrQueryProviderMissing
	}
	flag, err := e.query.GetFlag(identifier)
	if err != nil {
		register(flagExistsCheck, false)
		return FlagVariation{}, err
	}
	register(flagExistsCheck, true)

	variation, err := e.getVariationForTheFlag(flag, target, register)
	if err != nil {
		return FlagVariation{}, err
	}

	register(returnVariation, true)
	return FlagVariation{flag.Feature, flag.Kind, variation}, nil
}

func (e Evaluator) ExplainEvaluate(identifier string, target *Target) (FlagVariation, EvaluationSummary, error) {
	summary := EvaluationSummary{}
	registerCallback := func(name Stage, result bool) {
		summary.Checks = append(summary.Checks, Check{
			stage:  name,
			result: result,
		})
	}
	variation, err := e.evaluate(identifier, target, registerCallback)
	return variation, summary, err
}

type Node struct {
	Id      Stage
	Label   string
	Enabled bool
}

type Edge struct {
	From    Stage
	To      Stage
	Label   string
	Enabled bool
}

// EvaluationSummaryToGraph converts an EvaluationSummary To a graph of nodes and edges
func EvaluationSummaryToGraph(summary EvaluationSummary) ([]Node, []Edge) {
	var nodes []Node
	var edges []Edge
	for i, check := range summary.Checks {
		nodes = append(nodes, Node{
			Id:      check.stage,
			Label:   string(check.stage),
			Enabled: true,
		})
		// add Edge From previous step To this step
		if i > 0 {
			edgeLabel := "No"
			if summary.Checks[i-1].result {
				edgeLabel = "Yes"
			}
			edges = append(edges, Edge{
				From:    nodes[i-1].Id,
				To:      nodes[i].Id,
				Label:   edgeLabel,
				Enabled: true,
			})
		}
	}

	return nodes, edges
}

// evaluates the flag and returns a proper variation.
func (e Evaluator) getVariationForTheFlag(flag rest.FeatureConfig, target *Target, register registerCheck) (rest.Variation, error) {
	if register == nil {
		register = noopRegisterFunc
	}
	if flag.Prerequisites != nil {
		register(prereqExistsCheck, true)
		prereq, err := e.checkPreRequisite(&flag, target)
		if err != nil || !prereq {
			register(prereqPassesCheck, false)
			return findVariation(flag.Variations, flag.OffVariation)
		}
		register(prereqPassesCheck, true)
	} else {
		register(prereqExistsCheck, false)
	}
	variation, err := e.evaluateFlag(flag, target, register)
	if err != nil {
		return rest.Variation{}, err
	}
	if e.postEvalCallback != nil {
		data := PostEvalData{
			FeatureConfig: &flag,
			Target:        target,
			Variation:     &variation,
		}

		e.postEvalCallback.PostEvaluateProcessor(&data)
	}
	return variation, nil
}

// BoolVariation returns boolean evaluation for target
func (e Evaluator) BoolVariation(identifier string, target *Target, defaultValue bool) bool {
	//flagVariation, err := e.evaluate(identifier, target, "boolean")
	// Check on/off and return
	flagVariation, err := e.evaluate(identifier, target, nil)
	if err != nil {
		e.logger.Errorf("Error while evaluating boolean flag '%s', err: %v", identifier, err)
		return defaultValue
	}
	return strings.ToLower(flagVariation.Variation.Value) == "true"
}

// StringVariation returns string evaluation for target
func (e Evaluator) StringVariation(identifier string, target *Target, defaultValue string) string {
	flagVariation, err := e.evaluate(identifier, target, nil)
	if err != nil {
		e.logger.Errorf("Error while evaluating string flag '%s', err: %v", identifier, err)
		return defaultValue
	}
	return flagVariation.Variation.Value
}

// IntVariation returns int evaluation for target
func (e Evaluator) IntVariation(identifier string, target *Target, defaultValue int) int {
	flagVariation, err := e.evaluate(identifier, target, nil)
	if err != nil {
		e.logger.Errorf("Error while evaluating int flag '%s', err: %v", identifier, err)
		return defaultValue
	}
	val, err := strconv.Atoi(flagVariation.Variation.Value)
	if err != nil {
		return defaultValue
	}
	return val
}

// NumberVariation returns number evaluation for target
func (e Evaluator) NumberVariation(identifier string, target *Target, defaultValue float64) float64 {
	//all numbers are stored as ints in the database
	flagVariation, err := e.evaluate(identifier, target, nil)
	if err != nil {
		e.logger.Errorf("Error while evaluating number flag '%s', err: %v", identifier, err)
		return defaultValue
	}
	val, err := strconv.ParseFloat(flagVariation.Variation.Value, 64)
	if err != nil {
		return defaultValue
	}
	return val
}

// JSONVariation returns json evaluation for target
func (e Evaluator) JSONVariation(identifier string, target *Target,
	defaultValue map[string]interface{}) map[string]interface{} {
	flagVariation, err := e.evaluate(identifier, target, nil)
	if err != nil {
		e.logger.Errorf("Error while evaluating json flag '%s', err: %v", identifier, err)
		return defaultValue
	}
	val := make(map[string]interface{})
	err = json.Unmarshal([]byte(flagVariation.Variation.Value), &val)
	if err != nil {
		return defaultValue
	}
	return val
}
