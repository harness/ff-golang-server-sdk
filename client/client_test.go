package client

import (
	"bytes"
	"encoding/json"
	"github.com/harness/ff-golang-server-sdk/dto"
	"github.com/harness/ff-golang-server-sdk/evaluation"
	"github.com/harness/ff-golang-server-sdk/log"
	"github.com/harness/ff-golang-server-sdk/rest"
	"github.com/harness/ff-golang-server-sdk/test_helpers"
	"github.com/harness/ff-golang-server-sdk/types"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	ValidSDKKey   = "27bed8d2-2610-462b-90eb-d80fd594b623"
	EmptySDKKey   = ""
	InvaliDSDKKey = "an invalid key"
	URL           = "http://localhost/api/1.0"

	//nolint
	ValidAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcm9qZWN0IjoiMTA0MjM5NzYtODQ1MS00NmZjLTg2NzctYmNiZDM3MTA3M2JhIiwiZW52aXJvbm1lbnQiOiI3ZWQxMDI1ZC1hOWIxLTQxMjktYTg4Zi1lMjdlZjM2MDk4MmQiLCJwcm9qZWN0SWRlbnRpZmllciI6IiIsImVudmlyb25tZW50SWRlbnRpZmllciI6IlByZVByb2R1Y3Rpb24iLCJhY2NvdW50SUQiOiIiLCJvcmdhbml6YXRpb24iOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDAiLCJjbHVzdGVySWRlbnRpZmllciI6IjEifQ.E4O_u42HkR0q4AwTTViFTCnNa89Kwftks7Gh-GvQfuE"
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
		// Stop running tests as the input is invalid at this stage.
		log.Fatal("Not enough responders provided to the test function being executed")
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
func registerMultipleResponseResponders(authResponder []httpmock.Responder, targetSegmentsResponder httpmock.Responder, featureConfigsResponder httpmock.Responder) {
	authQueue := makeResponderQueue(authResponder)
	httpmock.RegisterResponder("POST", "http://localhost/api/1.0/client/auth", authQueue.getNextResponder)

	// These responders don't need different responses per call
	httpmock.RegisterResponder("GET", "http://localhost/api/1.0/client/env/7ed1025d-a9b1-4129-a88f-e27ef360982d/target-segments", targetSegmentsResponder)
	httpmock.RegisterResponder("GET", "http://localhost/api/1.0/client/env/7ed1025d-a9b1-4129-a88f-e27ef360982d/feature-configs", featureConfigsResponder)
}

func TestCfClient_NewClient(t *testing.T) {

	tests := []struct {
		name          string
		newClientFunc func() (*CfClient, error)
		mockResponder func()
		err           error
	}{
		{
			name: "Synchronous Client: successfully initializes",
			newClientFunc: func() (*CfClient, error) {
				return newClient(http.DefaultClient, ValidSDKKey, WithWaitForInitialized(true))
			},
			mockResponder: func() {
				authSuccessResponse := AuthResponse(200, ValidAuthToken)
				registerResponders(authSuccessResponse, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			err: nil,
		},
		{
			name: "Synchronous Client: `IsInitialized` and `WaithWaitForInitialzed` called successfully initializes",
			newClientFunc: func() (*CfClient, error) {
				client, err := newClient(http.DefaultClient, ValidSDKKey, WithWaitForInitialized(true))
				if ok, err := client.IsInitialized(); !ok {
					return client, err
				}
				return client, err
			},
			mockResponder: func() {
				authSuccessResponse := AuthResponse(200, ValidAuthToken)
				registerResponders(authSuccessResponse, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			err: nil,
		},
		{
			name: "Synchronous client: Empty SDK key fails to initialize",
			newClientFunc: func() (*CfClient, error) {
				return newClient(http.DefaultClient, EmptySDKKey, WithWaitForInitialized(true))
			},
			mockResponder: nil,
			err:           types.ErrSdkCantBeEmpty,
		},
		{
			name: "Synchronous client: Authentication failed with 401 and no retry",
			newClientFunc: func() (*CfClient, error) {
				return newClient(http.DefaultClient, InvaliDSDKKey, WithWaitForInitialized(true))
			},
			mockResponder: func() {
				bodyString := `{
				"message": "invalid key or target provided",
				"code": "401"
				}`
				authErrorResponse := AuthResponseDetailed(401, "401", bodyString)
				registerResponders(authErrorResponse, TargetSegmentsResponse, FeatureConfigsResponse)
			},
			err: NonRetryableAuthError{
				StatusCode: "401",
				Message:    "invalid key or target provided",
			},
		},
		{
			name: "Asynchronous client: Authentication failed with 401 and no retry, times out waiting",
			newClientFunc: func() (*CfClient, error) {
				client, err := newClient(http.DefaultClient, InvaliDSDKKey, WithSleeper(test_helpers.MockSleeper{}))
				if ok, err := client.IsInitialized(); !ok {
					return client, err
				}
				return client, err
			},
			mockResponder: func() {
				bodyString := `{
				"message": "invalid key or target provided",
				"code": "401"
				}`
				authErrorResponse := AuthResponseDetailed(401, "401", bodyString)
				registerResponders(authErrorResponse, TargetSegmentsResponse, FeatureConfigsResponse)
			},
			// An async client cannot return an error for auth failures due
			err: InitializeTimeoutError{},
		},
		{
			name: "Synchronous client: Authentication failed with 403 and no retry",
			newClientFunc: func() (*CfClient, error) {
				return newClient(http.DefaultClient, ValidSDKKey, WithWaitForInitialized(true))
			},
			mockResponder: func() {
				bodyString := `{
				"message": "forbidden",
				"code": "403"
				}`
				authErrorResponse := AuthResponseDetailed(403, "403", bodyString)
				registerResponders(authErrorResponse, TargetSegmentsResponse, FeatureConfigsResponse)
			},
			err: NonRetryableAuthError{
				StatusCode: "403",
				Message:    "forbidden",
			},
		},
		{
			name: "Synchronous client: Authentication failed with 404 and no retry",
			newClientFunc: func() (*CfClient, error) {
				return newClient(http.DefaultClient, ValidSDKKey, WithWaitForInitialized(true))
			},
			mockResponder: func() {
				bodyString := `{
				"message": "not found",
				"code": "404"
				}`
				authErrorResponse := AuthResponseDetailed(404, "404", bodyString)
				registerResponders(authErrorResponse, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			err: NonRetryableAuthError{
				StatusCode: "404",
				Message:    "not found",
			},
		},
		{
			name: "Synchronous client: Authentication failed with 500 and succeeds after one retry",
			newClientFunc: func() (*CfClient, error) {
				return newClient(http.DefaultClient, ValidSDKKey, WithWaitForInitialized(true), WithSleeper(test_helpers.MockSleeper{}))
			},
			mockResponder: func() {
				bodyString := `{
				"message": "internal server error",
				"code": "500"
				}`
				firstAuthResponse := AuthResponseDetailed(500, "internal server error", bodyString)
				secondAuthResponse := AuthResponse(200, ValidAuthToken)

				registerMultipleResponseResponders([]httpmock.Responder{firstAuthResponse, secondAuthResponse}, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			err: nil,
		},
		{
			name: "Synchronous client: Authentication failed and succeeds just before exceeding max retries",
			newClientFunc: func() (*CfClient, error) {
				newClient, err := newClient(http.DefaultClient, ValidSDKKey, WithWaitForInitialized(true), WithMaxAuthRetries(10), WithSleeper(test_helpers.MockSleeper{}))
				return newClient, err
			},
			mockResponder: func() {
				bodyString := `{
				"message": "internal server error",
				"code": "500"
				}`
				var responses []httpmock.Responder
				// Add a bunch of error responses
				for i := 0; i < 10; i++ {
					responses = append(responses, AuthResponseDetailed(500, "internal server error", bodyString))
				}

				// Add the success response
				successResponse := AuthResponse(200, ValidAuthToken)
				responses = append(responses, successResponse)

				registerMultipleResponseResponders(responses, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			err: nil,
		},
		{
			name: "Synchronous client: Authentication failed and exceeds max retries",
			newClientFunc: func() (*CfClient, error) {
				newClient, err := newClient(http.DefaultClient, ValidSDKKey, WithWaitForInitialized(true), WithMaxAuthRetries(10), WithSleeper(test_helpers.MockSleeper{}))
				return newClient, err
			},
			mockResponder: func() {
				bodyString := `{
				"message": "internal server error",
				"code": "500"
				}`
				var responses []httpmock.Responder
				// Add a bunch of error responses
				for i := 0; i < 11; i++ {
					responses = append(responses, AuthResponseDetailed(500, "internal server error", bodyString))
				}

				registerMultipleResponseResponders(responses, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			err: RetryableAuthError{
				StatusCode: "500",
				Message:    "internal server error",
			},
		},
		{
			name: "Asynchronous client: `IsInitialized` called and successfully initializes",
			newClientFunc: func() (*CfClient, error) {
				client, err := newClient(http.DefaultClient, ValidSDKKey)
				if ok, err := client.IsInitialized(); !ok {
					return client, err
				}
				return client, err
			},
			mockResponder: func() {
				authSuccessResponse := AuthResponse(200, ValidAuthToken)
				registerResponders(authSuccessResponse, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			err: nil,
		},
		{
			name: "Asynchronous client: `IsInitialized` not called returns a client and no error",
			newClientFunc: func() (*CfClient, error) {
				client, err := newClient(http.DefaultClient, ValidSDKKey)
				return client, err
			},
			mockResponder: func() {
				authSuccessResponse := AuthResponse(200, ValidAuthToken)
				registerResponders(authSuccessResponse, TargetSegmentsResponse, FeatureConfigsResponse)
			},
			err: nil,
		},
		{
			name: "Asynchronous client: Empty SDK key, times out waiting",
			newClientFunc: func() (*CfClient, error) {
				client, err := newClient(http.DefaultClient, EmptySDKKey, WithSleeper(test_helpers.MockSleeper{}))
				if ok, err := client.IsInitialized(); !ok {
					return client, err
				}
				return client, err
			},
			mockResponder: nil,
			err:           InitializeTimeoutError{},
		},
		{
			name: "Asynchronous client: Authentication failed with 403 and no retry, times out waiting",
			newClientFunc: func() (*CfClient, error) {
				client, err := newClient(http.DefaultClient, ValidSDKKey, WithSleeper(test_helpers.MockSleeper{}))
				if ok, err := client.IsInitialized(); !ok {
					return client, err
				}
				return client, err
			},
			mockResponder: func() {
				bodyString := `{
				"message": "forbidden",
				"code": "403"
				}`
				authErrorResponse := AuthResponseDetailed(403, "403", bodyString)
				registerResponders(authErrorResponse, TargetSegmentsResponse, FeatureConfigsResponse)
			},
			err: InitializeTimeoutError{},
		},
		{
			name: "Asynchronous client: Authentication failed with 404 and times out waiting",
			newClientFunc: func() (*CfClient, error) {
				client, err := newClient(http.DefaultClient, ValidSDKKey, WithSleeper(test_helpers.MockSleeper{}))
				if ok, err := client.IsInitialized(); !ok {
					return client, err
				}
				return client, err
			},
			mockResponder: func() {
				bodyString := `{
				"message": "not found",
				"code": "404"
				}`
				authErrorResponse := AuthResponseDetailed(404, "404", bodyString)
				registerResponders(authErrorResponse, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			err: InitializeTimeoutError{},
		},
		{
			name: "Asynchronous client: Authentication failed with 500 and succeeds after one retry",
			newClientFunc: func() (*CfClient, error) {
				client, err := newClient(http.DefaultClient, ValidSDKKey, WithSleeper(test_helpers.MockSleeper{SleepTime: time.Millisecond}))
				if ok, err := client.IsInitialized(); !ok {
					return client, err
				}
				return client, err
			},
			mockResponder: func() {
				bodyString := `{
				"message": "internal server error",
				"code": "500"
				}`
				firstAuthResponse := AuthResponseDetailed(500, "internal server error", bodyString)
				secondAuthResponse := AuthResponse(200, ValidAuthToken)

				registerMultipleResponseResponders([]httpmock.Responder{firstAuthResponse, secondAuthResponse}, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			err: nil,
		},
		{
			name: "Asynchronous client: Authentication failed and succeeds just before exceeding max retries",
			newClientFunc: func() (*CfClient, error) {
				client, err := newClient(http.DefaultClient, ValidSDKKey, WithSleeper(test_helpers.MockSleeper{SleepTime: time.Millisecond}))
				if ok, err := client.IsInitialized(); !ok {
					return client, err
				}
				return client, err
			},
			mockResponder: func() {
				bodyString := `{
				"message": "internal server error",
				"code": "500"
				}`
				var responses []httpmock.Responder
				// Add a bunch of error responses
				for i := 0; i < 10; i++ {
					responses = append(responses, AuthResponseDetailed(500, "internal server error", bodyString))
				}

				// Add the success response
				successResponse := AuthResponse(200, ValidAuthToken)
				responses = append(responses, successResponse)

				registerMultipleResponseResponders(responses, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			err: nil,
		},
		{
			name: "Asynchronous client: Authentication failed and exceeds max retries",
			newClientFunc: func() (*CfClient, error) {
				client, err := newClient(http.DefaultClient, ValidSDKKey, WithSleeper(test_helpers.MockSleeper{SleepTime: time.Nanosecond}))
				if ok, err := client.IsInitialized(); !ok {
					return client, err
				}
				return client, err
			},
			mockResponder: func() {
				bodyString := `{
				"message": "internal server error",
				"code": "500"
				}`
				var responses []httpmock.Responder
				// Add a bunch of error responses
				for i := 0; i < 11; i++ {
					responses = append(responses, AuthResponseDetailed(500, "internal server error", bodyString))
				}

				registerMultipleResponseResponders(responses, TargetSegmentsResponse, FeatureConfigsResponse)

			},
			err: InitializeTimeoutError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.mockResponder != nil {
				tt.mockResponder()
			}
			client, err := tt.newClientFunc()

			// Even if we encounter an error during initialization, we still return the client but in an
			// un-unitialized state, meaning variation/close calls are handled in a special way, so we always
			// exect a non-nil client
			assert.NotNil(t, client)

			assert.Equal(t, tt.err, err)
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
				t.Errorf("BoolVariation() error = %v, err %v", err, test.wantErr)
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
				t.Errorf("BoolVariation() error = %v, err %v", err, test.wantErr)
				return
			}
			assert.Equal(t, test.want, flag, "%s didn't get expected value", test.name)
		})
	}
}

func TestCfClient_DefaultVariationReturned(t *testing.T) {
	tests := []struct {
		name           string
		clientFunc     func() (*CfClient, error)
		mockResponder  func()
		expectedBool   bool
		expectedString string
		expectedInt    int64
		expectedNumber float64
		expectedJSON   types.JSON
	}{
		{
			name: "Evaluations with Synchronous client with empty SDK key",
			clientFunc: func() (*CfClient, error) {
				return newClient(http.DefaultClient, EmptySDKKey, WithWaitForInitialized(true))
			},
			expectedBool:   false,
			expectedString: "a default value",
			expectedInt:    45555,
			expectedNumber: 45.222,
			expectedJSON:   types.JSON{"a default key": "a default value"},
		},
		{
			name: "Evaluations with Synchronous client with invalid SDK key",
			clientFunc: func() (*CfClient, error) {
				return newClient(http.DefaultClient, InvaliDSDKKey, WithWaitForInitialized(true))
			},
			mockResponder: func() {
				bodyString := `{
				"message": "invalid key or target provided",
				"code": "401"
				}`
				authErrorResponse := AuthResponseDetailed(401, "401", bodyString)
				registerResponders(authErrorResponse, TargetSegmentsResponse, FeatureConfigsResponse)
			},
			expectedBool:   false,
			expectedString: "a default value",
			expectedInt:    45555,
			expectedNumber: 45.222,
			expectedJSON:   types.JSON{"a default key": "a default value"},
		},
		{
			name: "Evaluations with Synchronous client with a server error",
			clientFunc: func() (*CfClient, error) {
				return newClient(http.DefaultClient, ValidSDKKey, WithWaitForInitialized(true), WithMaxAuthRetries(2), WithSleeper(test_helpers.MockSleeper{}))
			},
			mockResponder: func() {
				bodyString := `{
				"message": "internal server error",
				"code": "500"
				}`
				authErrorResponse := AuthResponseDetailed(500, "internal server error", bodyString)
				registerResponders(authErrorResponse, TargetSegmentsResponse, FeatureConfigsResponse)
			},
			expectedBool:   false,
			expectedString: "a default value",
			expectedInt:    45555,
			expectedNumber: 45.222,
			expectedJSON:   types.JSON{"a default key": "a default value"},
		},
		{
			name: "Evaluations with Synchronous client with empty SDK key",
			clientFunc: func() (*CfClient, error) {
				return newClient(http.DefaultClient, EmptySDKKey)
			},
			expectedBool:   false,
			expectedString: "a default value",
			expectedInt:    45555,
			expectedNumber: 45.222,
			expectedJSON:   types.JSON{"a default key": "a default value"},
		},
		{
			name: "Evaluations with Async client with invalid SDK key",
			clientFunc: func() (*CfClient, error) {
				return newClient(http.DefaultClient, InvaliDSDKKey)
			},
			mockResponder: func() {
				bodyString := `{
				"message": "invalid key or target provided",
				"code": "401"
				}`
				authErrorResponse := AuthResponseDetailed(401, "401", bodyString)
				registerResponders(authErrorResponse, TargetSegmentsResponse, FeatureConfigsResponse)
			},
			expectedBool:   false,
			expectedString: "a default value",
			expectedInt:    45555,
			expectedNumber: 45.222,
			expectedJSON:   types.JSON{"a default key": "a default value"},
		},
		{
			name: "Evaluations with Async client with a server error",
			clientFunc: func() (*CfClient, error) {
				return newClient(http.DefaultClient, ValidSDKKey, WithMaxAuthRetries(2), WithSleeper(test_helpers.MockSleeper{}))
			},
			mockResponder: func() {
				bodyString := `{
				"message": "internal server error",
				"code": "500"
				}`
				authErrorResponse := AuthResponseDetailed(500, "internal server error", bodyString)
				registerResponders(authErrorResponse, TargetSegmentsResponse, FeatureConfigsResponse)
			},
			expectedBool:   false,
			expectedString: "a default value",
			expectedInt:    45555,
			expectedNumber: 45.222,
			expectedJSON:   types.JSON{"a default key": "a default value"},
		},
		{
			name: "Evaluations with Async client with empty SDK key",
			clientFunc: func() (*CfClient, error) {
				return newClient(http.DefaultClient, EmptySDKKey)
			},
			expectedBool:   false,
			expectedString: "a default value",
			expectedInt:    45555,
			expectedNumber: 45.222,
			expectedJSON:   types.JSON{"a default key": "a default value"},
		},
	}
	target := target()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockResponder != nil {
				tt.mockResponder()
			}
			client, _ := tt.clientFunc()

			boolResult, err := client.BoolVariation("TestTrueOn", target, false)
			assert.Equal(t, tt.expectedBool, boolResult)
			assert.Nil(t, err)

			stringResult, err := client.StringVariation("TestTrueOn", target, "a default value")
			assert.Equal(t, tt.expectedString, stringResult)
			assert.Nil(t, err)

			intResult, err := client.IntVariation("TestTrueOn", target, tt.expectedInt)
			assert.Equal(t, tt.expectedInt, intResult)
			assert.Nil(t, err)

			numerResult, err := client.NumberVariation("TestTrueOn", target, tt.expectedNumber)
			assert.Equal(t, tt.expectedNumber, numerResult)
			assert.Nil(t, err)

			jsonResult, _ := client.JSONVariation("TestTrueOn", target, tt.expectedJSON)
			assert.Equal(t, tt.expectedJSON, jsonResult)
			assert.Nil(t, err)
		})
	}
}

// MakeNewSynchronousClientAndTarget creates a new synchronous client and target.  If it returns
// error then something went wrong.
func MakeNewSynchronousClientAndTarget(sdkKey string) (*CfClient, *evaluation.Target, error) {
	target := target()
	client, err := newClient(http.DefaultClient, sdkKey, WithWaitForInitialized(true))
	if err != nil {
		return nil, nil, err
	}
	return client, target, nil
}

// newClient creates a new client with some default options, and allows extra options if required.
func newClient(httpClient *http.Client, sdkKey string, extraOptions ...ConfigOption) (*CfClient, error) {
	baseOptions := []ConfigOption{
		WithURL(URL),
		WithStreamEnabled(false),
		WithHTTPClient(httpClient),
		WithStoreEnabled(false),
	}

	allOptions := append(baseOptions, extraOptions...)
	return NewCfClient(sdkKey, allOptions...)
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
		response := &http.Response{
			StatusCode: statusCode,
			Status:     status,
			Body:       io.NopCloser(bytes.NewReader([]byte(bodyString))), // this is your JSON body as io.ReadCloser
			Header:     make(http.Header),
		}

		response.Header.Add("Content-Type", "application/json")

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
		return test_helpers.JsonError(err)
	}
	return httpmock.NewJsonResponse(200, AllSegmentsResponse)
}

var FeatureConfigsResponse = func(req *http.Request) (*http.Response, error) {
	var FeatureConfigResponse []rest.FeatureConfig
	FeatureConfigResponse = append(FeatureConfigResponse, test_helpers.MakeBoolFeatureConfigs("TestTrueOn", "true", "false", "on")...)
	FeatureConfigResponse = append(FeatureConfigResponse, test_helpers.MakeBoolFeatureConfigs("TestTrueOff", "true", "false", "off")...)

	FeatureConfigResponse = append(FeatureConfigResponse, test_helpers.MakeBoolFeatureConfigs("TestFalseOn", "false", "true", "on")...)
	FeatureConfigResponse = append(FeatureConfigResponse, test_helpers.MakeBoolFeatureConfigs("TestFalseOff", "false", "true", "off")...)

	FeatureConfigResponse = append(FeatureConfigResponse, test_helpers.MakeBoolFeatureConfigs("TestTrueOnWithPreReqFalse", "true", "false", "on", test_helpers.MakeBoolPreRequisite("PreReq1", "false"))...)
	FeatureConfigResponse = append(FeatureConfigResponse, test_helpers.MakeBoolFeatureConfigs("TestTrueOnWithPreReqTrue", "true", "false", "on", test_helpers.MakeBoolPreRequisite("PreReq1", "true"))...)

	FeatureConfigResponse = append(FeatureConfigResponse, test_helpers.MakeStringFeatureConfigs("TestStringAOn", "Alpha", "Bravo", "on")...)
	FeatureConfigResponse = append(FeatureConfigResponse, test_helpers.MakeStringFeatureConfigs("TestStringAOff", "Alpha", "Bravo", "off")...)

	FeatureConfigResponse = append(FeatureConfigResponse, test_helpers.MakeStringFeatureConfigs("TestStringAOnWithPreReqFalse", "Alpha", "Bravo", "on", test_helpers.MakeBoolPreRequisite("PreReq1", "false"))...)
	FeatureConfigResponse = append(FeatureConfigResponse, test_helpers.MakeStringFeatureConfigs("TestStringAOnWithPreReqTrue", "Alpha", "Bravo", "on", test_helpers.MakeBoolPreRequisite("PreReq1", "true"))...)

	return httpmock.NewJsonResponse(200, FeatureConfigResponse)
}

func TestCfClient_Close(t *testing.T) {
	registerResponders(AuthResponse(200, ValidAuthToken), TargetSegmentsResponse, TargetSegmentsResponse)
	client, err := newClient(&http.Client{}, ValidSDKKey, WithWaitForInitialized(true))
	if err != nil {
		t.Error(err)
	}

	t.Log("When I close the client for the first time I should not get an error")
	assert.Nil(t, client.Close())

	t.Log("When I close the client for the second time I should an error")
	assert.NotNil(t, client.Close())

	t.Log("When I close the client before it's been initialized I should get an error")

	client2, err := newClient(&http.Client{}, ValidSDKKey)
	if err != nil {
		t.Error(err)
	}
	assert.NotNil(t, client2.Close())
}
