package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/harness/ff-golang-server-sdk/analyticsservice"
	"github.com/harness/ff-golang-server-sdk/metricsclient"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/golang-jwt/jwt"
	"github.com/harness/ff-golang-server-sdk/cache"
	"github.com/harness/ff-golang-server-sdk/dto"
	"github.com/harness/ff-golang-server-sdk/evaluation"
	"github.com/harness/ff-golang-server-sdk/rest"
	"github.com/harness/ff-golang-server-sdk/stream"
	"github.com/harness/ff-golang-server-sdk/types"

	"github.com/r3labs/sse"
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
//
type CfClient struct {
	mux                 sync.RWMutex
	api                 rest.ClientWithResponsesInterface
	metricsapi          metricsclient.ClientWithResponsesInterface
	sdkKey              string
	auth                rest.AuthenticationRequest
	config              *config
	environmentID       string
	token               string
	persistence         cache.Persistence
	cancelFunc          context.CancelFunc
	streamConnected     bool
	streamConnectedLock sync.RWMutex
	authenticated       chan struct{}
	initialized         bool
	initializedLock     sync.RWMutex
	analyticsService    *analyticsservice.AnalyticsService
	clusterIdentifier   string
}

// NewCfClient creates a new client instance that connects to CF with the default configuration.
// For advanced configuration options use ConfigOptions functions
func NewCfClient(sdkKey string, options ...ConfigOption) (*CfClient, error) {

	var (
		ctx context.Context
		err error
	)

	//  functional options for config
	config := newDefaultConfig()
	for _, opt := range options {
		opt(config)
	}

	analyticsService := analyticsservice.NewAnalyticsService(time.Minute, config.Logger)

	client := &CfClient{
		sdkKey:            sdkKey,
		config:            config,
		authenticated:     make(chan struct{}),
		analyticsService:  analyticsService,
		clusterIdentifier: "1",
	}
	ctx, client.cancelFunc = context.WithCancel(context.Background())

	if sdkKey == "" {
		return client, types.ErrSdkCantBeEmpty
	}

	client.persistence = cache.NewPersistence(config.Store, config.Cache, config.Logger)
	// load from storage
	if config.enableStore {
		if err = client.persistence.LoadFromStore(); err != nil {
			config.Logger.Errorf("error loading from store err: %s", err.Error())
		}
	}

	go client.initAuthentication(ctx, client.config.target)

	go client.setAnalyticsServiceClient(ctx)

	go client.pullCronJob(ctx)

	go client.persistCronJob(ctx)

	return client, nil
}

// IsStreamConnected determines if the stream is currently connected
func (c *CfClient) IsStreamConnected() bool {
	return c.streamConnected
}

// GetClusterIdentifier returns the cluster identifier we're connected to
func (c *CfClient) GetClusterIdentifier() string {
	return c.clusterIdentifier
}

// IsInitialized determines if the client is ready to be used.  This is true if it has both authenticated
// and successfully retrieved flags.  If it takes longer than 1 minute the call will timeout and return an error.
func (c *CfClient) IsInitialized() (bool, error) {
	for i := 0; i < 30; i++ {
		c.initializedLock.RLock()
		if c.initialized {
			c.initializedLock.RUnlock()
			return true, nil
		}
		c.initializedLock.RUnlock()
		time.Sleep(time.Second * 2)
	}
	return false, fmt.Errorf("timeout waiting to initialize")
}

func (c *CfClient) retrieve(ctx context.Context) bool {
	ok := true
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		rCtx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		err := c.retrieveFlags(rCtx)
		if err != nil {
			ok = false
			c.config.Logger.Errorf("error while retrieving flags: %v", err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		rCtx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		err := c.retrieveSegments(rCtx)
		if err != nil {
			ok = false
			c.config.Logger.Errorf("error while retrieving segments: %v", err.Error())
		}
	}()
	wg.Wait()
	if ok {
		c.config.Logger.Info("Data poll finished successfully")
	} else {
		c.config.Logger.Error("Data poll finished with errors")
	}

	if ok {
		c.initializedLock.Lock()
		c.initialized = true
		c.initializedLock.Unlock()
	}
	return ok
}

func (c *CfClient) streamConnect(ctx context.Context) {
	// we only ever want one stream to be setup - other threads must wait before trying to establish a connection
	c.streamConnectedLock.Lock()
	defer c.streamConnectedLock.Unlock()
	if !c.config.enableStream || c.streamConnected {
		return
	}

	<-c.authenticated

	c.mux.RLock()
	defer c.mux.RUnlock()
	c.config.Logger.Info("Registering SSE consumer")
	sseClient := sse.NewClient(fmt.Sprintf("%s/stream?cluster=%s", c.config.url, c.clusterIdentifier))

	streamErr := func() {
		c.config.Logger.Error("Stream disconnected. Swapping to polling mode")
		c.mux.RLock()
		defer c.mux.RUnlock()
		c.streamConnected = false
	}
	conn := stream.NewSSEClient(c.sdkKey, c.token, sseClient, c.config.Cache, c.api, c.config.Logger, streamErr, c.config.eventStreamListener)

	// Connect kicks off a goroutine that attempts to establish a stream connection
	// while this is happening we set streamConnected to true - if any errors happen
	// in this process streamConnected will be set back to false by the streamErr function
	conn.Connect(ctx, c.environmentID, c.sdkKey)
	c.streamConnected = true
}

func (c *CfClient) initAuthentication(ctx context.Context, target evaluation.Target) {
	// attempt to authenticate every minute until we succeed
	for {
		err := c.authenticate(ctx, target)
		if err == nil {
			return
		}
		c.config.Logger.Errorf("Authentication failed. Trying again in 1 minute: %s", err)
		time.Sleep(1 * time.Minute)
	}
}

func (c *CfClient) authenticate(ctx context.Context, target evaluation.Target) error {
	t := struct {
		Anonymous  *bool                   `json:"anonymous,omitempty"`
		Attributes *map[string]interface{} `json:"attributes,omitempty"`
		Identifier string                  `json:"identifier"`
		Name       *string                 `json:"name,omitempty"`
	}{
		target.Anonymous,
		target.Attributes,
		target.Identifier,
		&target.Name,
	}
	c.auth.Target = &t
	c.mux.RLock()
	defer c.mux.RUnlock()

	// dont check err just retry
	httpClient, err := rest.NewClientWithResponses(c.config.url, rest.WithHTTPClient(c.config.httpClient))
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
	// should be login to harness and get account data (JWT token)
	if response.JSON200 == nil {
		return fmt.Errorf("error while authenticating %v", ErrUnauthorized)
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

	merticsClient, err := metricsclient.NewClientWithResponses(c.config.eventsURL,
		metricsclient.WithRequestEditorFn(bearerTokenProvider.Intercept),
		metricsclient.WithHTTPClient(http.DefaultClient),
	)
	if err != nil {
		return err
	}

	c.api = restClient
	c.metricsapi = merticsClient
	c.config.Logger.Info("Authentication complete")
	close(c.authenticated)
	return nil
}

func (c *CfClient) makeTicker(interval uint) *time.Ticker {
	return time.NewTicker(time.Minute * time.Duration(interval))
}

func (c *CfClient) pullCronJob(ctx context.Context) {
	poll := func() {
		c.mux.RLock()
		if !c.streamConnected {
			ok := c.retrieve(ctx)
			// we should only try and start the stream after the poll succeeded to make sure we get the latest changes
			if ok && c.config.enableStream {
				// here stream is enabled but not connected, so we attempt to reconnect
				c.config.Logger.Info("Attempting to start stream")
				c.streamConnect(ctx)
			}
		}
		c.mux.RUnlock()
	}
	// wait until authenticated
	<-c.authenticated

	// pull initial data
	poll()

	// start cron
	pullingTicker := c.makeTicker(c.config.pullInterval)
	for {
		select {
		case <-ctx.Done():
			pullingTicker.Stop()
			return
		case <-pullingTicker.C:
			poll()
		}
	}
}

func (c *CfClient) persistCronJob(ctx context.Context) {
	// if store is disabled don't setup the cron job
	if !c.config.enableStore {
		return
	}
	persistingTicker := c.makeTicker(1)
	for {
		select {
		case <-ctx.Done():
			persistingTicker.Stop()
			return
		case <-persistingTicker.C:
			err := c.persistence.SaveToStore()
			if err != nil {
				c.config.Logger.Errorf("error while persisting data, err: %v", err)
			}
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
		return nil
	}

	for _, flag := range *flags.JSON200 {
		c.config.Cache.Set(dto.Key{
			Type: dto.KeyFeature,
			Name: flag.Feature,
		}, *flag.Convert()) // dereference for holding object in cache instead of address
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

	for _, segment := range *segments.JSON200 {
		c.config.Cache.Set(dto.Key{
			Type: dto.KeySegment,
			Name: segment.Identifier,
		}, segment.Convert())
	}
	c.config.Logger.Info("Retrieving segments finished")
	return nil
}

func (c *CfClient) getFlagFromCache(key string) *evaluation.FeatureConfig {
	value, _ := c.config.Cache.Get(dto.Key{
		Type: dto.KeyFeature,
		Name: key,
	})
	fc, ok := value.(evaluation.FeatureConfig)
	if ok {
		return &fc
	}
	return nil
}

func (c *CfClient) getSegmentsFromCache(fc *evaluation.FeatureConfig) {
	segments := fc.GetSegmentIdentifiers()
	for _, segmentIdentifier := range segments {
		value, _ := c.config.Cache.Get(dto.Key{
			Type: dto.KeySegment,
			Name: segmentIdentifier,
		})
		segment, ok := value.(evaluation.Segment)
		if ok {
			if fc.Segments == nil {
				fc.Segments = make(map[string]*evaluation.Segment)
			}
			fc.Segments[segmentIdentifier] = &segment
		}
	}
}

func (c *CfClient) setAnalyticsServiceClient(ctx context.Context) {
	<-c.authenticated
	c.analyticsService.Start(ctx, &c.metricsapi, c.environmentID)
}

// BoolVariation returns the value of a boolean feature flag for a given target.
//
// Returns defaultValue if there is an error or if the flag doesn't exist
func (c *CfClient) BoolVariation(key string, target *evaluation.Target, defaultValue bool) (bool, error) {
	fc := c.getFlagFromCache(key)
	if fc != nil {
		// load segments dep
		c.getSegmentsFromCache(fc)

		result := checkPreRequisite(c, fc, target)
		if !result {
			return fc.Variations.FindByIdentifier(fc.OffVariation).Bool(defaultValue), nil
		}
		variation, err := fc.BoolVariation(target)
		if err != nil {
			return defaultValue, err
		}
		c.analyticsService.PushToQueue(target, fc, variation)

		return variation.Bool(defaultValue), nil
	}
	return defaultValue, nil
}

// StringVariation returns the value of a string feature flag for a given target.
//
// Returns defaultValue if there is an error or if the flag doesn't exist
func (c *CfClient) StringVariation(key string, target *evaluation.Target, defaultValue string) (string, error) {
	fc := c.getFlagFromCache(key)
	if fc != nil {
		// load segments dep
		c.getSegmentsFromCache(fc)

		result := checkPreRequisite(c, fc, target)
		if !result {
			return fc.Variations.FindByIdentifier(fc.OffVariation).String(defaultValue), nil
		}
		variation, err := fc.StringVariation(target)
		if err != nil {
			return defaultValue, err
		}

		c.analyticsService.PushToQueue(target, fc, variation)

		return variation.String(defaultValue), nil
	}
	return defaultValue, nil
}

// IntVariation returns the value of a integer feature flag for a given target.
//
// Returns defaultValue if there is an error or if the flag doesn't exist
func (c *CfClient) IntVariation(key string, target *evaluation.Target, defaultValue int64) (int64, error) {
	fc := c.getFlagFromCache(key)
	if fc != nil {
		// load segments dep
		c.getSegmentsFromCache(fc)

		result := checkPreRequisite(c, fc, target)
		if !result {
			return fc.Variations.FindByIdentifier(fc.OffVariation).Int(defaultValue), nil
		}
		variation, err := fc.IntVariation(target)
		if err != nil {
			return defaultValue, err
		}

		c.analyticsService.PushToQueue(target, fc, variation)

		return variation.Int(defaultValue), nil
	}
	return defaultValue, nil
}

// NumberVariation returns the value of a float64 feature flag for a given target.
//
// Returns defaultValue if there is an error or if the flag doesn't exist
func (c *CfClient) NumberVariation(key string, target *evaluation.Target, defaultValue float64) (float64, error) {
	fc := c.getFlagFromCache(key)
	if fc != nil {
		// load segments dep
		c.getSegmentsFromCache(fc)

		result := checkPreRequisite(c, fc, target)
		if !result {
			return fc.Variations.FindByIdentifier(fc.OffVariation).Number(defaultValue), nil
		}
		variation, err := fc.NumberVariation(target)
		if err != nil {
			return defaultValue, err
		}

		c.analyticsService.PushToQueue(target, fc, variation)

		return variation.Number(defaultValue), nil
	}
	return defaultValue, nil
}

// JSONVariation returns the value of a feature flag for the given target, allowing the value to be
// of any JSON type.
//
// Returns defaultValue if there is an error or if the flag doesn't exist
func (c *CfClient) JSONVariation(key string, target *evaluation.Target, defaultValue types.JSON) (types.JSON, error) {
	fc := c.getFlagFromCache(key)
	if fc != nil {
		// load segments dep
		c.getSegmentsFromCache(fc)

		result := checkPreRequisite(c, fc, target)
		if !result {
			return fc.Variations.FindByIdentifier(fc.OffVariation).JSON(defaultValue), nil
		}
		variation, err := fc.JSONVariation(target)
		if err != nil {
			return defaultValue, err
		}
		c.analyticsService.PushToQueue(target, fc, variation)

		return variation.JSON(defaultValue), err
	}
	return defaultValue, nil
}

// Close shuts down the Feature Flag client. After calling this, the client
// should no longer be used
func (c *CfClient) Close() error {
	err := c.persistence.SaveToStore()
	if err != nil {
		return err
	}
	c.cancelFunc()
	return nil
}

// Environment returns environment based on authenticated SDK key
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

// contains determines if the string variation is in the slice of variations.
// returns true if found, otherwise false.
func contains(variations []string, variation string) bool {
	for _, x := range variations {
		if x == variation {
			return true
		}
	}
	return false
}

func checkPreRequisite(client *CfClient, featureConfig *evaluation.FeatureConfig, target *evaluation.Target) bool {
	result := true

	for _, preReq := range featureConfig.Prerequisites {
		preReqFeature := client.getFlagFromCache(preReq.Feature)
		if preReqFeature == nil {
			client.config.Logger.Errorf("Could not retrieve the pre requisite details of feature flag :[%s]", preReq.Feature)
			continue
		}

		// Get Variation (this performs evaluation and returns the current variation to be served to this target)
		preReqVariationName := preReqFeature.GetVariationName(target)
		preReqVariation := preReqFeature.Variations.FindByIdentifier(preReqVariationName)
		if preReqVariation == nil {
			client.config.Logger.Infof("Could not retrieve the pre requisite variation: %s", preReqVariationName)
			continue
		}
		client.config.Logger.Debugf("Pre requisite flag %s has variation %s for target %s", preReq.Feature, preReqVariation.Value, target.Identifier)

		if !contains(preReq.Variations, preReqVariation.Value) {
			return false
		}

		// Check this pre-requisites, own pre-requisite.  If we get a false anywhere we need to stop
		result = checkPreRequisite(client, preReqFeature, target)
		if !result {
			return false
		}
	}

	return result
}
