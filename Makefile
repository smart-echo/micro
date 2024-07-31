GOPATH:=$(shell go env GOPATH)

.PHONY: init
init:
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go get -u google.golang.org/protobuf/proto
	@go install github.com/smart-echo/micro-toolkit/cmd/protoc-gen-micro@latest

.PHONY: proto
proto:
	@buf dep update
	@buf generate

.PHONY: update
update:
	@go get -u

.PHONY: tidy
tidy:
	@go mod tidy