package evaluation

import (
	"log"
	"strings"

	"github.com/spaolacci/murmur3"
)

const (
	// OneHundred MAX value for bucket
	OneHundred = 100
)

// GetNormalizedNumber returns normalized value with normalizer OneHundred
func GetNormalizedNumber(identifier string, bucketBy string) int {
	return GetNormalizedNumberWithNormalizer(identifier, bucketBy, OneHundred)
}

// GetNormalizedNumberWithNormalizer returns a murmur hash value based on input arguments
func GetNormalizedNumberWithNormalizer(identifier string, bucketBy string, normalizer int) int {
	value := []byte(strings.Join([]string{bucketBy, identifier}, ":"))
	hasher := murmur3.New32()
	_, err := hasher.Write(value)
	if err != nil {
		log.Printf("error %v", err)
	}
	hash := int(hasher.Sum32())
	return (hash % normalizer) + 1
}
