package stream

// Connection is simple interface for streams
type Connection interface {
	Connect(environment string) error
	OnDisconnect(func() error) error
}
