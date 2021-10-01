package types

// ValueType interface used for different clause types
type ValueType interface {
	StartsWith(value []string) bool
	EndsWith(value []string) bool
	Match(value []string) bool
	Contains(value []string) bool
	EqualSensitive(value []string) bool
	Equal(value []string) bool
	GreaterThan(value []string) bool
	GreaterThanEqual(value []string) bool
	LessThan(value []string) bool
	LessThanEqual(value []string) bool
	In(value []string) bool
}
