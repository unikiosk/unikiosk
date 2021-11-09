package api

type Event struct {
	Type      EventType
	KioskMode KioskMode
	Request   KioskRequest
	Response  KioskResponse
}

type EventType int64

const (
	// EventTypeProxyUpdate will require chaning proxy destination and reload. And trigger reload on webview.
	EventTypeProxyUpdate EventType = iota
	// EventTypeWebViewUpdate - will need only Webview reload.
	// In the future we might want to server this from application bundle using our webserver, so flow stays the same
	EventTypeWebViewUpdate

	// EventTypePowerAction will power off or on the screen
	EventTypePowerAction
)
