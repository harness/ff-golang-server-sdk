package evaluation

import (
	"fmt"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/harness/ff-golang-server-sdk/sdk_codes"

	"github.com/harness/ff-golang-server-sdk/log"
	"github.com/harness/ff-golang-server-sdk/rest"
	"github.com/spaolacci/murmur3"
)

func getAttrValue(target *Target, attr string) string {
	if target == nil || attr == "" {
		return ""
	}

	switch attr {
	case "identifier":
		return target.Identifier
	case "name":
		return target.Name
	default:
		if target.Attributes != nil {
			if val, ok := (*target.Attributes)[attr]; ok {
				switch v := val.(type) {
				case string:
					return v
				case int:
					return strconv.Itoa(v)
				case float64:
					return strconv.FormatFloat(v, 'f', -1, 64)
				case bool:
					return strconv.FormatBool(v)
				case map[string]interface{}:
					marshalledValue, err := jsoniter.MarshalToString(v)
					if err != nil {
						return fmt.Sprint(v)
					}
					return marshalledValue
				default:
					return fmt.Sprint(v)
				}
			}
		}
	}
	return ""
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
	log.Debugf("MM3 input [%s]", string(value))
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
	if value == "" {
		// If the original bucketBy attribute is not found, fallback to "identifier".
		value = getAttrValue(target, "identifier")
		if value == "" {
			return false
		}
		log.Debugf("%s BucketBy attribute not found in target attributes, falling back to 'identifier': missing=%s, using value=%s", sdk_codes.MissingBucketBy, bucketBy, value)
		bucketBy = "identifier"
	}

	// Calculate bucketID once with the resolved value and bucketBy.
	bucketID := getNormalizedNumber(value, bucketBy)
	log.Debugf("MM3 percentage_check=%d bucket_by=%s value=%s bucket=%d", percentage, bucketBy, value, bucketID)

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
