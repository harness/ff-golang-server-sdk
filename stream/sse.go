package stream

import (
	"context"
	"fmt"
	"github.com/harness/ff-golang-server-sdk/sdk_codes"
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

	proxyMode bool
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

	proxyMode bool,
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
		proxyMode:           proxyMode,
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
	c.logger.Infof("%s Start subscribing to Stream", sdk_codes.StreamStarted)
	// don't use the default exponentialBackoff strategy - we have our own disconnect logic
	// of polling the service then re-establishing a new stream once we can connect
	c.client.ReconnectStrategy = &backoff.StopBackOff{}
	// it is blocking operation, it needs to go in go routine

	out := make(chan Event)
	go func() {
		defer close(out)

		err := c.client.SubscribeWithContext(ctx, "*", func(msg *sse.Event) {
			c.logger.Infof("%s Event received: %s", sdk_codes.StreamEvent, msg.Data)

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
			c.logger.Warnf("Error initializing stream: %s", err.Error())
		}

		// The SSE library we use swallows the EOF error returned if a connection is closed by the server
		// so we need to call onStreamError any time we've exited SubscribeWithContext. If we don't do
		// this and the server closes the connection the Go SDK will still think it's connected to the stream
		// even though it isn't
		c.onStreamError()
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
			c.repository.DeleteFlags(event.Environment, cfMsg.Identifier)
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

			if c.proxyMode {
				updateFeaturesWithTimeout := func() {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
					defer cancel()

					response, err := c.api.GetFeatureConfigWithResponse(ctx, event.Environment)
					if err != nil {
						c.logger.Errorf("error while pulling flags, err: %s", err.Error())
						return
					}

					if response.JSON200 != nil {
						c.repository.SetFlags(false, event.Environment, *response.JSON200...)
					}
				}
				updateFeaturesWithTimeout()
			}
		}

	case dto.KeySegment:
		// need open client spec change
		switch cfMsg.Event {
		case dto.SseDeleteEvent:
			c.repository.DeleteSegment(cfMsg.Identifier)
			c.repository.DeleteSegments(event.Environment, cfMsg.Identifier)
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

			if c.proxyMode {
				updateSegmentsWithTimeout := func() {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
					defer cancel()

					response, err := c.api.GetAllSegmentsWithResponse(ctx, event.Environment)
					if err != nil {
						c.logger.Errorf("error while pulling segment, err: %s", err.Error())
						return
					}

					if response.JSON200 != nil {
						c.repository.SetSegments(false, event.Environment, *response.JSON200...)
					}
				}
				updateSegmentsWithTimeout()
			}
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
