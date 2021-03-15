package rest

import "github.com/drone/ff-golang-server-sdk.v1/evaluation"

func (wv WeightedVariation) convert() *evaluation.WeightedVariation {
	return &evaluation.WeightedVariation{
		Variation: wv.Variation,
		Weight:    wv.Weight,
	}
}

func (d *Distribution) convert() *evaluation.Distribution {
	if d == nil {
		return nil
	}
	vars := make([]evaluation.WeightedVariation, len(d.Variations))
	for i, val := range d.Variations {
		vars[i] = *val.convert()
	}
	return &evaluation.Distribution{
		BucketBy:   d.BucketBy,
		Variations: vars,
	}
}

func (v Variation) convert() *evaluation.Variation {
	return &evaluation.Variation{
		Description: v.Description,
		Identifier:  v.Identifier,
		Name:        v.Name,
		Value:       v.Value,
	}
}

func (s Serve) convert() *evaluation.Serve {
	return &evaluation.Serve{
		Distribution: s.Distribution.convert(),
		Variation:    s.Variation,
	}
}

func (c Clause) convert() *evaluation.Clause {
	return &evaluation.Clause{
		Attribute: c.Attribute,
		ID:        c.Id,
		Negate:    c.Negate,
		Op:        c.Op,
		Value:     c.Values,
	}
}

func (r ServingRule) convert() *evaluation.ServingRule {
	clauses := make([]evaluation.Clause, len(r.Clauses))
	for i, val := range r.Clauses {
		clauses[i] = *val.convert()
	}
	return &evaluation.ServingRule{
		Clauses:  clauses,
		Priority: r.Priority,
		RuleID:   r.RuleId,
		Serve:    *r.Serve.convert(),
	}
}

func (p Prerequisite) convert() *evaluation.Prerequisite {
	return &evaluation.Prerequisite{
		Feature:    p.Feature,
		Variations: p.Variations,
	}
}

// Convert feature flag from ff server to evaluation object
func (fc FeatureConfig) Convert() *evaluation.FeatureConfig {
	vars := make(evaluation.Variations, len(fc.Variations))
	for i, val := range fc.Variations {
		vars[i] = *val.convert()
	}

	var rules evaluation.ServingRules
	if fc.Rules != nil {
		rules = make(evaluation.ServingRules, len(*fc.Rules))
		for i, val := range *fc.Rules {
			rules[i] = *val.convert()
		}
	}

	var pre []evaluation.Prerequisite
	if fc.Prerequisites != nil {
		pre = make([]evaluation.Prerequisite, len(*fc.Prerequisites))
		for i, val := range *fc.Prerequisites {
			pre[i] = *val.convert()
		}
	}
	defaultServe := evaluation.Serve{}
	if fc.DefaultServe.Distribution != nil {
		defaultServe.Distribution = fc.DefaultServe.Distribution.convert()
	}
	if fc.DefaultServe.Variation != nil {
		defaultServe.Variation = fc.DefaultServe.Variation
	}
	return &evaluation.FeatureConfig{
		DefaultServe:         defaultServe,
		Environment:          fc.Environment,
		Feature:              fc.Feature,
		Kind:                 fc.Kind,
		OffVariation:         fc.OffVariation,
		Prerequisites:        pre,
		Project:              fc.Project,
		Rules:                rules,
		State:                evaluation.FeatureState(fc.State),
		VariationToTargetMap: nil,
		Variations:           vars,
	}
}
