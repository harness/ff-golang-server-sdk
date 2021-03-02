package dto

// Key holds information for unique data
// for cache and store
type Key struct {
	Type string
	Name string
}

const (
	KeyFeature = "flag"
	KeySegment = "target-segment"
)
