package types

// ErrorType is custom type for using it in const declarations
type ErrorType string

func (e ErrorType) Error() string {
	return string(e)
}

const (
	// ErrWrongTypeAssertion is custom error for wrong type assertions
	ErrWrongTypeAssertion = ErrorType("Wrong type assertion")
)
