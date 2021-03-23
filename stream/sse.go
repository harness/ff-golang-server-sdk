package stream

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/drone/ff-golang-server-sdk.v0/cache"
	"github.com/drone/ff-golang-server-sdk.v0/dto"
	"github.com/drone/ff-golang-server-sdk.v0/rest"
	jsoniter "github.com/json-iterator/go"
	"github.com/r3labs/sse"
)

// SSEClient is Server Send Event object
type SSEClient struct {
	api    rest.ClientWithResponsesInterface
	client *sse.Client
	cache  cache.Cache
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// NewSSEClient creates an object for stream interactions
func NewSSEClient(
	apiKey string,
	token string,
	client *sse.Client,
	cache cache.Cache,
	api rest.ClientWithResponsesInterface,
) *SSEClient {
	client.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	client.Headers["API-Key"] = apiKey
	return &SSEClient{
		client: client,
		cache:  cache,
		api:    api,
	}
}

// Connect will subscribe to SSE stream
func (c *SSEClient) Connect(environment string) error {
	log.Println("Start subscribing to Stream")
	// it is blocking operation, it needs to go in go routine
	go func() {
		err := c.client.Subscribe("*", func(msg *sse.Event) {
			log.Printf("Event received: %s", msg.Data)

			cfMsg := Message{}
			err := json.Unmarshal(msg.Data, &cfMsg)
			if err != nil {
				log.Fatal(err)
			}

			switch cfMsg.Domain {
			case dto.KeyFeature:
				// maybe is better to send event on memory bus that we get new message
				// and subscribe to that event
				go func(env, identifier string) {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
					response, err := c.api.GetFeatureConfigByIdentifierWithResponse(ctx, env, identifier)
					if err != nil {
						log.Printf("error while pulling flag, err: %s", err)
						cancel()
						return
					}
					if response.JSON200 != nil {
						c.cache.Set(dto.Key{
							Type: dto.KeyFeature,
							Name: cfMsg.Identifier,
						}, *response.JSON200.Convert())
					}
					cancel()
				}(environment, cfMsg.Identifier)
			case dto.KeySegment:
				// need open client spec change
			}
		})
		if err != nil {
			log.Printf("Error: %s", err)
		}
	}()
	return nil
}

// OnDisconnect will trigger func f when stream disconnects
func (c SSEClient) OnDisconnect(f func() error) error {
	c.client.OnDisconnect(func(c *sse.Client) {
		if err := f(); err != nil {
			log.Printf("error invoking func on stream disconnect, err: %v", err)
		}
	})
	return nil
}
