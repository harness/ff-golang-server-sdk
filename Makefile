GOLANGCI_LINT_VERSION=v1.10.2

LINTER=./bin/golangci-lint

.PHONY: build clean test lint

generate:
	oapi-codegen -generate client,spec -package=rest ./api.yaml > ./pkg/rest/services.gen.go
	oapi-codegen -generate types -package=rest ./api.yaml > ./pkg/rest/types.gen.go

build:
	go build ./...

clean:
	go clean

test:
	go test -race -v --cover ./pkg/...

$(LINTER):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s $(GOLANGCI_LINT_VERSION)

lint: $(LINTER)
	$(LINTER) run ./...