# UniKiosk - Universal Kiosk

Universal Kiosk (UniKiosk) is a minimal container image running `webview` kiosk with a built-in webserver and proxy,
controllable via GRPC API. You can reload, start, stop, resize window via remote calls.

Container initiates X session via xinit and starts the service so you don't need any X environment running on the target server!

## Pre-requisites

1. Install RPI or any other device with ubuntu server image. No need for graphic environment
2. Install docker


## Run

Port 7000 is used for remote management (GRPC API). Mounted volume - for state management. 
Run using Docker:

```
docker run -d -v /data/unikiosk:/data -p 7000:7000 quay.io/unikiosk/unikiosk 
```

Change default webpage:
```
./release/cli --url https://synpse.net
```

In addition you can provide your own single page to show:
```
./release/cli --file examples/index.html
```

## Configuration options

Al available options available at `pkg/config.go`

```
# custom headers set by proxy. Very useful for CORS or NOC dashboards when you want to inject
# credentials into requests in the form of headers
PROXY_HEADERS="red:1,blue:2"

# log level
LOG_LEVEL=info

# Kiosk mode. Options: direct, proxy. This option will tell if to load webpage directly using webview or via proxy. 
# Some webpages do not like being served via reverse proxy due to fact how they handle assets urls (Jenkins and Prometheus are good examples).
# in cases like this you can use direct mode.
KIOSK_MODE=proxy
```

## Roadmap

- [ ] Ability to provide application bundle
- [ ] Ability to schedule changing URL's/bundles
- [ ] Ability to turn off/on view and maybe screen
- [ ] Ability to resize window via CLI (Server already supports this)
- [ ] Move to https://pkg.go.dev/github.com/goproxy/goproxy for better proxy'ing and caching

## Development

To run binary locally you will need to have few installed. This was tested on Ubuntu and Fedora.
Full dependency list can be found in `dockerfiles/Dockerfile`

WebKit development dependencies:
```
libwebkit2gtk-4.0-dev build-essential xinit libglib2.0-dev 
```

Start server using `go run`:
```
make run
```

You should see default landing page

1. Update page to URL:

```
go run ./cmd/cli --url https://synpse.net
```

2. Update page with local file content:

```
go run ./cmd/cli/ --file example/index.html
```

Run using docker:
```
docker run -v /data/unikiosk:/data -p 7000:7000 quay.io/unikiosk/unikiosk 
```

Interact with container:
```
make build-cli
./release/cli --url https://synpse.net
```
