# Before you Begin

Harness Feature Flags (FF) is a feature management solution that enables users to change the software’s functionality, without deploying new code. FF uses feature flags to hide code or behaviours without having to ship new versions of the software. A feature flag is like a powerful if statement.

For more information, see https://harness.io/products/feature-flags/

To read more, see https://ngdocs.harness.io/category/vjolt35atg-feature-flags

To sign up, https://app.harness.io/auth/#/signup/


# Harness FFM Server-side SDK for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/drone/ff-golang-server-sdk)](https://goreportcard.com/report/github.com/drone/ff-golang-server-sdk)

## FFM overview
FFM is feature flag management platform for helping teams to deliver better software and faster.

## Supported GO versions
This version of FFM has been tested with GO 1.14

## Install
`go get github.com/wings-software/ff-client-sdk-go`

## Usage
First we need to import lib with harness alias
`import harness "github.com/wings-software/ff-client-sdk-go/pkg/api"`

Next we create client instance for interaction with api
`client := harness.NewClient(sdkKey)`

Target definition can be user, device, app etc.
```
target := dto.NewTargetBuilder("key").
 		Firstname("John").
 		Lastname("doe").
 		Email("johndoe@acme.com").
 		Country("USA").
 		Custom("height", 160).
 		Build()
```

Evaluating Feature Flag
`showFeature, err := client.BoolVariation(featureFlagKey, target, false)`
