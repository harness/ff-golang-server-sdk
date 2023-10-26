package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/harness/ff-golang-server-sdk/sdk_codes"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/harness/ff-golang-server-sdk/evaluation"
	"github.com/harness/ff-golang-server-sdk/logger"

	"github.com/harness/ff-golang-server-sdk/pkg/repository"

	"github.com/harness/ff-golang-server-sdk/analyticsservice"
	"github.com/harness/ff-golang-server-sdk/metricsclient"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/golang-jwt/jwt"
	"github.com/harness/ff-golang-server-sdk/rest"
	"github.com/harness/ff-golang-server-sdk/stream"
	"github.com/harness/ff-golang-server-sdk/types"

	"github.com/r3labs/sse/v2"
)

// CfClient is the Feature Flag client.
//
// This object evaluates feature flags and communicates with Feature Flag services.
// Applications should instantiate a single instance for the lifetime of their application
// and share it wherever feature flags need to be evaluated.
//
// When an application is shutting down or no longer needs to use the CfClient instance, it
// should call Close() to ensure that all of its connections and goroutines are shut down and
// that any pending analytics events have been delivered.
type CfClient struct {
	evaluator               *evaluation.Evaluator
	repository              repository.Repository
	mux                     sync.RWMutex
	api                     rest.ClientWithResponsesInterface
	metricsapi              metricsclient.ClientWithResponsesInterface
	sdkKey                  string
	auth                    rest.AuthenticationRequest
	config                  *config
	environmentID           string
	token                   string
	streamConnectedBool     bool
	streamConnectedBoolLock sync.RWMutex
	streamConnected         chan struct{}
	streamDisconnected      chan error
	authenticated           chan struct{}
	postEvalChan            chan evaluation.PostEvalData
	initializedBool         bool
	initializedBoolLock     sync.RWMutex
	initialized             chan struct{}
	initializedErr          chan error
	analyticsService        *analyticsservice.AnalyticsService
	clusterIdentifier       string
	stop                    chan struct{}
	stopped                 *atomicBool
}

// NewCfClient creates a new client instance that connects to CF with the default configuration.
// For advanced configuration options use ConfigOptions functions
func NewCfClient(sdkKey string, options ...ConfigOption) (*CfClient, error) {

	//  functional options for config
	config := newDefaultConfig(getLogger(options...))
	for _, opt := range options {
		opt(config)
	}

	analyticsService := analyticsservice.NewAnalyticsService(time.Minute, config.Logger)

	client := &CfClient{
		sdkKey:             sdkKey,
		config:             config,
		authenticated:      make(chan struct{}),
		analyticsService:   analyticsService,
		clusterIdentifier:  "1",
		postEvalChan:       make(chan evaluation.PostEvalData),
		stop:               make(chan struct{}),
		stopped:            newAtomicBool(false),
		initialized:        make(chan struct{}),
		initializedErr:     make(chan error),
		streamConnected:    make(chan struct{}),
		streamDisconnected: make(chan error),
	}

	if sdkKey == "" {
		config.Logger.Errorf("%s Initialization failed: SDK Key cannot be empty. Please provide a valid SDK Key to initialize the client.", sdk_codes.InitMissingKey)
		return client, EmptySDKKeyError
	}

	var err error

	lruCache, err := repository.NewLruCache(10000)
	if err != nil {
		return nil, err
	}
	client.repository = repository.New(lruCache)

	if client.config != nil {
		client.repository = repository.New(config.Cache)
	}

	client.evaluator, err = evaluation.NewEvaluator(client.repository, client, config.Logger)
	if err != nil {
		return nil, err
	}

	client.start()
	if config.waitForInitialized {
		config.Logger.Infof("%s The SDK is waiting for initialization to complete'", sdk_codes.InitWaiting)

		var initErr error

		select {
		case <-client.initialized:
			config.Logger.Infof("%s The SDK has successfully initialized", sdk_codes.InitSuccess)
			return client, nil
		case err := <-client.initializedErr:
			initErr = err
		}

		if initErr != nil {
			config.Logger.Errorf("Initialization failed: '%v'", initErr)
			// We return the client but leave it in un-initialized state by not setting the relevant initialized flag.
			// This ensures any subsequent calls to the client don't potentially result in a panic. For example, if a user
			// calls BoolVariation we can log that the client is not initialized and return the user the default variation.
			return client, initErr
		}
	}

	return client, nil
}

func (c *CfClient) start() {

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-c.stop
		cancel()
	}()

	go func() {
		if err := c.initAuthentication(context.Background()); err != nil {
			c.config.Logger.Errorf("%s The SDK has failed to initialize due to an authentication error:  %v' ", sdk_codes.InitAuthError, err)
			c.initializedErr <- err
		}
	}()
	go c.setAnalyticsServiceClient(ctx)
	go c.pullCronJob(ctx)
	if c.config.enableStream {
		go c.stream(ctx)
	}
}

// PostEvaluateProcessor push the data to the analytics service
func (c *CfClient) PostEvaluateProcessor(data *evaluation.PostEvalData) {
	c.analyticsService.PushToQueue(data.FeatureConfig, data.Target, data.Variation)
}

// IsStreamConnected determines if the stream is currently connected
func (c *CfClient) IsStreamConnected() bool {
	return c.streamConnectedBool
}

// GetClusterIdentifier returns the cluster identifier we're connected to
func (c *CfClient) GetClusterIdentifier() string {
	return c.clusterIdentifier
}

// IsInitialized determines if the client is ready to be used.  This is true if it has both authenticated
// and successfully retrieved flags.  If it takes longer than 1 minute the call will timeout and return an error.
func (c *CfClient) IsInitialized() (bool, error) {
	for i := 0; i < 30; i++ {
		c.initializedBoolLock.RLock()
		if c.initializedBool {
			c.initializedBoolLock.RUnlock()
			return true, nil
		}
		c.initializedBoolLock.RUnlock()
		c.config.sleeper.Sleep(time.Second * 2)
	}
	return false, InitializeTimeoutError{}
}

func (c *CfClient) retrieve(ctx context.Context) {
	var g errgroup.Group

	rCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	// First goroutine for retrieving flags.
	g.Go(func() error {
		err := c.retrieveFlags(rCtx)
		if err != nil {
			c.config.Logger.Errorf("error while retrieving flags: %v", err)
			return err
		}
		return nil
	})

	// Second goroutine for retrieving segments.
	g.Go(func() error {
		err := c.retrieveSegments(rCtx)
		if err != nil {
			c.config.Logger.Errorf("error while retrieving segments: %v", err)
			return err
		}
		return nil
	})

	err := g.Wait()

	if err != nil {
		// We just log the error and continue. In the case of initialization, this means we mark the client as initialized
		// if we can't poll for initial state, and default evaluations are likely to be returned.
		c.config.Logger.Errorf("Data poll finished with errors: %s", err)
	} else {
		c.config.Logger.Info("Data poll finished successfully")
	}

	c.initializedBoolLock.Lock()
	defer c.initializedBoolLock.Unlock()

	// This function is used to mark the client as "initialized" once flags and segments have been loaded,
	// but it's also used for the polling thread, so we check if the client is already initialized before
	// marking it as such.
	if !c.initializedBool {
		c.initializedBool = true
		close(c.initialized)
	}
}

func (c *CfClient) streamConnect(ctx context.Context) {
	// we only ever want one stream to be setup - other threads must wait before trying to establish a connection
	c.streamConnectedBoolLock.Lock()
	defer c.streamConnectedBoolLock.Unlock()
	if !c.config.enableStream || c.streamConnectedBool {
		return
	}

	<-c.authenticated

	c.mux.RLock()
	defer c.mux.RUnlock()
	sseClient := sse.NewClient(fmt.Sprintf("%s/stream?cluster=%s", c.config.url, c.clusterIdentifier))

	// Use the SDKs http client
	sseClient.Connection = c.config.httpClient

	conn := stream.NewSSEClient(c.sdkKey, c.token, sseClient, c.repository, c.api, c.config.Logger,
		c.config.eventStreamListener, c.config.proxyMode, c.streamConnected, c.streamDisconnected)

	// Connect kicks off a goroutine that attempts to establish a stream connection
	// while this is happening we set streamConnectedBool to true - if any errors happen
	// in this process streamConnectedBool will be set back to false by the streamDisconnected function
	conn.Connect(ctx, c.environmentID, c.sdkKey)
	c.streamConnectedBool = true
}

func (c *CfClient) initAuthentication(ctx context.Context) error {

	// Variable to count the number of attempts.
	var attempts int

	// Define the operation to be retried.
	operation := func() error {
		err := c.authenticate(ctx)
		if err == nil {
			c.config.Logger.Infof("%s Authenticated successfully'", sdk_codes.AuthSuccess)
			return nil
		}

		var nonRetryableAuthError NonRetryableAuthError
		if errors.As(err, &nonRetryableAuthError) {
			c.config.Logger.Error("%s Authentication failed with a non-retryable error: '%s %s'. Default variations will now be served.", sdk_codes.AuthFailed, nonRetryableAuthError.StatusCode, nonRetryableAuthError.Message)
			return backoff.Permanent(err)
		}

		attempts++

		// If the error is retryable, check if we've exceeded the max retries.
		if c.config.maxAuthRetries != -1 && attempts >= c.config.maxAuthRetries {
			c.config.Logger.Errorf("%s Authentication failed with error: '%s'. Exceeded max attempts: '%v'.", sdk_codes.AuthExceededRetries, err, c.config.maxAuthRetries)
			return backoff.Permanent(err) // Making this error non-retryable.
		}

		return err
	}

	retryStrategy := backoff.WithContext(c.config.authRetryStrategy, ctx)

	notify := func(err error, duration time.Duration) {
		c.config.Logger.Warnf("%s Authentication attempt %d failed with error: '%s'. Retrying in %v.", sdk_codes.AuthAttempt, attempts, err, duration)
	}

	err := backoff.RetryNotify(operation, retryStrategy, notify)

	if err != nil {
		// Handle the case where the operation has failed after all retries.
		c.config.Logger.Errorf("%s Authentication failed after %d attempts: '%s'.", sdk_codes.AuthExceededRetries, attempts, err)
	}

	return err
}

func (c *CfClient) authenticate(ctx context.Context) error {
	c.mux.RLock()
	defer c.mux.RUnlock()

	// dont check err just retry
	httpClient, err := rest.NewClientWithResponses(c.config.url, rest.WithHTTPClient(c.config.authHttpClient))
	if err != nil {
		return err
	}

	response, err := httpClient.AuthenticateWithResponse(ctx, rest.AuthenticateJSONRequestBody{
		ApiKey: c.sdkKey,
		Target: c.auth.Target,
	})
	if err != nil {
		return err
	}

	responseError := findErrorInResponse(response)

	// Indicate that we should retry
	if responseError != nil && responseError.Code == "500" {
		return RetryableAuthError{
			StatusCode: responseError.Code,
			Message:    responseError.Message,
		}
	}

	// Indicate that we shouldn't retry on non-500 errors
	if responseError != nil {
		return NonRetryableAuthError{
			StatusCode: responseError.Code,
			Message:    responseError.Message,
		}
	}

	// Defensive check to handle the case that all responses are nil
	if response.JSON200 == nil {
		return RetryableAuthError{
			StatusCode: "No error status code returned from server",
			Message:    "No error message returned from server ",
		}
	}

	c.token = response.JSON200.AuthToken
	// initialize client go for communicating to ff-server
	payloadIndex := 1
	payload := strings.Split(c.token, ".")[payloadIndex]
	payloadData, err := jwt.DecodeSegment(payload)
	if err != nil {
		return err
	}

	var claims map[string]interface{}
	if err = json.Unmarshal(payloadData, &claims); err != nil {
		return err
	}

	var ok bool
	c.environmentID, ok = claims["environment"].(string)
	if !ok {
		return fmt.Errorf("environment uuid not present")
	}

	c.clusterIdentifier, ok = claims["clusterIdentifier"].(string)
	if !ok {
		c.clusterIdentifier = "1"
		return fmt.Errorf("cluster identifier not present")
	}

	// network layer setup
	bearerTokenProvider, bearerTokenProviderErr := securityprovider.NewSecurityProviderBearerToken(c.token)
	if bearerTokenProviderErr != nil {
		return bearerTokenProviderErr
	}
	restClient, err := rest.NewClientWithResponses(c.config.url,
		rest.WithRequestEditorFn(bearerTokenProvider.Intercept),
		rest.WithRequestEditorFn(c.InterceptAddCluster),
		rest.WithHTTPClient(c.config.httpClient),
	)
	if err != nil {
		return err
	}

	metricsClient, err := metricsclient.NewClientWithResponses(c.config.eventsURL,
		metricsclient.WithRequestEditorFn(bearerTokenProvider.Intercept),
		metricsclient.WithRequestEditorFn(c.InterceptAddCluster),
		metricsclient.WithHTTPClient(c.config.httpClient),
	)
	if err != nil {
		return err
	}

	c.api = restClient
	c.metricsapi = metricsClient
	c.config.Logger.Info("Authentication complete")
	close(c.authenticated)
	return nil
}

func (c *CfClient) makeTicker(interval uint) *time.Ticker {
	return time.NewTicker(time.Second * time.Duration(interval))
}

func (c *CfClient) stream(ctx context.Context) {
	// wait until initialized with initial state
	<-c.initialized
	c.config.Logger.Infof("%s Polling Stopped", sdk_codes.PollStop)
	c.streamConnect(ctx)

	streamingRetryStrategy := c.config.streamingRetryStrategy
	// Create an initial ticker to handle the case where we don't open a succesful connection on our first attempt
	ticker := backoff.NewTicker(streamingRetryStrategy)

	defer ticker.Stop()
	reconnectionAttempt := 1
	for {
		select {
		case <-ctx.Done():
			c.config.Logger.Infof("%s Stream stopped", sdk_codes.StreamStop)
			if ticker != nil {
				ticker.Stop()
			}
			return

		case <-c.streamConnected:
			c.config.Logger.Infof("%s Stream successfully connected", sdk_codes.StreamStarted)
			// Reset the reconnection attempt and ticker
			reconnectionAttempt = 1
			streamingRetryStrategy.Reset()
			ticker.Stop()
			ticker = nil

		case err := <-c.streamDisconnected:
			c.config.Logger.Warnf("%s Stream disconnected: '%s' Swapping to polling mode", sdk_codes.StreamDisconnected, err)
			c.mux.RLock()
			c.streamConnectedBool = false
			c.mux.RUnlock()

			// If an eventStreamListener has been passed to the Proxy lets notify it of the disconnected
			// to let it know something is up with the stream it has been listening to
			if c.config.eventStreamListener != nil {
				c.config.eventStreamListener.Pub(context.Background(), stream.Event{
					APIKey:      c.sdkKey,
					Environment: c.environmentID,
					Err:         stream.ErrStreamDisconnect,
				})
			}

			if ticker == nil {
				ticker = backoff.NewTicker(streamingRetryStrategy)
			}

			c.config.Logger.Infof("%s Stream connection lost: '%v' Retrying in %fs (attempt %d)", sdk_codes.StreamRetry, err, streamingRetryStrategy.NextBackOff().Seconds(), reconnectionAttempt)
			reconnectionAttempt += 1

			// Backoff before retrying
			select {
			case <-ticker.C:
			case <-ctx.Done():
				c.config.Logger.Infof("%s Stream stopped during reconnection", sdk_codes.StreamStop)
				ticker.Stop()
				return
			}
			c.streamConnect(ctx)
		}
	}
}

func (c *CfClient) pullCronJob(ctx context.Context) {
	poll := func() {
		c.mux.RLock()
		defer c.mux.RUnlock()
		if !c.streamConnectedBool {
			c.retrieve(ctx)
		}
	}
	// wait until authenticated
	<-c.authenticated

	c.config.Logger.Infof("%s Polling started, interval: %v seconds", sdk_codes.PollStart, c.config.pullInterval)
	// pull initial data
	poll()

	// start cron
	pullingTicker := c.makeTicker(c.config.pullInterval)
	for {
		select {
		case <-ctx.Done():
			pullingTicker.Stop()
			c.config.Logger.Infof("%s Polling stopped", sdk_codes.PollStop)
			return
		case <-pullingTicker.C:
			poll()
		}
	}
}

func (c *CfClient) retrieveFlags(ctx context.Context) error {

	<-c.authenticated

	c.mux.RLock()
	defer c.mux.RUnlock()
	c.config.Logger.Info("Retrieving flags started")
	flags, err := c.api.GetFeatureConfigWithResponse(ctx, c.environmentID)
	if err != nil {
		// log
		return err
	}

	if flags.JSON200 == nil {
		return fmt.Errorf("%w: `%v`", FetchFlagsError, flags.HTTPResponse.Status)
	}

	c.repository.SetFlags(true, c.environmentID, *flags.JSON200...)
	for _, flag := range *flags.JSON200 {
		c.repository.SetFlag(flag, true)
	}
	c.config.Logger.Info("Retrieving flags finished")
	return nil
}

func (c *CfClient) retrieveSegments(ctx context.Context) error {

	<-c.authenticated

	c.mux.RLock()
	defer c.mux.RUnlock()
	c.config.Logger.Info("Retrieving segments started")
	segments, err := c.api.GetAllSegmentsWithResponse(ctx, c.environmentID)
	if err != nil {
		// log
		return err
	}

	if segments.JSON200 == nil {
		return nil
	}

	c.repository.SetSegments(true, c.environmentID, *segments.JSON200...)
	for _, segment := range *segments.JSON200 {
		c.repository.SetSegment(segment, true)
	}
	c.config.Logger.Info("Retrieving segments finished")
	return nil
}

func (c *CfClient) setAnalyticsServiceClient(ctx context.Context) {

	<-c.authenticated
	c.mux.RLock()
	defer c.mux.RUnlock()
	if !c.config.enableAnalytics {
		c.config.Logger.Info("Posting analytics data disabled")
		return
	}
	c.config.Logger.Info("Posting analytics data enabled")
	c.analyticsService.Start(ctx, &c.metricsapi, c.environmentID)
}

// BoolVariation returns the value of a boolean feature flag for a given target.
// Returns defaultValue if there is an error or if the flag doesn't exist
func (c *CfClient) BoolVariation(key string, target *evaluation.Target, defaultValue bool) (bool, error) {
	if !c.initializedBool {
		c.config.Logger.Infof("%s Error while evaluating boolean flag and returning default variation: 'Client is not initialized'", sdk_codes.EvaluationFailed)
		return defaultValue, fmt.Errorf("%w: Client is not initialized", DefaultVariationReturnedError)
	}
	value, err := c.evaluator.BoolVariation(key, target, defaultValue)
	if err != nil {
		c.config.Logger.Infof("%s Error while evaluating boolean flag and returning default variation '%s', err: %v", sdk_codes.EvaluationFailed, key, err)
		return value, fmt.Errorf("%w: `%v`", DefaultVariationReturnedError, err)
	}
	c.config.Logger.Debugf("%s Evaluated boolean flag successfully: '%s'", sdk_codes.EvaluationSuccess, key)
	return value, nil
}

// StringVariation returns the value of a string feature flag for a given target.
//
// Returns defaultValue if there is an error or if the flag doesn't exist
func (c *CfClient) StringVariation(key string, target *evaluation.Target, defaultValue string) (string, error) {
	if !c.initializedBool {
		c.config.Logger.Infof("%s Error while evaluating string flag and returning default variation: 'Client is not initialized'", sdk_codes.EvaluationFailed)
		return defaultValue, fmt.Errorf("%w: Client is not initialized", DefaultVariationReturnedError)
	}
	value, err := c.evaluator.StringVariation(key, target, defaultValue)
	if err != nil {
		c.config.Logger.Infof("%s Error while evaluating string flag '%s', err: %v", sdk_codes.EvaluationFailed, key, err)
		return value, fmt.Errorf("%w: `%v`", DefaultVariationReturnedError, err)
	}
	c.config.Logger.Debugf("%s Evaluated string flag successfully: '%s'", sdk_codes.EvaluationSuccess, key)
	return value, nil
}

// IntVariation returns the value of a integer feature flag for a given target.
//
// Returns defaultValue if there is an error or if the flag doesn't exist
func (c *CfClient) IntVariation(key string, target *evaluation.Target, defaultValue int64) (int64, error) {
	if !c.initializedBool {
		c.config.Logger.Infof("%s Error while evaluating int flag and returning default variation: 'Client is not initialized'", sdk_codes.EvaluationFailed)
		return defaultValue, fmt.Errorf("%w: Client is not initialized", DefaultVariationReturnedError)
	}
	value, err := c.evaluator.IntVariation(key, target, int(defaultValue))
	if err != nil {
		c.config.Logger.Infof("%s Error while evaluating int flag '%s', err: %v", sdk_codes.EvaluationFailed, key, err)
		return int64(value), fmt.Errorf("%w: `%v`", DefaultVariationReturnedError, err)
	}
	c.config.Logger.Debugf("%s Evaluated int flag successfully: '%s'", sdk_codes.EvaluationSuccess, key)
	return int64(value), nil
}

// NumberVariation returns the value of a float64 feature flag for a given target.
//
// Returns defaultValue if there is an error or if the flag doesn't exist
func (c *CfClient) NumberVariation(key string, target *evaluation.Target, defaultValue float64) (float64, error) {
	if !c.initializedBool {
		c.config.Logger.Infof("%s Error while number number flag and returning default variation: 'Client is not initialized'", sdk_codes.EvaluationFailed)
		return defaultValue, fmt.Errorf("%w: Client is not initialized", DefaultVariationReturnedError)
	}
	value, err := c.evaluator.NumberVariation(key, target, defaultValue)
	if err != nil {
		c.config.Logger.Infof("%s Error while evaluating number flag '%s', err: %v", sdk_codes.EvaluationFailed, key, err)
		return value, fmt.Errorf("%w: `%v`", DefaultVariationReturnedError, err)
	}
	c.config.Logger.Debugf("%s Evaluated number flag successfully: '%s'", sdk_codes.EvaluationSuccess, key)
	return value, nil
}

// JSONVariation returns the value of a feature flag for the given target, allowing the value to be
// of any JSON type.
//
// Returns defaultValue if there is an error or if the flag doesn't exist
func (c *CfClient) JSONVariation(key string, target *evaluation.Target, defaultValue types.JSON) (types.JSON, error) {
	if !c.initializedBool {
		c.config.Logger.Infof("%s Error while evaluating json flag and returning default variation: 'Client is not initialized'", sdk_codes.EvaluationFailed)
		return defaultValue, fmt.Errorf("%w: Client is not initialized", DefaultVariationReturnedError)
	}
	value, err := c.evaluator.JSONVariation(key, target, defaultValue)
	if err != nil {
		c.config.Logger.Infof("%s Error while evaluating json flag '%s', err: %v", sdk_codes.EvaluationFailed, key, err)
		return value, fmt.Errorf("%w: `%v`", DefaultVariationReturnedError, err)
	}
	c.config.Logger.Debugf("%s Evaluated json flag successfully: '%s'", sdk_codes.EvaluationSuccess, key)
	return value, nil
}

// Close shuts down the Feature Flag client. After calling this, the client
// should no longer be used
func (c *CfClient) Close() error {
	if !c.initializedBool {
		return errors.New("attempted to close client that is not initialized")
	}
	if c.stopped.get() {
		return errors.New("client already closed")
	}
	c.config.Logger.Infof("%s Closing SDK", sdk_codes.CloseStarted)
	close(c.stop)

	c.stopped.set(true)

	// This flag is used by `IsInitialized` so set to true.
	c.initializedBoolLock.Lock()
	c.initializedBool = false
	c.initializedBoolLock.Unlock()
	c.config.Logger.Infof("%s SDK Closed successfully", sdk_codes.CloseSuccess)

	return nil
}

// Environment returns environment based on authenticated SDK flagIdentifier
func (c *CfClient) Environment() string {
	return c.environmentID
}

// InterceptAddCluster adds cluster ID to calls
func (c *CfClient) InterceptAddCluster(ctx context.Context, req *http.Request) error {
	q := req.URL.Query()
	q.Add("cluster", c.clusterIdentifier)
	req.URL.RawQuery = q.Encode()
	return nil
}

type atomicBool struct {
	flag int32
}

func newAtomicBool(value bool) *atomicBool {
	b := new(atomicBool)
	b.set(value)
	return b
}

func (a *atomicBool) set(value bool) {
	var i int32 = 0
	if value {
		i = 1
	}
	atomic.StoreInt32(&(a.flag), i)
}

func (a *atomicBool) get() bool {
	return atomic.LoadInt32(&(a.flag)) != int32(0)
}

// getLogger returns either the custom passed in logger or our default zap logger
func getLogger(options ...ConfigOption) logger.Logger {
	dummyConfig := &config{}
	for _, opt := range options {
		opt(dummyConfig)
	}
	if dummyConfig.Logger == nil {
		defaultLogger, err := logger.NewZapLogger(false)
		if err != nil {
			log.Printf("Error creating zap logger instance, %v", err)
		}
		dummyConfig.Logger = defaultLogger
	}
	return dummyConfig.Logger
}

// findErrorInResponse parses an auth response and returns the response error if it exists
func findErrorInResponse(resp *rest.AuthenticateResponse) *rest.Error {
	responseErrors := []*rest.Error{resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500}
	for _, responseError := range responseErrors {
		if responseError != nil {
			return responseError
		}
	}
	return nil
}
