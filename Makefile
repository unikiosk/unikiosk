build-screen:
	 GOARCH=amd64 go build -o ./release/screen ./cmd/screen

build-screen-arm64:
	 GOARCH=arm64 go build -o ./release/screen-arm64 ./cmd/screen


build-cli:
	GOARCH=amd64 go build -o ./release/cli ./cmd/cli

build: build-screen build-cli

build-image:
	docker build -t quay.io/unikiosk/unikiosk -f dockerfiles/Dockerfile .

# fix error: /bin/sh: Invalid ELF image for this architecture 
fix-multiarch:
	docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

# for arm32 add linux/arm/v7
buildx-image:
	docker buildx create --use && \
	docker buildx build --platform linux/amd64,linux/arm64 -t quay.io/unikiosk/unikiosk --push -f dockerfiles/Dockerfile.firefox .

dev-buildx-image:
	docker buildx create --use && \
	docker buildx build --platform linux/amd64,linux/arm64 -t quay.io/unikiosk/unikiosk --push -f dockerfiles/Dockerfile.firefox .

dev-buildx-image-chromium:
	docker buildx create --use && \
	docker buildx build --platform linux/amd64,linux/arm64 -t quay.io/unikiosk/unikiosk:chromium --push -f dockerfiles/Dockerfile.chromium .

push-image-test:
	docker push quay.io/unikiosk/unikiosk

build-image-test:
	docker build -t quay.io/unikiosk/test-base -f dockerfiles/Dockerfile.test  .

push-image-test:
	docker push quay.io/unikiosk/test-base

.PHONY: proto
proto:
	@echo "--> Generating proto bindings..."
	@buf --config hack/protoc/buf.yaml --template hack/protoc/buf.gen.yaml generate

run-ui:
	go run ./hack/ui

run-screen:
	STATE_DIR=$(shell pwd)/data \
	WEB_SERVER_DIR=$(shell pwd)/ui \
	PROXY_HTTPS_CERT=$(pwd)/rootCA.pem
	PROXY_HTTPS_CERT_KEY=$(pwd)/rootCA-key.pem
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

install-mkcert:
	go install ./vendor/filippo.io/mkcert

# generate-mkcerts will generate certificate and copy for dev to local dir
generate-mkcert:
	mkcert -install
	cp $(shell mkcert -CAROOT)/* ./


emulator-install:
	docker run --privileged --rm tonistiigi/binfmt --install all

emulator-remote:
	docker run --privileged --rm tonistiigi/binfmt --uninstall qemu-*
