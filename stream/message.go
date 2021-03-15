package stream

// Message is basic object for marshalling data from ff stream
type Message struct {
	Event      string `json:"event"`
	Domain     string `json:"domain"`
	Identifier string `json:"identifier"`
	Version    int    `json:"version"`
}
