package observable

// Event object in observables
type Event struct {
	EventType string
	Key       string
	Value     interface{}
}

const (
	// SAVE event type used for storing
	SAVE = "save"
	// DELETE event type used for deletion
	DELETE = "delete"
)

// Observable can be used for listening changes on cache
type Observable interface {
	AddObserver(observer Observer)
	Notify(event *Event)
	RemoveObserver(observer interface{})
}

// Observer notify all observable clients
type Observer interface {
	NotifyCallback(event *Event)
}
