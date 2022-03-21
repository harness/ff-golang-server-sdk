package evaluation

import "github.com/harness/ff-golang-server-sdk/rest"

func convertWV(wv *rest.WeightedVariation) WeightedVariation {
	return WeightedVariation{
		Variation: wv.Variation,
		Weight:    wv.Weight,
	}
}

func convertDistribution(d *rest.Distribution) *Distribution {
	if d == nil {
		return nil
	}
	vars := make([]WeightedVariation, len(d.Variations))
	for i, val := range d.Variations {
		vars[i] = convertWV(&val)
	}
	return &Distribution{
		BucketBy:   d.BucketBy,
		Variations: vars,
	}
}

func convertVariation(v *rest.Variation) Variation {
	return Variation{
		Description: v.Description,
		Identifier:  v.Identifier,
		Name:        v.Name,
		Value:       v.Value,
	}
}

func convertServe(s *rest.Serve) Serve {
	return Serve{
		Distribution: convertDistribution(s.Distribution),
		Variation:    s.Variation,
	}
}

func convertClause(c *rest.Clause) Clause {
	return Clause{
		Attribute: c.Attribute,
		ID:        c.Id,
		Negate:    c.Negate,
		Op:        c.Op,
		Value:     c.Values,
	}
}

func convertServingRule(r *rest.ServingRule) ServingRule {
	clauses := make([]Clause, len(r.Clauses))
	for i, val := range r.Clauses {
		clauses[i] = convertClause(&val)
	}
	return ServingRule{
		Clauses:  clauses,
		Priority: r.Priority,
		RuleID:   r.RuleId,
		Serve:    convertServe(&r.Serve),
	}
}

func convertPrereq(p *rest.Prerequisite) Prerequisite {
	return Prerequisite{
		Feature:    p.Feature,
		Variations: p.Variations,
	}
}

//convert converts variation map to evaluation object
func convertVariationMap(v *rest.VariationMap) VariationMap {
	return VariationMap{
		TargetSegments: *v.TargetSegments,
		Targets:        convertTargetToIdentifier(*v.Targets),
		Variation:      v.Variation,
	}
}

func convertTargetToIdentifier(tm []rest.TargetMap) []string {
	result := make([]string, 0, len(tm))
	for j := range tm {
		result = append(result, *tm[j].Identifier)
	}
	return result
}

// NewFC feature flag from ff server to evaluation object
func NewFC(fc *rest.FeatureConfig) *FeatureConfig {
	vars := make(Variations, len(fc.Variations))
	for i, val := range fc.Variations {
		vars[i] = convertVariation(&val)
	}

	var rules ServingRules
	if fc.Rules != nil {
		rules = make(ServingRules, len(*fc.Rules))
		for i, val := range *fc.Rules {
			rules[i] = convertServingRule(&val)
		}
	}

	var pre []Prerequisite
	if fc.Prerequisites != nil {
		pre = make([]Prerequisite, len(*fc.Prerequisites))
		for i, val := range *fc.Prerequisites {
			pre[i] = convertPrereq(&val)
		}
	}
	defaultServe := Serve{}
	if fc.DefaultServe.Distribution != nil {
		defaultServe.Distribution = convertDistribution(fc.DefaultServe.Distribution)
	}
	if fc.DefaultServe.Variation != nil {
		defaultServe.Variation = fc.DefaultServe.Variation
	}
	var vtm []VariationMap
	if fc.VariationToTargetMap != nil {
		vtm = make([]VariationMap, len(*fc.VariationToTargetMap))
		for i, val := range *fc.VariationToTargetMap {
			vtm[i] = convertVariationMap(&val)
		}
	}
	return &FeatureConfig{
		DefaultServe:         defaultServe,
		Environment:          fc.Environment,
		Feature:              fc.Feature,
		Kind:                 fc.Kind,
		OffVariation:         fc.OffVariation,
		Prerequisites:        pre,
		Project:              fc.Project,
		Rules:                rules,
		State:                FeatureState(fc.State),
		VariationToTargetMap: vtm,
		Variations:           vars,
	}
}

// Convert REST segment response to evaluation segment object
func NewSegment(s *rest.Segment) *Segment {
	// need openspec change: included, excluded and rules should be required in response
	excluded := make(StrSlice, 0)
	if s.Excluded != nil {
		excluded = make(StrSlice, len(*s.Excluded))
		for i, excl := range *s.Excluded {
			excluded[i] = excl.Identifier
		}
	}

	included := make(StrSlice, 0)
	if s.Included != nil {
		included = make(StrSlice, len(*s.Included))
		for i, incl := range *s.Included {
			included[i] = incl.Identifier
		}
	}

	rules := make(SegmentRules, 0)
	if s.Rules != nil {
		rules = make(SegmentRules, len(*s.Rules))
		for i, rule := range *s.Rules {
			rules[i] = Clause{
				Attribute: rule.Attribute,
				ID:        rule.Id,
				Negate:    rule.Negate,
				Op:        rule.Op,
				Value:     rule.Values,
			}
		}
	}

	tags := make([]Tag, 0)
	if s.Rules != nil {
		if s.Tags != nil {
			tags = make([]Tag, len(*s.Tags))
			for i, tag := range *s.Tags {
				tags[i] = Tag{
					Name:  tag.Name,
					Value: tag.Value,
				}
			}
		}
	}

	var version int64
	if s.Version != nil {
		version = *s.Version
	}
	return &Segment{
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
