module github.com/harness/ff-golang-server-sdk/test_wrapper

go 1.16

replace github.com/harness/ff-golang-server-sdk => ../

require (
	github.com/getkin/kin-openapi v0.94.0
	github.com/harness/ff-golang-server-sdk v0.0.24
	github.com/labstack/echo/v4 v4.9.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.10.1
)
