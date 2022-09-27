package stream

import (
	"context"
	"fmt"
	"time"

	"github.com/harness/ff-golang-server-sdk/pkg/repository"

	"github.com/harness/ff-golang-server-sdk/dto"
	"github.com/harness/ff-golang-server-sdk/logger"
	"github.com/harness/ff-golang-server-sdk/rest"
	backoff "gopkg.in/cenkalti/backoff.v1"

	jsoniter "github.com/json-iterator/go"
	"github.com/r3labs/sse"
)

// SSEClient is Server Send Event object
type SSEClient struct {
	api                 rest.ClientWithResponsesInterface
	client              *sse.Client
	repository          repository.Repository
	logger              logger.Logger
	onStreamError       func()
	eventStreamListener EventStreamListener
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// NewSSEClient creates an object for stream interactions
func NewSSEClient(
	apiKey string,
	token string,
	client *sse.Client,
	repository repository.Repository,
	api rest.ClientWithResponsesInterface,
	logger logger.Logger,
	onStreamError func(),
	eventStreamListener EventStreamListener,
) *SSEClient {
	client.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	client.Headers["API-Key"] = apiKey
	client.OnDisconnect(func(client *sse.Client) {
		onStreamError()
	})
	sseClient := &SSEClient{
		client:              client,
		repository:          repository,
		api:                 api,
		logger:              logger,
		onStreamError:       onStreamError,
		eventStreamListener: eventStreamListener,
	}
	return sseClient
}

// Connect will subscribe to SSE stream
func (c *SSEClient) Connect(ctx context.Context, environment string, apiKey string) {
	go func() {
		for event := range orDone(ctx, c.subscribe(ctx, environment, apiKey)) {
			c.handleEvent(event)
		}
	}()
}

// Connect will subscribe to SSE stream
func (c *SSEClient) subscribe(ctx context.Context, environment string, apiKey string) <-chan Event {
	c.logger.Infof("Start subscribing to Stream")
	// don't use the default exponentialBackoff strategy - we have our own disconnect logic
	// of polling the service then re-establishing a new stream once we can connect
	c.client.ReconnectStrategy = &backoff.StopBackOff{}
	// it is blocking operation, it needs to go in go routine

	out := make(chan Event)
	go func() {
		defer close(out)

		err := c.client.SubscribeWithContext(ctx, "*", func(msg *sse.Event) {
			c.logger.Infof("Event received: %s", msg.Data)

			if len(msg.Data) <= 0 {
				return
			}

			event := Event{
				APIKey:      apiKey,
				Environment: environment,
				SSEEvent:    msg,
			}

			select {
			case <-ctx.Done():
				return
			case out <- event:
			}

		})
		if err != nil {
			c.logger.Errorf("Error initializing stream: %s", err.Error())
			c.onStreamError()
		}
	}()

	return out
}

func (c *SSEClient) handleEvent(event Event) {
	cfMsg := Message{}
	err := json.Unmarshal(event.SSEEvent.Data, &cfMsg)
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
			c.repository.DeleteFlag(cfMsg.Identifier)
		case dto.SsePatchEvent, dto.SseCreateEvent:
			fallthrough
		default:
			updateWithTimeout := func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
				defer cancel()

				response, err := c.api.GetFeatureConfigByIdentifierWithResponse(ctx, event.Environment, cfMsg.Identifier)
				if err != nil {
					c.logger.Errorf("error while pulling flag, err: %s", err.Error())
					return
				}

				if response.JSON200 != nil {
					c.repository.SetFlag(*response.JSON200, false)
				}
			}

			updateWithTimeout()
		}

	case dto.KeySegment:
		// need open client spec change
		switch cfMsg.Event {
		case dto.SseDeleteEvent:
			c.repository.DeleteSegment(cfMsg.Identifier)
		case dto.SsePatchEvent, dto.SseCreateEvent:
			fallthrough
		default:
			updateWithTimeout := func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
				defer cancel()

				response, err := c.api.GetSegmentByIdentifierWithResponse(ctx, event.Environment, cfMsg.Identifier)
				if err != nil {
					c.logger.Errorf("error while pulling segment, err: %s", err.Error())
					return
				}
				if response.JSON200 != nil {
					c.repository.SetSegment(*response.JSON200, false)
				}
			}
			updateWithTimeout()
		}
	}

	if c.eventStreamListener != nil {
		sendWithTimeout := func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			return c.eventStreamListener.Pub(ctx, Event{APIKey: event.APIKey, Environment: event.Environment, SSEEvent: event.SSEEvent})
		}

		if err := sendWithTimeout(); err != nil {
			c.logger.Errorf("error while forwarding SSE Event to change stream: %s", err)
		}
	}
}

// orDone is a helper that encapsulates the logic for reading from a channel
// whilst waiting for a cancellation.
func orDone(ctx context.Context, c <-chan Event) <-chan Event {
	out := make(chan Event)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case cp, ok := <-c:
				if !ok {
					return
				}

				select {
				case <-ctx.Done():
				case out <- cp:
				}
			}
		}
	}()

	return out
}
