FROM golang:1.14-alpine as builder
RUN apk update && apk add --no-cache make gcc musl-dev git ca-certificates && update-ca-certificates
WORKDIR /app

# Fetch dependencies.
COPY ./go.mod .
RUN go mod download

COPY . .
RUN go mod verify
RUN make install
RUN make generate

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/go_wrapper .

FROM alpine:latest
RUN apk update && apk add --no-cache bash
COPY --from=builder /app/go_wrapper /app/go_wrapper

CMD ["/app/go_wrapper"]