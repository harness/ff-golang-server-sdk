module github.com/harness/ff-golang-server-sdk

go 1.20

require (
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/deepmap/oapi-codegen/v2 v2.1.0
	github.com/getkin/kin-openapi v0.124.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.5.0
	github.com/harness-community/sse/v3 v3.1.0
	github.com/hashicorp/go-retryablehttp v0.7.7
	github.com/hashicorp/golang-lru v0.5.4
	github.com/jarcoal/httpmock v1.0.8
	github.com/json-iterator/go v1.1.12
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.3.3
	github.com/oapi-codegen/runtime v1.1.1
	github.com/spaolacci/murmur3 v1.1.0
	github.com/stretchr/testify v1.8.4
	go.uber.org/zap v1.16.0
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-openapi/jsonpointer v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.8 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/invopop/yaml v0.2.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	gopkg.in/cenkalti/backoff.v1 v1.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract [v0.1.21, v0.1.22] // Panic in metrics code if target attributes are not provided (nil)
