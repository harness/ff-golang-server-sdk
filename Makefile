GOLANGCI_LINT_VERSION=v1.37.1

ifndef GOPATH
	GOPATH := $(shell go env GOPATH)
endif

# Default target will fetch deps, build and run tests.
all: tidy generate build check test

generate:
	oapi-codegen --config ./resources/config/rest-client-config.yaml ./resources/client-v1.yaml > rest/services.gen.go
	oapi-codegen --config ./resources/config/rest-types-config.yaml ./resources/client-v1.yaml > rest/types.gen.go
	oapi-codegen --config ./resources/config/metrics-client-config.yaml ./resources/metrics-v1.yaml > metricsclient/services.gen.go
	oapi-codegen --config ./resources/config/metrics-types-config.yaml ./resources/metrics-v1.yaml > metricsclient/types.gen.go

tidy:
	go mod tidy

build:
	go build ./...

check: format lint sec

clean:
	go clean

test:
	go test -race -v --cover ./...

report:
	go test ./... -covermode=atomic -coverpkg=./...  -coverprofile=c.out
	gocov convert ./c.out | gocov-html > ~/go-sdk-test-report.html


build-test-wrapper:
	docker build -t us.gcr.io/${PROJECT_ID}/${IMAGE}:latest -f ./docker/Dockerfile .

# Format go code and error if any changes are made
PHONY+= format
format:
	@echo "Checking that go fmt does not make any changes..."
	@test -z $$(go fmt $(go list ./...)) || (echo "go fmt would make a change. Please verify and commit the proposed changes"; exit 1)
	@echo "Checking go fmt complete"
	@echo "Running goimports"
	@test -z $$(goimports -w ./..) || (echo "goimports would make a change. Please verify and commit the proposed changes"; exit 1)

PHONY+= lint
lint: $(GOPATH)/bin/golangci-lint $(GOPATH)/bin/golint
	@echo "Linting $(1)"
	@golint -set_exit_status ./...
	@go vet ./...
	@golangci-lint run \
		-E asciicheck \
		-E bodyclose \
		-E exhaustive \
		-E exportloopref \
		-E gofmt \
		-E goimports \
		-E gosec \
		-E noctx \
		-E nolintlint \
		-E rowserrcheck \
		-E exportloopref \
		-E sqlclosecheck \
		-E stylecheck \
		-E unconvert \
		-E unparam
	@echo "Lint-free"

#
# Install Tools 
#
PHONY+= sec
sec: $(GOPATH)/bin/gosec
	@echo "Checking for security problems ..."
	@gosec -quiet -confidence high -severity medium ./...
	@echo "No problems found"; \

$(GOPATH)/bin/golangci-lint:
	@echo "🔘 Installing golangci-lint... (`date '+%H:%M:%S'`)"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin

$(GOPATH)/bin/golint:
	@echo "🔘 Installing golint ... (`date '+%H:%M:%S'`)"
	@GO111MODULE=off go get -u golang.org/x/lint/golint

$(GOPATH)/bin/goimports:
	@echo "🔘 Installing goimports ... (`date '+%H:%M:%S'`)"
	@GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports

$(GOPATH)/bin/gosec:
	@echo "🔘 Installing gosec ... (`date '+%H:%M:%S'`)"
	@curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(GOPATH)/bin


$(GOPATH)/bin/oapi-codegen:
	@echo "🔘 Installing oapicodegen ... (`date '+%H:%M:%S'`)"
	@go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.11.0

PHONY+= tools
tools: $(GOPATH)/bin/golangci-lint $(GOPATH)/bin/golint $(GOPATH)/bin/gosec $(GOPATH)/bin/goimports $(GOPATH)/bin/oapi-codegen

PHONY+= update-tools
update-tools: delete-tools $(GOPATH)/bin/golangci-lint $(GOPATH)/bin/golint $(GOPATH)/bin/gosec $(GOPATH)/bin/goimports

PHONY+= delete-tools
delete-tools:
	@rm $(GOPATH)/bin/golangci-lint
	@rm $(GOPATH)/bin/gosec
	@rm $(GOPATH)/bin/golint
	@rm $(GOPATH)/bin/goimports


.PHONY: all tidy generate build clean test lint
