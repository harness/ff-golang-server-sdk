package evaluation

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/harness/ff-golang-server-sdk/log"
	"github.com/harness/ff-golang-server-sdk/rest"
	"github.com/spaolacci/murmur3"
)

func getAttrValue(target *Target, attr string) reflect.Value {
	var value reflect.Value
	if target == nil {
		return value
	}

	attrs := make(map[string]interface{})
	if target.Attributes != nil {
		attrs = *target.Attributes
	}

	attrVal, ok := attrs[attr] // first check custom attributes
	if ok {
		value = reflect.ValueOf(attrVal)
	} else {
		// We only have two fields here, so we will access the fields directly, and use reflection if we start adding
		// more in the future
		switch strings.ToLower(attr) {
		case "identifier":
			value = reflect.ValueOf(target.Identifier)
		case "name":
			value = reflect.ValueOf(target.Name)
		}
	}
	return value
}

func reflectValueToString(val reflect.Value) string {
	stringValue := ""
	switch val.Kind() {
	case reflect.Int, reflect.Int64:
		stringValue = strconv.FormatInt(val.Int(), 10)
	case reflect.Bool:
		stringValue = strconv.FormatBool(val.Bool())
	case reflect.String:
		stringValue = val.String()
	case reflect.Array, reflect.Chan, reflect.Complex128, reflect.Complex64, reflect.Func, reflect.Interface,
		reflect.Invalid, reflect.Ptr, reflect.Slice, reflect.Struct, reflect.Uintptr, reflect.UnsafePointer,
		reflect.Float32, reflect.Float64, reflect.Int16, reflect.Int32, reflect.Int8, reflect.Map, reflect.Uint,
		reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		stringValue = fmt.Sprintf("%v", val)
	default:
		// Use string formatting as last ditch effort for any unexpected values
		stringValue = fmt.Sprintf("%v", val)
	}
	return stringValue
}

func findVariation(variations []rest.Variation, identifier string) (rest.Variation, error) {
	for _, variation := range variations {
		if variation.Identifier == identifier {
			return variation, nil
		}
	}
	return rest.Variation{}, fmt.Errorf("%w: %s", ErrVariationNotFound, identifier)
}

func getNormalizedNumber(identifier, bucketBy string) int {
	value := []byte(strings.Join([]string{bucketBy, identifier}, ":"))
	hasher := murmur3.New32()
	_, err := hasher.Write(value)
	if err != nil {
		log.Debugf("error %v", err)
	}
	hash := int(hasher.Sum32())
	return (hash % oneHundred) + 1
}

func isEnabled(target *Target, bucketBy string, percentage int) bool {
	value := getAttrValue(target, bucketBy)
	identifier := value.String()
	if identifier == "" {
		return false
	}

	bucketID := getNormalizedNumber(identifier, bucketBy)
	return percentage > 0 && bucketID <= percentage
}

func evaluateDistribution(distribution *rest.Distribution, target *Target) string {
	variation := ""
	if distribution == nil {
		return variation
	}

	totalPercentage := 0
	for _, wv := range distribution.Variations {
		variation = wv.Variation
		totalPercentage += wv.Weight
		if isEnabled(target, distribution.BucketBy, totalPercentage) {
			return wv.Variation
		}
	}
	return variation
}

func isTargetInList(target *Target, targets []rest.Target) bool {
	if targets == nil || target == nil {
		return false
	}
	for _, includedTarget := range targets {
		if includedTarget.Identifier == target.Identifier {
			return true
		}
	}
	return false
}
