# UniKiosk - Universal Kiosk

Universal Kiosk (UniKiosk) is a minimal container image running `lorca` kiosk with a built-in webserver and proxy,
controllable via API. You can reload, start, stop, resize window via remote calls.

Container initiates X session via xinit and starts the service so you don't need any X environment running on the target server!

## High level overview

Any url, provided to UniKiosk is routed via internal MITM proxy. This way we can do offline caching, certificate spoofing and other
interesting things to make sure content is represented as it should be. 

```
Lorca -> proxy -> internet
```

When UniKiosk is started it serves initial web content from built-in webserver (same process which runs API)
```
Lorca -> proxy -> internal web server
```

`pkg/web` api accepts requests and routes them via event hub. 

## Pre-requisites

1. Install RPI or any other device with ubuntu server image. No need for graphic environment
2. Install docker

## Run

Port 8081 is used for remote management (API). Mounted volume - for state management. 
Run using Docker:

```
docker run -d -v /data/unikiosk:/data -p 8081:8081 quay.io/unikiosk/unikiosk 
```

Change default webpage:
```
./release/cli set --url https://synpse.net
```

In addition you can provide your own single page to show:
```
./release/cli set --file examples/index.html
```

## Configuration options

Al available options available at `pkg/config.go`

```
# custom headers set by proxy. Very useful for CORS or NOC dashboards when you want to inject
# credentials into requests in the form of headers
PROXY_HEADERS="red:1,blue:2"

# log level
LOG_LEVEL=info
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
go run ./cmd/cli set --url https://synpse.net
```

2. Update page with local file content:

```
go run ./cmd/cli/ set --file example/index.html
```

Run using docker:
```
docker run --privileged -v /data:/data -p 8081:8081 quay.io/unikiosk/unikiosk 
```

Interact with container:
```
make build-cli
./release/cli set --url https://synpse.net
```

## Debug

You will be able to build binaries with `make build` and test them in the container environment so avoiding installing dependencies locally.
You might need to install `xhost` and grant access to docker:
```
export DISPLAY=:0.0
xhost +local:docker
```

Run using docker:
```
docker run  -e DISPLAY=:0 --net=host  -v /tmp/.X11-unix:/tmp/.X11-unix -v $(pwd):/go -it --entrypoint bash --privileged quay.io/unikiosk/unikiosk
```
