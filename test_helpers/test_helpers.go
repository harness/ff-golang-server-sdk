package test_helpers

import (
	"fmt"
	"github.com/harness/ff-golang-server-sdk/rest"
	"github.com/jarcoal/httpmock"
	"net/http"
	"time"
)

func Cool() {

}

func MakeBoolFeatureConfigs(name, defaultVariation, offVariation, state string, preReqs ...rest.Prerequisite) []rest.FeatureConfig {
	var featureConfig []rest.FeatureConfig
	featureConfig = append(featureConfig, MakeBoolFeatureConfig(name, defaultVariation, offVariation, state, preReqs))

	// If there are any PreReqs then we need to store them as flags as well.
	for _, x := range preReqs {
		featureConfig = append(featureConfig, MakeBoolFeatureConfig(x.Feature, "true", "false",
			state, nil))
	}

	return featureConfig
}

func MakeBoolFeatureConfig(name, defaultVariation, offVariation, state string, preReqs []rest.Prerequisite) rest.FeatureConfig {
	return rest.FeatureConfig{
		DefaultServe: rest.Serve{
			Variation: &defaultVariation,
		},
		Environment:  "PreProduction",
		Feature:      name,
		Kind:         "boolean",
		OffVariation: offVariation,
		State:        rest.FeatureState(state),
		Variations: []rest.Variation{
			{Identifier: "true", Name: strPtr("True"), Value: "true"},
			{Identifier: "false", Name: strPtr("False"), Value: "false"},
		},
		Prerequisites: &preReqs,
		Version:       intPtr(1),
	}
}

func MakeBoolPreRequisite(name string, variation string) rest.Prerequisite {
	return rest.Prerequisite{
		Feature:    name,
		Variations: []string{variation},
	}
}

func MakeStringFeatureConfigs(name, defaultVariation, offVariation, state string, preReqs ...rest.Prerequisite) []rest.FeatureConfig {
	var featureConfig []rest.FeatureConfig
	featureConfig = append(featureConfig, MakeStringFeatureConfig(name, defaultVariation, offVariation, state, preReqs))

	// If there are any PreReqs then we need to store them as flags as well.
	for _, x := range preReqs {
		state := "off"
		if x.Variations[0] == "true" {
			state = "on"
		}
		featureConfig = append(featureConfig, MakeBoolFeatureConfig(x.Feature, "true", "false", state, nil))
	}

	return featureConfig
}

/*
{
		"defaultServe": {
			"variation": "Alpha"
		},
		"environment": "PreProduction",
		"feature": "TestStringFlag",
		"kind": "string",
		"offVariation": "Bravo",
		"prerequisites": [],
		"project": "Customer_Self_Service_Portal",
		"rules": [],
		"state": "off",
		"variations": [
			{
				"identifier": "Alpha",
				"name": "Bravo",
				"value": "A"
			},
			{
				"identifier": "Bravo",
				"name": "Bravo",
				"value": "B"
			}
		],
		"version": 1
	}
*/

func MakeStringFeatureConfig(name, defaultVariation, offVariation, state string, preReqs []rest.Prerequisite) rest.FeatureConfig {
	return rest.FeatureConfig{
		DefaultServe: rest.Serve{
			Variation: &defaultVariation,
		},
		Environment:  "PreProduction",
		Feature:      name,
		Kind:         "string",
		OffVariation: offVariation,
		State:        rest.FeatureState(state),
		Variations: []rest.Variation{
			{Identifier: "Alpha", Name: strPtr("Alpha"), Value: "A"},
			{Identifier: "Bravo", Name: strPtr("Bravo"), Value: "B"},
		},
		Prerequisites: &preReqs,
		Version:       intPtr(1),
	}
}

func JsonError(err error) (*http.Response, error) {
	return httpmock.NewJsonResponse(500, fmt.Errorf(`{"error" : "%s"}`, err))
}

func intPtr(value int64) *int64 {
	return &value
}

func strPtr(value string) *string {
	return &value
}

type MockSleeper struct {
	SleepTime time.Duration
}

func (ms MockSleeper) Sleep(d time.Duration) {
	time.Sleep(ms.SleepTime)
}
