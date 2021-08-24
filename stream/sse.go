package stream

import (
	"context"
	"fmt"
	"time"

	"github.com/drone/ff-golang-server-sdk/cache"
	"github.com/drone/ff-golang-server-sdk/dto"
	"github.com/drone/ff-golang-server-sdk/logger"
	"github.com/drone/ff-golang-server-sdk/rest"

	jsoniter "github.com/json-iterator/go"
	"github.com/r3labs/sse"
)

// SSEClient is Server Send Event object
type SSEClient struct {
	api    rest.ClientWithResponsesInterface
	client *sse.Client
	cache  cache.Cache
	logger logger.Logger
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// NewSSEClient creates an object for stream interactions
func NewSSEClient(
	apiKey string,
	token string,
	client *sse.Client,
	cache cache.Cache,
	api rest.ClientWithResponsesInterface,
	logger logger.Logger,
) *SSEClient {
	client.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	client.Headers["API-Key"] = apiKey
	return &SSEClient{
		client: client,
		cache:  cache,
		api:    api,
		logger: logger,
	}
}

// Connect will subscribe to SSE stream
func (c *SSEClient) Connect(environment string) error {
	c.logger.Infof("Start subscribing to Stream")
	// it is blocking operation, it needs to go in go routine
	go func() {
		err := c.client.Subscribe("*", func(msg *sse.Event) {
			c.logger.Infof("Event received: %s", msg.Data)

			cfMsg := Message{}
			if len(msg.Data) > 0 {
				err := json.Unmarshal(msg.Data, &cfMsg)
				if err != nil {
					c.logger.Errorf("%s", err.Error())
					return
				}

				switch cfMsg.Domain {
				case dto.KeyFeature:
					// maybe is better to send event on memory bus that we get new message
					// and subscribe to that event
					switch cfMsg.Event {
					case dto.SseDeleteEvent:
						go func(identifier string) {
							c.cache.Remove(dto.Key{
								Type: dto.KeyFeature,
								Name: identifier,
							})
						}(cfMsg.Identifier)
					case dto.SsePatchEvent, dto.SseCreateEvent:
						fallthrough
					default:
						go func(env, identifier string) {
							ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
							defer cancel()
							response, err := c.api.GetFeatureConfigByIdentifierWithResponse(ctx, env, identifier)
							if err != nil {
								c.logger.Errorf("error while pulling flag, err: %s", err.Error())
								return
							}
							if response.JSON200 != nil {
								c.cache.Set(dto.Key{
									Type: dto.KeyFeature,
									Name: identifier,
								}, *response.JSON200.Convert())
							}
						}(environment, cfMsg.Identifier)
					}
				case dto.KeySegment:
					// need open client spec change
					switch cfMsg.Event {
					case dto.SseDeleteEvent:
						go func(identifier string) {
							c.cache.Remove(dto.Key{
								Type: dto.KeySegment,
								Name: identifier,
							})
						}(cfMsg.Identifier)
					case dto.SsePatchEvent, dto.SseCreateEvent:
						fallthrough
					default:
						go func(env, identifier string) {
							ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
							defer cancel()
							response, err := c.api.GetSegmentByIdentifierWithResponse(ctx, env, identifier)
							if err != nil {
								c.logger.Errorf("error while pulling segment, err: %s", err.Error())
								return
							}
							if response.JSON200 != nil {
								c.cache.Set(dto.Key{
									Type: dto.KeySegment,
									Name: identifier,
								}, response.JSON200.Convert())
							}
						}(environment, cfMsg.Identifier)
					}
				}
			}
		})
		if err != nil {
			c.logger.Errorf("Error: %s", err.Error())
		}
	}()
	return nil
}

// OnDisconnect will trigger func f when stream disconnects
func (c *SSEClient) OnDisconnect(f func() error) error {
	c.client.OnDisconnect(func(client *sse.Client) {
		if err := f(); err != nil {
			c.logger.Errorf("error invoking func on stream disconnect, err: %s", err.Error())
		}
	})
	return nil
}
