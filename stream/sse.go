package stream

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/harness/ff-golang-server-sdk/apiconfig"
	"github.com/harness/ff-golang-server-sdk/sdk_codes"

	"github.com/harness/ff-golang-server-sdk/pkg/repository"

	"github.com/cenkalti/backoff/v4"
	"github.com/harness/ff-golang-server-sdk/dto"
	"github.com/harness/ff-golang-server-sdk/logger"
	"github.com/harness/ff-golang-server-sdk/rest"

	"github.com/harness-community/sse/v3"
	jsoniter "github.com/json-iterator/go"
)

// SSEClient is Server Send Event object
type SSEClient struct {
	api                 rest.ClientWithResponsesInterface
	client              *sse.Client
	repository          repository.Repository
	logger              logger.Logger
	eventStreamListener EventStreamListener
	streamConnected     chan struct{}
	streamDisconnected  chan error
	apiConfig           apiconfig.ApiConfiguration
	proxyMode           bool
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
	eventStreamListener EventStreamListener,
	proxyMode bool,
	streamConnected chan struct{},
	streamDisconnected chan error,
	apiConfig apiconfig.ApiConfiguration,

) *SSEClient {
	client.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	client.Headers["API-Key"] = apiKey
	sseClient := &SSEClient{
		client:              client,
		repository:          repository,
		api:                 api,
		logger:              logger,
		eventStreamListener: eventStreamListener,
		proxyMode:           proxyMode,
		streamConnected:     streamConnected,
		streamDisconnected:  streamDisconnected,
		apiConfig:           apiConfig,
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
	c.logger.Info("Attempting to start stream")
	// don't use the default exponentialBackoff strategy - we have our own disconnect logic
	// of polling the service then re-establishing a new stream once we can connect
	c.client.ReconnectStrategy = &backoff.StopBackOff{}

	// If we haven't received a change event or heartbeat in 30 seconds, we consider the stream to be "dead" and force a
	// reconnection
	const timeout = 30 * time.Second
	deadStreamTimer := time.NewTimer(timeout)
	// Stop the timer immediately, it will only start when the connection is established
	deadStreamTimer.Stop()

	onConnect := func(s *sse.Client) {
		// Start the dead stream timer
		deadStreamTimer.Reset(timeout)

		// The SDK stops reading this channel once its context is cancelled, so guard the send
		// to avoid parking this goroutine for the lifetime of the process.
		select {
		case c.streamConnected <- struct{}{}:
		case <-ctx.Done():
		}
	}
	c.client.OnConnect(onConnect)
	out := make(chan Event)
	go func() {
		defer close(out)

		// Create another context off of the main SDK context, so we can close dead streams.
		deadStreamCtx, deadStreamCancel := context.WithCancel(ctx)
		defer deadStreamCancel()
		defer deadStreamTimer.Stop()

		// Signals that this subscription was torn down because the stream went dead, rather than
		// because the server closed the connection or the SDK shut down.
		deadStream := make(chan struct{})

		// This waiter must observe deadStreamCtx and not the SDK context: the SDK context lives for
		// the lifetime of the process, whereas deadStreamCtx is cancelled every time a subscription
		// ends. Stopping the timer does not send on its channel, so watching the SDK context here
		// parks one goroutine per reconnect for the lifetime of the process.
		go func() {
			select {
			case <-deadStreamCtx.Done():
				return
			case <-deadStreamTimer.C:
				close(deadStream)
				deadStreamCancel()
				return
			}
		}()

		// The SDK stops reading this channel once its context is cancelled, so guard the send
		// to avoid parking this goroutine for the lifetime of the process.
		notifyDisconnected := func(err error) {
			select {
			case c.streamDisconnected <- err:
			case <-ctx.Done():
			}
		}

		err := c.client.SubscribeWithContext(deadStreamCtx, "*", func(msg *sse.Event) {

			deadStreamTimer.Stop()
			deadStreamTimer.Reset(timeout)

			// Heartbeat event
			// Data should always be nil here, but use an extra defensive check in case it's a byte slice
			if msg.Data == nil || len(msg.Data) < 1 {
				c.logger.Debugf("%s Heartbeat event received", sdk_codes.StreamEvent)
				return
			}

			c.logger.Infof("%s Event received: %s", sdk_codes.StreamEvent, msg.Data)

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

		// Release the dead stream waiter before signalling the disconnect: the signal can block
		// until the SDK reconnects, and the waiter must not outlive this subscription.
		deadStreamTimer.Stop()
		deadStreamCancel()

		if err != nil {
			notifyDisconnected(err)
			return
		}

		// When we cancel the deadStreamContext, we exit the `SubscribeWithContext` function with a nil error,
		// so we need an explicit check to see why the subscription ended.
		select {
		case <-deadStream:
			notifyDisconnected(fmt.Errorf("no SSE events received for 30 seconds. Assuming stream is dead and restarting"))
			return
		default:
		}

		// The SDK is shutting down, there is nothing to reconnect to.
		if errors.Is(ctx.Err(), context.Canceled) {
			return
		}

		// The SSE library we use currently returns a nil error for io.EOF errors, which can
		// happen if ff-server closes the connection at the 24 hours point.
		// So we need to signal the stream disconnected channel any time we've exited SubscribeWithContext.
		// If we don't do this and the server closes the connection the Go SDK will still think it's connected to the stream
		// even though it isn't.
		notifyDisconnected(fmt.Errorf("server closed the connection"))
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

				response, err := c.api.GetFeatureConfigByIdentifierWithResponse(ctx, event.Environment, cfMsg.Identifier, nil)
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

					response, err := c.api.GetFeatureConfigWithResponse(ctx, event.Environment, nil)
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
				requestParams := &rest.GetSegmentByIdentifierParams{
					Rules: c.apiConfig.GetSegmentRulesV2QueryParam(),
				}
				response, err := c.api.GetSegmentByIdentifierWithResponse(ctx, event.Environment, cfMsg.Identifier, requestParams)
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
					requestParams := &rest.GetAllSegmentsParams{
						Rules: c.apiConfig.GetSegmentRulesV2QueryParam(),
					}
					response, err := c.api.GetAllSegmentsWithResponse(ctx, event.Environment, requestParams)
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
