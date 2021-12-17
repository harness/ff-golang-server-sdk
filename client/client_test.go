package client_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/harness/ff-golang-server-sdk/client"
	"github.com/harness/ff-golang-server-sdk/dto"
	"github.com/harness/ff-golang-server-sdk/evaluation"
	"github.com/harness/ff-golang-server-sdk/rest"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	sdkKey = "27bed8d2-2610-462b-90eb-d80fd594b623"
	URL    = "http://localhost/api/1.0"
	//nolint
	AuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcm9qZWN0IjoiMTA0MjM5NzYtODQ1MS00NmZjLTg2NzctYmNiZDM3MTA3M2JhIiwiZW52aXJvbm1lbnQiOiI3ZWQxMDI1ZC1hOWIxLTQxMjktYTg4Zi1lMjdlZjM2MDk4MmQiLCJwcm9qZWN0SWRlbnRpZmllciI6IiIsImVudmlyb25tZW50SWRlbnRpZmllciI6IlByZVByb2R1Y3Rpb24iLCJhY2NvdW50SUQiOiIiLCJvcmdhbml6YXRpb24iOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDAiLCJjbHVzdGVySWRlbnRpZmllciI6IjEifQ.E4O_u42HkR0q4AwTTViFTCnNa89Kwftks7Gh-GvQfuE"
)

// TestMain runs before the other tests
func TestMain(m *testing.M) {
	// httpMock overwrites the http.DefaultClient
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Register Default Responders
	httpmock.RegisterResponder("POST", "http://localhost/api/1.0/client/auth", ValidAuthResponse)
	httpmock.RegisterResponder("GET", "http://localhost/api/1.0/client/env/7ed1025d-a9b1-4129-a88f-e27ef360982d/target-segments", TargetSegmentsResponse)
	httpmock.RegisterResponder("GET", "http://localhost/api/1.0/client/env/7ed1025d-a9b1-4129-a88f-e27ef360982d/feature-configs", FeatureConfigsResponse)

	os.Exit(m.Run())
}

func TestCfClient_BoolVariation(t *testing.T) {

	client, target, err := MakeNewClientAndTarget()
	if err != nil {
		t.Error(err)
	}

	type args struct {
		key          string
		target       *evaluation.Target
		defaultValue bool
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"Test Invalid Flag Name returns default value", args{"MadeUpIDontExist", target, false}, false, false},
		{"Test Default True Flag when On returns true", args{"TestTrueOn", target, false}, true, false},
		{"Test Default True Flag when Off returns false", args{"TestTrueOff", target, true}, false, false},
		{"Test Default False Flag when On returns false", args{"TestTrueOn", target, false}, true, false},
		{"Test Default False Flag when Off returns true", args{"TestTrueOff", target, true}, false, false},
		{"Test Default True Flag when Pre-Req is False returns false", args{"TestTrueOnWithPreReqFalse", target, true}, false, false},
		{"Test Default True Flag when Pre-Req is True returns true", args{"TestTrueOnWithPreReqTrue", target, true}, true, false},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			flag, err := client.BoolVariation(test.args.key, test.args.target, test.args.defaultValue)
			if (err != nil) != test.wantErr {
				t.Errorf("BoolVariation() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			assert.Equal(t, test.want, flag, "%s didn't get expected value", test.name)
		})
	}
}

func TestCfClient_StringVariation(t *testing.T) {

	client, target, err := MakeNewClientAndTarget()
	if err != nil {
		t.Error(err)
	}

	type args struct {
		key          string
		target       *evaluation.Target
		defaultValue string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Test Invalid Flag Name returns default value", args{"MadeUpIDontExist", target, "foo"}, "foo", false},
		{"Test Default String Flag with when On returns A", args{"TestStringAOn", target, "foo"}, "A", false},
		{"Test Default String Flag when Off returns B", args{"TestStringAOff", target, "foo"}, "B", false},
		{"Test Default String Flag when Pre-Req is False returns B", args{"TestStringAOnWithPreReqFalse", target, "foo"}, "B", false},
		{"Test Default String Flag when Pre-Req is True returns A", args{"TestStringAOnWithPreReqTrue", target, "foo"}, "A", false},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			flag, err := client.StringVariation(test.args.key, test.args.target, test.args.defaultValue)
			if (err != nil) != test.wantErr {
				t.Errorf("BoolVariation() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			assert.Equal(t, test.want, flag, "%s didn't get expected value", test.name)
		})
	}
}

// MakeNewClientAndTarget creates a new client and target.  If it returns
// error then something went wrong.
func MakeNewClientAndTarget() (*client.CfClient, *evaluation.Target, error) {
	target := target()
	client, err := newClient(http.DefaultClient)
	if err != nil {
		return nil, nil, err
	}

	// Wait to be authenticated - we can timeout if the channel doesn't return
	if ok, err := client.IsInitialized(); !ok {
		return nil, nil, err
	}

	return client, target, nil
}

// newClient creates a new client with some default options
func newClient(httpClient *http.Client) (*client.CfClient, error) {
	return client.NewCfClient(sdkKey,
		client.WithURL(URL),
		client.WithStreamEnabled(false),
		client.WithHTTPClient(httpClient),
		client.WithStoreEnabled(false),
	)
}

// target creates a new Target with some default values
func target() *evaluation.Target {
	target := dto.NewTargetBuilder("john").
		Firstname("John").
		Lastname("Doe").
		Email("john@doe.com").
		Build()
	return &target
}

var ValidAuthResponse = func(req *http.Request) (*http.Response, error) {
	return httpmock.NewJsonResponse(200, rest.AuthenticationResponse{
		AuthToken: AuthToken})
}

var TargetSegmentsResponse = func(req *http.Request) (*http.Response, error) {
	var AllSegmentsResponse []rest.Segment

	err := json.Unmarshal([]byte(`[
		{
			"environment": "PreProduction",
			"excluded": [],
			"identifier": "Beta_Users",
			"included": [
				{
					"identifier": "john",
					"name": "John",
				},
				{
					"identifier": "paul",
					"name": "Paul",
				}
			],
			"name": "Beta Users"
		}
	]`), &AllSegmentsResponse)
	if err != nil {
		return jsonError(err)
	}
	return httpmock.NewJsonResponse(200, AllSegmentsResponse)
}

var FeatureConfigsResponse = func(req *http.Request) (*http.Response, error) {
	var FeatureConfigResponse []rest.FeatureConfig
	FeatureConfigResponse = append(FeatureConfigResponse, MakeBoolFeatureConfigs("TestTrueOn", "true", "false", "on")...)
	FeatureConfigResponse = append(FeatureConfigResponse, MakeBoolFeatureConfigs("TestTrueOff", "true", "false", "off")...)

	FeatureConfigResponse = append(FeatureConfigResponse, MakeBoolFeatureConfigs("TestFalseOn", "false", "true", "on")...)
	FeatureConfigResponse = append(FeatureConfigResponse, MakeBoolFeatureConfigs("TestFalseOff", "false", "true", "off")...)

	FeatureConfigResponse = append(FeatureConfigResponse, MakeBoolFeatureConfigs("TestTrueOnWithPreReqFalse", "true", "false", "on", MakeBoolPreRequisite("PreReq1", "false"))...)
	FeatureConfigResponse = append(FeatureConfigResponse, MakeBoolFeatureConfigs("TestTrueOnWithPreReqTrue", "true", "false", "on", MakeBoolPreRequisite("PreReq1", "true"))...)

	FeatureConfigResponse = append(FeatureConfigResponse, MakeStringFeatureConfigs("TestStringAOn", "Alpha", "Bravo", "on")...)
	FeatureConfigResponse = append(FeatureConfigResponse, MakeStringFeatureConfigs("TestStringAOff", "Alpha", "Bravo", "off")...)

	FeatureConfigResponse = append(FeatureConfigResponse, MakeStringFeatureConfigs("TestStringAOnWithPreReqFalse", "Alpha", "Bravo", "on", MakeBoolPreRequisite("PreReq1", "false"))...)
	FeatureConfigResponse = append(FeatureConfigResponse, MakeStringFeatureConfigs("TestStringAOnWithPreReqTrue", "Alpha", "Bravo", "on", MakeBoolPreRequisite("PreReq1", "true"))...)

	return httpmock.NewJsonResponse(200, FeatureConfigResponse)
}
