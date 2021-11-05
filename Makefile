build-screen:
	CGO_ENABLED=1 GOARCH=amd64 go build -o ./release/screen ./cmd/screen

build-cli:
	CGO_ENABLED=1 GOARCH=amd64 go build -o ./release/cli ./cmd/cli

build: build-screen build-cli

build-image:
	docker build -t quay.io/mangirdas/unikiosk .

# for arm32 add linux/arm/v7
buildx-image:
	docker buildx build  --platform linux/amd64,linux/arm64 -t quay.io/mangirdas/unikiosk --push .

run:
	STATE_DIR=$(shell pwd)/data go run ./cmd/screen

.PHONY: proto
proto:
	@echo "--> Generating proto bindings..."
	@buf --config hack/protoc/buf.yaml --template hack/protoc/buf.gen.yaml generate