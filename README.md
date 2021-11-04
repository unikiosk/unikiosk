# UniKiosk - Universal Kiosk

Universal kioks is minimal container image running webview instance,
controllable via GRPC api. You can reload, start, stop, resize window 
via remote calls.

Container initiates X session via xinit and start webview binary.