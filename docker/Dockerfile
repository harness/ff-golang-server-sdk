FROM golang:1.16-alpine as builder
RUN apk update && apk add --no-cache make gcc musl-dev git ca-certificates && update-ca-certificates
WORKDIR /app

COPY . .
WORKDIR /app/test_wrapper
RUN go mod download
RUN go mod verify
RUN go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.6.0
RUN oapi-codegen -generate server,spec -package=restapi ./test_wrapper_openapi/sdk_wrapper-v1.yaml > ./restapi/test_wrapper_service.gen.go
RUN oapi-codegen -generate types -package=restapi ./test_wrapper_openapi/sdk_wrapper-v1.yaml > ./restapi/test_wrapper_types.gen.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/go_wrapper .

FROM alpine:latest
RUN apk update && apk add --no-cache bash
COPY --from=builder /app/go_wrapper /app/go_wrapper

CMD ["/app/go_wrapper"]
