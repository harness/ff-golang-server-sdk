

install:
	go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.6.0

generate:
	# generate admin
	oapi-codegen -generate server,spec -package=restapi ./test_wrapper_openapi/sdk_wrapper-v1.yaml > ./restapi/test_wrapper_service.gen.go
	oapi-codegen -generate types -package=restapi ./test_wrapper_openapi/sdk_wrapper-v1.yaml > ./restapi/test_wrapper_types.gen.go

build:
	docker build -t us.gcr.io/${PROJECT_ID}/${IMAGE}:latest -f ./docker/Dockerfile .