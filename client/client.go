package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/dgrijalva/jwt-go"
	"github.com/drone/ff-golang-server-sdk.v0/cache"
	"github.com/drone/ff-golang-server-sdk.v0/dto"
	"github.com/drone/ff-golang-server-sdk.v0/evaluation"
	"github.com/drone/ff-golang-server-sdk.v0/rest"
	"github.com/drone/ff-golang-server-sdk.v0/stream"
	"github.com/drone/ff-golang-server-sdk.v0/types"
	"github.com/hashicorp/go-retryablehttp"
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
	mux             sync.RWMutex
	api             rest.ClientWithResponsesInterface
	sdkKey          string
	config          *config
	environmentID   string
	token           string
	persistence     cache.Persistence
	cancelFunc      context.CancelFunc
	streamConnected bool
	authenticated   chan struct{}
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

	client := &CfClient{
		sdkKey:        sdkKey,
		config:        config,
		authenticated: make(chan struct{}),
	}
	ctx, client.cancelFunc = context.WithCancel(context.Background())

	if sdkKey == "" {
		return client, types.ErrSdkCantBeEmpty
	}

	client.persistence = cache.NewPersistence(config.Store, config.Cache, config.Logger)
	// load from storage
	if err = client.persistence.LoadFromStore(); err != nil {
		log.Printf("error loading from store err: %s", err)
	}

	go client.authenticate(ctx)

	go client.retrieve(ctx)

	go client.streamConnect()

	go client.pullCronJob(ctx)

	go client.persistCronJob(ctx)

	return client, nil
}

func (c *CfClient) retrieve(ctx context.Context) {
	// check for first cycle of cron job
	// for registering stream consumer
	c.config.Logger.Info("Pooling")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := c.retrieveFlags(ctx)
		if err != nil {
			log.Printf("error while retreiving flags at startup: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		err := c.retrieveSegments(ctx)
		if err != nil {
			log.Printf("error while retreiving segments at startup: %v", err)
		}
	}()
	wg.Wait()
	c.config.Logger.Info("Sync run finished")
}

func (c *CfClient) streamConnect() {
	if !c.config.enableStream {
		return
	}

	<-c.authenticated

	c.mux.RLock()
	defer c.mux.RUnlock()
	c.config.Logger.Info("Registering SSE consumer")
	sseClient := sse.NewClient(fmt.Sprintf("%s/stream", c.config.url))
	conn := stream.NewSSEClient(c.sdkKey, c.token, sseClient, c.config.Cache, c.api)
	err := conn.Connect(c.environmentID)
	if err != nil {
		c.streamConnected = false
		return
	}

	c.streamConnected = true
	err = conn.OnDisconnect(func() error {
		c.mux.RLock()
		defer c.mux.RUnlock()
		c.streamConnected = false
		return nil
	})
	if err != nil {
		c.config.Logger.Errorf("error disconnecting the stream, err: %v", err)
	}
}

func (c *CfClient) authenticate(ctx context.Context) {

	c.mux.RLock()
	defer c.mux.RUnlock()

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 10

	// dont check err just retry
	httpClient, err := rest.NewClientWithResponses(c.config.url, rest.WithHTTPClient(retryClient.StandardClient()))
	if err != nil {
		c.config.Logger.Error(err)
		return
	}

	response, err := httpClient.AuthenticateWithResponse(ctx, rest.AuthenticateJSONRequestBody{
		ApiKey: c.sdkKey,
	})
	if err != nil {
		c.config.Logger.Error(err)
		return
	}
	// should be login to harness and get account data (JWT token)
	if response.JSON200 == nil {
		c.config.Logger.Errorf("error while authenticating %v", ErrUnauthorized)
		return
	}

	c.token = response.JSON200.AuthToken
	// initialize client go for communicating to ff-server
	payloadIndex := 1
	payload := strings.Split(c.token, ".")[payloadIndex]
	payloadData, err := jwt.DecodeSegment(payload)
	if err != nil {
		c.config.Logger.Error(err)
		return
	}

	var claims map[string]interface{}
	if err = json.Unmarshal(payloadData, &claims); err != nil {
		c.config.Logger.Error(err)
		return
	}

	var ok bool
	c.environmentID, ok = claims["environment"].(string)
	if !ok {
		c.config.Logger.Error(errors.New("environment uuid not present"))
		return
	}

	// network layer setup
	bearerTokenProvider, bearerTokenProviderErr := securityprovider.NewSecurityProviderBearerToken(c.token)
	if bearerTokenProviderErr != nil {
		c.config.Logger.Error(bearerTokenProviderErr)
		return
	}
	restClient, err := rest.NewClientWithResponses(c.config.url,
		rest.WithRequestEditorFn(bearerTokenProvider.Intercept),
		rest.WithHTTPClient(retryClient.StandardClient()),
	)
	if err != nil {
		c.config.Logger.Error(err)
		return
	}

	c.api = restClient
	close(c.authenticated)
}

func (c *CfClient) makeTicker(interval uint) *time.Ticker {
	return time.NewTicker(time.Minute * time.Duration(interval))
}

func (c *CfClient) pullCronJob(ctx context.Context) {
	pullingTicker := c.makeTicker(c.config.pullInterval)
	for {
		select {
		case <-ctx.Done():
			pullingTicker.Stop()
			return
		case <-pullingTicker.C:
			c.mux.RLock()
			if !c.streamConnected {
				c.retrieve(ctx)
			}
			c.mux.RUnlock()
		}
	}
}

func (c *CfClient) persistCronJob(ctx context.Context) {
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

// BoolVariation returns the value of a boolean feature flag for a given target.
//
// Returns defaultValue if there is an error or if the flag doesn't exist
func (c *CfClient) BoolVariation(key string, target *evaluation.Target, defaultValue bool) (bool, error) {
	fc := c.getFlagFromCache(key)
	if fc != nil {
		// load segments dep
		c.getSegmentsFromCache(fc)
		return fc.BoolVariation(target, defaultValue), nil
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
		return fc.StringVariation(target, defaultValue), nil
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
		return fc.IntVariation(target, defaultValue), nil
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
		return fc.NumberVariation(target, defaultValue), nil
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
		return fc.JSONVariation(target, defaultValue), nil
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
