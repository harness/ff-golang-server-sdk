package apiconfig

import "github.com/harness/ff-golang-server-sdk/rest"

// ApiConfiguration is a type that provides the required configuration for requests
type ApiConfiguration interface {
	GetSegmentRulesV2QueryParam() *rest.SegmentRulesV2QueryParam
}
