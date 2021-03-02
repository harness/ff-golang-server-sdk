# Harness FFM Server-side SDK for Go

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
