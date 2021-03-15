package types

// ErrorType is custom type for using it in const declarations
type ErrorType string

func (e ErrorType) Error() string {
	return string(e)
}

const (
	// ErrSdkCantBeEmpty represent error when no SDK key found in client api
	ErrSdkCantBeEmpty = ErrorType("Sdk can't be empty!")
	// ErrWrongTypeAssertion is custom error for wrong type assertions
	ErrWrongTypeAssertion = ErrorType("Wrong type assertion")
)
