package apiconfig

import "github.com/harness/ff-golang-server-sdk/rest"

type ApiConfiguration interface {
	GetSegmentRulesV2QueryParam() *rest.SegmentRulesV2QueryParam
}
