package dto

// Key holds information for unique data for cache and store
type Key struct {
	Type string
	Name string
}

const (
	// KeyFeatures ...
	KeyFeatures = "flags"

	// KeyFeature identifies flag messages from ff server or stream
	KeyFeature = "flag"

	// KeySegment identifies segment messages from ff server or stream
	KeySegment = "target-segment"

	// KeySegment identifies segment messages from ff server or stream
	KeySegments = "target-segments"

	// SsePatchEvent identifies a patch event from the SSE stream
	SsePatchEvent = "patch"

	// SseDeleteEvent identifies a delete event from the SSE stream
	SseDeleteEvent = "delete"

	// SseCreateEvent identifies a create event from the SSE stream
	SseCreateEvent = "create"
)
