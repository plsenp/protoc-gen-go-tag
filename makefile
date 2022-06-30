GOPATH:=$(shell go env GOPATH)
API_PROTO_FILES=$(shell find examples  -name *.proto)
GO_PB_FILES=$(shell find examples -name *.pb.go)

.PHONY: api proto
proto: 
	protoc --proto_path=./examples \
 	    --go_out=paths=source_relative:./examples \
	    $(API_PROTO_FILES)
	

api: proto
	protoc-gen-go-tag $(GO_PB_FILES)



# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help