package evaluation

import (
	"strings"

	"github.com/spaolacci/murmur3"
)

const (
	OneHundred = 100
)

func GetNormalizedNumber(identifier string, bucketBy string) int {
	return GetNormalizedNumberWithNormalizer(identifier, bucketBy, OneHundred)
}

func GetNormalizedNumberWithNormalizer(identifier string, bucketBy string, normalizer int) int {
	value := []byte(strings.Join([]string{bucketBy, identifier}, ":"))
	hasher := murmur3.New32()
	hasher.Write(value)
	hash := int(hasher.Sum32())
	return (hash % normalizer) + 1
}
