package types

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/drone/ff-golang-server-sdk/log"
)

// Slice type for clause attribute evaluation
type Slice struct {
	data interface{}
}

// NewSlice creates a Slice instance with the object value
func NewSlice(value interface{}) Slice {
	return Slice{
		data: value,
	}
}

// In compares the attributes held by the slice with the input values.
// it will first determine what type of slice this is, before casting values
// to the appropriate type.
func (s Slice) In(values []string) bool {

	data := validateSlice(s.data)
	if data == nil {
		return false
	}

	// Determine what kind of slice we have
	switch attributes := data.(type) {
	case []string:
		for _, attribute := range attributes {
			if stringcmp(attribute, values) {
				return true
			}
		}
	case []float64:
		for _, attribute := range attributes {
			if float64cmp(attribute, values) {
				return true
			}
		}
	case []bool:
		for _, attribute := range attributes {
			if boolcmp(attribute, values) {
				return true
			}
		}
	case []interface{}:
		// case we get a []interface, this can store different types at the same time
		// e.g the first element could be a string, the next a float, so we need to
		// iterate over the slice and determine the type of each element
		for _, attribute := range attributes {
			switch attr := attribute.(type) {
			case string:
				if stringcmp(attr, values) {
					return true
				}
			case float64:
				if float64cmp(attr, values) {
					return true
				}
			case bool:
				if boolcmp(attr, values) {
					return true
				}
			}
		}
	default:
		log.Warn("unsupported attributes for 'in' comparison: [%+v]", data)

	}

	return false
}

// Equal always return false
func (s Slice) Equal(value []string) bool {
	return s.In(value)
}

// Contains always return false
func (s Slice) Contains(values []string) bool {
	return s.In(values)
}

// StartsWith always return false
func (s Slice) StartsWith(value []string) bool {
	return false
}

// EndsWith always return false
func (s Slice) EndsWith(value []string) bool {
	return false
}

// Match always return false
func (s Slice) Match(value []string) bool {
	return false
}

// EqualSensitive always return false
func (s Slice) EqualSensitive(value []string) bool {
	return false
}

// GreaterThan always return false
func (s Slice) GreaterThan(value []string) bool {
	return false
}

// GreaterThanEqual always return false
func (s Slice) GreaterThanEqual(value []string) bool {
	return false
}

// LessThan always return false
func (s Slice) LessThan(value []string) bool {
	return false
}

// LessThanEqual always return false
func (s Slice) LessThanEqual(value []string) bool {
	return false
}

// stringcmp compares the attribute with each element in values
// if any one of values matches it returns true, otherwise false.
func stringcmp(attribute string, values []string) bool {
	for _, value := range values {
		if strings.EqualFold(value, attribute) {
			return true
		}
	}
	return false
}

// float64cmp compares the attribute with each element in values
// It will cast the elements in the values slice to float64.
// if any one of values matches it returns true, otherwise false.
func float64cmp(attribute float64, values []string) bool {
	for _, value := range values {
		numberValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Warnf("input contains invalid value for number comparisons: %s\n", value)
			continue
		}
		if numberValue == attribute {
			return true
		}
	}
	return false
}

// boolcmp compares the attribute with each element in values
// It will cast the elements in the values slice to bool.
// if any one of values matches it returns true, otherwise false.
func boolcmp(attribute bool, values []string) bool {
	for _, value := range values {
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			log.Warnf("input contains invalid value for bool comparisons: %s\n", value)
			continue
		}
		if boolValue == attribute {
			return true
		}
	}
	return false
}

// validateSlice checks if this is a ptr and deference it
// then validates that we are working with a slice.
func validateSlice(data interface{}) interface{} {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		data = v.Interface()
	}

	if v.Kind() != reflect.Slice {
		return nil
	}

	return data
}
