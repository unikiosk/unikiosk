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

.PHONY: proto
proto:
	@echo "--> Generating proto bindings..."
	@buf --config hack/protoc/buf.yaml --template hack/protoc/buf.gen.yaml generate

run-ui:
	go run ./hack/ui

run-screen:
	STATE_DIR=$(shell pwd)/data \
	WEB_SERVER_DIR=$(shell pwd)/ui \
	go run ./cmd/screen

lint:
	gofmt -s -w cmd hack pkg
	go run -mod vendor ./vendor/golang.org/x/tools/cmd/goimports -w -local=github.com/synpse-hq/synpse-core cmd hack pkg
	go run -mod vendor ./hack/validate-imports cmd hack pkg
	# TODO: Enable this at some point
	#staticcheck ./...

test:
	go test -mod=vendor -v -failfast `go list ./... | egrep -v /test/`

clean:
	rm -rf data