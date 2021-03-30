package rest

import "github.com/drone/ff-golang-server-sdk.v0/evaluation"

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

// Convert REST segment response to evaluation segment object
func (s Segment) Convert() evaluation.Segment {
	// need openspec change: included, excluded and rules should be required in response
	excluded := make(evaluation.StrSlice, 0)
	if s.Excluded != nil {
		excluded = make(evaluation.StrSlice, len(*s.Excluded))
		for i, excl := range *s.Excluded {
			excluded[i] = excl.Name
		}
	}

	included := make(evaluation.StrSlice, 0)
	if s.Included != nil {
		included = make(evaluation.StrSlice, len(*s.Included))
		for i, incl := range *s.Included {
			included[i] = incl.Name
		}
	}

	rules := make(evaluation.Clauses, 0)
	if s.Rules != nil {
		rules = make(evaluation.Clauses, len(*s.Rules))
		for i, rule := range *s.Rules {
			rules[i] = evaluation.Clause{
				Attribute: rule.Attribute,
				ID:        rule.Id,
				Negate:    rule.Negate,
				Op:        rule.Op,
				Value:     rule.Values,
			}
		}
	}

	tags := make([]evaluation.Tag, 0)
	if s.Rules != nil {
		tags = make([]evaluation.Tag, len(*s.Tags))
		for i, tag := range *s.Tags {
			tags[i] = evaluation.Tag{
				Name:  tag.Name,
				Value: tag.Value,
			}
		}
	}

	var version int64
	if s.Version != nil {
		version = *s.Version
	}
	return evaluation.Segment{
		Identifier:  s.Identifier,
		Name:        s.Name,
		CreatedAt:   s.CreatedAt,
		ModifiedAt:  s.ModifiedAt,
		Environment: s.Environment,
		Excluded:    excluded,
		Included:    included,
		Rules:       rules,
		Tags:        tags,
		Version:     version,
	}
}
