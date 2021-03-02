package utils

type Event struct {
	EventType string
	Key       string
	Value     interface{}
}

const (
	SAVE   = "save"
	DELETE = "delete"
)

type Observable interface {
	AddObserver(observer Observer)
	Notify(event *Event)
	RemoveObserver(observer interface{})
}

type Observer interface {
	NotifyCallback(event *Event)
}
