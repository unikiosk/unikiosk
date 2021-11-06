# UniKiosk - Universal Kiosk

Universal Kiosk (UniKiosk) is a minimal container image running webview kiosk with a built-in webserver,
controllable via GRPC API. You can reload, start, stop, resize window via remote calls.

Container initiates X session via xinit and starts the service. 

## Pre-requisites

1. Install RPI or any other device with ubuntu server image. No need for graphic environment
2. Install docker


## Run

Port 7000 is used for remote management (GRPC API). Mounted volume - for state management. Run using Docker:

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

## Roadmap

- [ ] Ability to provide application bundle-
- [ ] Ability to schedule changing URL's/bundles
- [ ] Ability to turn off/on view and maybe screen

## Development

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
