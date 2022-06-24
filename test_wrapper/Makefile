

install:
	go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.11.0

generate:
	# generate admin
	oapi-codegen --old-config-style -generate server,spec -package=restapi ./test_wrapper_openapi/sdk_wrapper-v1.yaml > ./restapi/test_wrapper_service.gen.go
	oapi-codegen --old-config-style -generate types -package=restapi ./test_wrapper_openapi/sdk_wrapper-v1.yaml > ./restapi/test_wrapper_types.gen.go
