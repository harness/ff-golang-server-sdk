package client_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/harness/ff-golang-server-sdk/types"
	"io"
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
	ValidSDKKey = "27bed8d2-2610-462b-90eb-d80fd594b623"
	EmptySDKKey = ""
	URL         = "http://localhost/api/1.0"

	//nolint
	ValidAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcm9qZWN0IjoiMTA0MjM5NzYtODQ1MS00NmZjLTg2NzctYmNiZDM3MTA3M2JhIiwiZW52aXJvbm1lbnQiOiI3ZWQxMDI1ZC1hOWIxLTQxMjktYTg4Zi1lMjdlZjM2MDk4MmQiLCJwcm9qZWN0SWRlbnRpZmllciI6IiIsImVudmlyb25tZW50SWRlbnRpZmllciI6IlByZVByb2R1Y3Rpb24iLCJhY2NvdW50SUQiOiIiLCJvcmdhbml6YXRpb24iOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDAiLCJjbHVzdGVySWRlbnRpZmllciI6IjEifQ.E4O_u42HkR0q4AwTTViFTCnNa89Kwftks7Gh-GvQfuE"
	EmptyAuthToken = ""
)

// responderQueue is a type that manages a queue of responders
type responderQueue struct {
	responders []httpmock.Responder
	index      int
}

// makeResponderQueue creates a new instance of responderQueue with the provided responders
func makeResponderQueue(responders []httpmock.Responder) *responderQueue {
	return &responderQueue{
		responders: responders,
		index:      0,
	}
}

// getNextResponder is a method that returns the next responder in the queue
func (q *responderQueue) getNextResponder(req *http.Request) (*http.Response, error) {
	if q.index >= len(q.responders) {
		return nil, fmt.Errorf("no more responders in the queue")
	}
	responder := q.responders[q.index]
	q.index++
	return responder(req)
}

// TestMain runs before the other tests
func TestMain(m *testing.M) {
	// httpMock overwrites the http.DefaultClient
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	os.Exit(m.Run())
}

func registerResponders(authResponder httpmock.Responder, targetSegmentsResponder httpmock.Responder, featureConfigsResponder httpmock.Responder) {
	httpmock.RegisterResponder("POST", "http://localhost/api/1.0/client/auth", authResponder)
	httpmock.RegisterResponder("GET", "http://localhost/api/1.0/client/env/7ed1025d-a9b1-4129-a88f-e27ef360982d/target-segments", targetSegmentsResponder)
	httpmock.RegisterResponder("GET", "http://localhost/api/1.0/client/env/7ed1025d-a9b1-4129-a88f-e27ef360982d/feature-configs", featureConfigsResponder)
}

// Same as registerResponders except the auth response can be different per call
func registerStatefulResponders(authResponder []httpmock.Responder, targetSegmentsResponder httpmock.Responder, featureConfigsResponder httpmock.Responder) {
	authQueue := makeResponderQueue(authResponder)
	httpmock.RegisterResponder("POST", "http://localhost/api/1.0/client/auth", authQueue.getNextResponder)

	// These responders don't need different responses per call
	httpmock.RegisterResponder("GET", "http://localhost/api/1.0/client/env/7ed1025d-a9b1-4129-a88f-e27ef360982d/target-segments", targetSegmentsResponder)
	httpmock.RegisterResponder("GET", "http://localhost/api/1.0/client/env/7ed1025d-a9b1-4129-a88f-e27ef360982d/feature-configs", featureConfigsResponder)
}

func TestCfClient_NewClient(t *testing.T) {

	tests := []struct {
		name          string
		newClientFunc func() (*client.CfClient, error)
		mockResponder func()
		wantErr       error
	}{
		{
			name: "Successful client creation",
			newClientFunc: func() (*client.CfClient, error) {
				return newSynchronousClient(http.DefaultClient, ValidSDKKey)
			},
			mockResponder: func() {
				authSuccessResponse := AuthResponse(200, ValidAuthToken)
				registerResponders(authSuccessResponse, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			wantErr: nil,
		},
		{
			name: "Empty SDK key",
			newClientFunc: func() (*client.CfClient, error) {
				return newSynchronousClient(http.DefaultClient, EmptySDKKey)
			},
			mockResponder: nil,
			wantErr:       types.ErrSdkCantBeEmpty,
		},
		{
			name: "Authentication failed with 401 and no retry",
			newClientFunc: func() (*client.CfClient, error) {
				return newSynchronousClient(http.DefaultClient, ValidSDKKey) // A function that returns a CfClient instance
			},
			mockResponder: func() {
				bodyString := `{
				"message": "invalid key or target provided",
				"code": "401"
				}`
				authErrorResponse := AuthResponseDetailed(403, "403", bodyString)
				registerResponders(authErrorResponse, TargetSegmentsResponse, FeatureConfigsResponse)
			},
			wantErr: client.NonRetryableAuthError{
				StatusCode: "401",
				Message:    "invalid key or target provided",
			},
		},
		{
			name: "Authentication failed with 403 and no retry",
			newClientFunc: func() (*client.CfClient, error) {
				return newSynchronousClient(http.DefaultClient, ValidSDKKey) // A function that returns a CfClient instance
			},
			mockResponder: func() {
				bodyString := `{
				"message": "forbidden",
				"code": "403"
				}`
				authErrorResponse := AuthResponseDetailed(403, "403", bodyString)
				registerResponders(authErrorResponse, TargetSegmentsResponse, FeatureConfigsResponse)
			},
			wantErr: client.NonRetryableAuthError{
				StatusCode: "403",
				Message:    "forbidden",
			},
		},
		{
			name: "Authentication failed with 404 and no retry",
			newClientFunc: func() (*client.CfClient, error) {
				return newSynchronousClient(http.DefaultClient, ValidSDKKey) // A function that returns a CfClient instance
			},
			mockResponder: func() {
				bodyString := `{
				"message": "not found",
				"code": "404"
				}`
				authErrorResponse := AuthResponseDetailed(404, "404", bodyString)
				registerResponders(authErrorResponse, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			wantErr: client.NonRetryableAuthError{
				StatusCode: "404",
				Message:    "not found",
			},
		},
		{
			name: "Authentication failed with 500 and retries once",
			newClientFunc: func() (*client.CfClient, error) {
				return newSynchronousClient(http.DefaultClient, ValidSDKKey)
			},
			mockResponder: func() {
				bodyString := `{
				"message": "internal server error",
				"code": "500"
				}`
				firstAuthResponse := AuthResponseDetailed(500, "success", bodyString)
				secondAuthResponse := AuthResponse(200, ValidAuthToken)

				registerStatefulResponders([]httpmock.Responder{firstAuthResponse, secondAuthResponse}, TargetSegmentsResponse, FeatureConfigsResponse)
				//registerResponders(authSuccessResponse, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.mockResponder != nil {
				tt.mockResponder()
			}
			_, err := tt.newClientFunc()

			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestCfClient_BoolVariation(t *testing.T) {
	authSuccessResponse := AuthResponse(200, ValidAuthToken)
	registerResponders(authSuccessResponse, TargetSegmentsResponse, FeatureConfigsResponse)
	client, target, err := MakeNewSynchronousClientAndTarget(ValidSDKKey)
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
	authSuccessResponse := AuthResponse(200, ValidAuthToken)
	registerResponders(authSuccessResponse, TargetSegmentsResponse, FeatureConfigsResponse)

	client, target, err := MakeNewSynchronousClientAndTarget(ValidSDKKey)
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

// MakeNewSynchronousClientAndTarget creates a new synchronous client and target.  If it returns
// error then something went wrong.
func MakeNewSynchronousClientAndTarget(sdkKey string) (*client.CfClient, *evaluation.Target, error) {
	target := target()
	client, err := newSynchronousClient(http.DefaultClient, sdkKey)
	if err != nil {
		return nil, nil, err
	}
	return client, target, nil
}

// newAsyncClient creates a new client with some default options
func newAsyncClient(httpClient *http.Client) (*client.CfClient, error) {
	return client.NewCfClient(ValidSDKKey,
		client.WithURL(URL),
		client.WithStreamEnabled(false),
		client.WithHTTPClient(httpClient),
		client.WithStoreEnabled(false),
	)
}

func newSynchronousClient(httpClient *http.Client, sdkKey string) (*client.CfClient, error) {
	return client.NewCfClient(sdkKey,
		client.WithURL(URL),
		client.WithStreamEnabled(false),
		client.WithHTTPClient(httpClient),
		client.WithStoreEnabled(false),
		client.WithWaitForInitialized(true),
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

var AuthResponse = func(statusCode int, authToken string) func(req *http.Request) (*http.Response, error) {

	return func(req *http.Request) (*http.Response, error) {
		// Return the appropriate error based on the provided status code
		return httpmock.NewJsonResponse(statusCode, rest.AuthenticationResponse{
			AuthToken: authToken})
	}
}

var AuthResponseDetailed = func(statusCode int, status string, bodyString string) func(req *http.Request) (*http.Response, error) {

	return func(req *http.Request) (*http.Response, error) {
		// Return the appropriate error based on the provided status code
		response := &http.Response{
			StatusCode: statusCode,
			Status:     status,
			Body:       io.NopCloser(bytes.NewReader([]byte(bodyString))), // this is your JSON body as io.ReadCloser
			Header:     make(http.Header),
		}

		// Add headers to the response
		response.Header.Add("Content-Type", "application/json")
		// You can add other headers as needed:
		// response.Header.Add("Another-Header", "Header-Value")

		return response, nil
	}
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

func TestCfClient_Close(t *testing.T) {
	client, err := newSynchronousClient(&http.Client{}, ValidSDKKey)
	if err != nil {
		t.Error(err)
	}

	t.Log("When I close the client for the first time I should not get an error")
	assert.Nil(t, client.Close())

	t.Log("When I close the client for the second time I should an error")
	assert.NotNil(t, client.Close())
}
