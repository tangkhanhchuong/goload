PROTO_FILE := api/goload.proto
OUT_DIR := internal/generated
SWAGGER_DIR := api

.PHONY: all generate clean

all: generate

generate:
	protoc -I=. \
		-I=$(shell go list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway/v2) \
		-I=$(shell go list -f '{{ .Dir }}' -m github.com/googleapis/googleapis) \
		-I=$(shell go list -f '{{ .Dir }}' -m github.com/envoyproxy/protoc-gen-validate) \
		--go_out=${OUT_DIR} \
		--go-grpc_out=${OUT_DIR} \
		--grpc-gateway_out=$(OUT_DIR) \
		--grpc-gateway_opt=generate_unbound_methods=true \
		--openapiv2_out=. \
		--openapiv2_opt=generate_unbound_methods=true \
		--validate_out=lang=go:$(OUT_DIR) \
		$(PROTO_FILE)

clean:
	rm -rf $(OUT_DIR)/*
	rm -rf $(SWAGGER_DIR)/*.swagger.json
