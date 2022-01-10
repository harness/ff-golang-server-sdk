module github.com/harness/ff-golang-server-sdk/test_wrapper

go 1.16

replace github.com/harness/ff-golang-server-sdk => ../

require (
	github.com/deepmap/oapi-codegen v1.6.1 // indirect
	github.com/getkin/kin-openapi v0.53.0
	github.com/harness/ff-golang-server-sdk v0.0.22
	github.com/labstack/echo/v4 v4.2.2
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.7.1
)
