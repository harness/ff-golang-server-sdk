package stream

type Connection interface {
	Connect(environment string) error
	OnDisconnect(func() error) error
}
