PROTO_FILE := api/goload.proto
OUT_DIR := internal/generated
SWAGGER_DIR := api

.PHONY: all generate clean vendor run_dev

generate:
	buf generate api
	wire internal/wiring/wire.go

run_dev: generate
	go run cmd/main.go server

clean:
	rm -rf $(OUT_DIR)/*
	rm -rf $(SWAGGER_DIR)/*.swagger.json

vendor:
	go mod tidy
	go mod vendor
