package evaluation

import "errors"

var (
	// ErrQueryProviderMissing ...
	ErrQueryProviderMissing = errors.New("query field is missing in evaluator")
	// ErrVariationNotFound ...
	ErrVariationNotFound = errors.New("variation not found")
	// ErrEvaluationFlag ...
	ErrEvaluationFlag = errors.New("error while evaluating flag")
	// ErrFlagKindMismatch ...
	ErrFlagKindMismatch = errors.New("flag kind mismatch")
)
