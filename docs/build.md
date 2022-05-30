# Building ff-golang-server-sdk

This document shows the instructions on how to build and contribute to the SDK.

## Requirements
[Golang 1.6](https://go.dev/doc/install) or newer (go version)<br>

## Install Dependencies

```bash
go mod tidy
```

## Build the SDK
Some make targets have been provided to build and package the SDK

```bash
go build ./...
```

## Executing tests
```bash
 go test -race -v --cover ./...
```

## Linting and Formating
To ensure the project is correctly formatted you can use the following commands
```bash
go fmt $(go list ./...)
go vet ./...
```
