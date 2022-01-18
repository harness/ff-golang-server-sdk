package stream

import (
	"context"

	"github.com/r3labs/sse"
)

// Connection is simple interface for streams
type Connection interface {
	Connect(environment string) error
	OnDisconnect(func() error) error
}

// EventStreamListener provides a way to hook in to the SSE Events that the SDK
// recieves from the FeatureFlags server and forward them on to another type.
type EventStreamListener interface {
	// Pub publishes an event from the SDK to your Listener. Pub should implement
	// any backoff/retry logic as this is not handled in the SDK.
	Pub(ctx context.Context, event Event) error
}

// Event defines the structure of an event that gets sent to a EventStreamListener
type Event struct {
	// APIKey is the SDKs API Key
	APIKey string
	// Environment is the ID of the environment that the event occured for
	Environment string
	// Event is the SSEEvent that was sent from the FeatureFlags server to the SDK
	Event *sse.Event
}
