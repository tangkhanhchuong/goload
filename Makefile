PROTO_FILE := api/goload.proto
OUT_DIR := internal/generated
SWAGGER_DIR := api

.PHONY: all generate clean vendor run_dev

run_dev:
	go run cmd/main.go server

generate:
	buf generate api
	

clean:
	rm -rf $(OUT_DIR)/*
	rm -rf $(SWAGGER_DIR)/*.swagger.json

vendor:
	go mod tidy
	go mod vendor
