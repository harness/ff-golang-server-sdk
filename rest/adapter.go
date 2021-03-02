package rest

import "github.com/wings-software/ff-client-sdk-go/evaluation"

func (wv WeightedVariation) DomainEntity() *evaluation.WeightedVariation {
	return &evaluation.WeightedVariation{
		Variation: wv.Variation,
		Weight:    wv.Weight,
	}
}

func (d *Distribution) DomainEntity() *evaluation.Distribution {
	if d == nil {
		return nil
	}
	vars := make([]evaluation.WeightedVariation, len(d.Variations))
	for i, val := range d.Variations {
		vars[i] = *val.DomainEntity()
	}
	return &evaluation.Distribution{
		BucketBy:   d.BucketBy,
		Variations: vars,
	}
}

func (v Variation) DomainEntity() *evaluation.Variation {
	return &evaluation.Variation{
		Description: v.Description,
		Identifier:  v.Identifier,
		Name:        v.Name,
		Value:       v.Value,
	}
}

func (s Serve) DomainEntity() *evaluation.Serve {
	return &evaluation.Serve{
		Distribution: s.Distribution.DomainEntity(),
		Variation:    s.Variation,
	}
}

func (c Clause) DomainEntity() *evaluation.Clause {
	return &evaluation.Clause{
		Attribute: c.Attribute,
		Id:        c.Id,
		Negate:    c.Negate,
		Op:        c.Op,
		Value:     c.Values,
	}
}

func (r ServingRule) DomainEntity() *evaluation.ServingRule {
	clauses := make([]evaluation.Clause, len(r.Clauses))
	for i, val := range r.Clauses {
		clauses[i] = *val.DomainEntity()
	}
	return &evaluation.ServingRule{
		Clauses:  clauses,
		Priority: r.Priority,
		RuleId:   r.RuleId,
		Serve:    *r.Serve.DomainEntity(),
	}
}

func (p Prerequisite) DomainEntity() *evaluation.Prerequisite {
	return &evaluation.Prerequisite{
		Feature:    p.Feature,
		Variations: p.Variations,
	}
}

func (fc FeatureConfig) DomainEntity() *evaluation.FeatureConfig {
	vars := make(evaluation.Variations, len(fc.Variations))
	for i, val := range fc.Variations {
		vars[i] = *val.DomainEntity()
	}

	var rules evaluation.ServingRules
	if fc.Rules != nil {
		rules = make(evaluation.ServingRules, len(*fc.Rules))
		for i, val := range *fc.Rules {
			rules[i] = *val.DomainEntity()
		}
	}

	var pre []evaluation.Prerequisite
	if fc.Prerequisites != nil {
		pre = make([]evaluation.Prerequisite, len(*fc.Prerequisites))
		for i, val := range *fc.Prerequisites {
			pre[i] = *val.DomainEntity()
		}
	}
	defaultServe := evaluation.Serve{}
	if fc.DefaultServe.Distribution != nil {
		defaultServe.Distribution = fc.DefaultServe.Distribution.DomainEntity()
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
