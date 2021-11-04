# UniKiosk - Universal Kiosk

Universal kioks is minimal container image running webview instance,
controllable via GRPC api. You can reload, start, stop, resize window 
via remote calls.

Container initiates X session via xinit and start webview system


# Run

Start server using `go run`:
```
go run ./cmd/screen
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
docker run -p 7000:7000 quay.io/mangirdas/unikiosk 
```

Interact with container:
```
make build-cli
./release/cli --url https://synpse.net
```