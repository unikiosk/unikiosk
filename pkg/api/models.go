package api

const StaticFilePrefix = "data:text/html"
const ContentTypeApplicationJSON = "application/json"

// KioskRequest represents request to interact with kiosk
type KioskRequest struct {
	Content string
	Title   string
	SizeW   int
	SizeH   int
	Action  ScreenAction
}

// KioskResponse represents response payload for the api
type KioskResponse struct {
	Content    string
	Title      string
	SizeW      int
	SizeH      int
	PowerState PowerState
	KioskMode  KioskMode
	// optional fields
	Screenshot []byte
}

// KioskState respresent current Kiosk state and is used for eventing and storage
type KioskState struct {
	Content string
	// ContentHash is hash of content before it was modified
	ContentHash string
	Title       string
	SizeW       int
	SizeH       int
	PowerState  PowerState
	KioskMode   KioskMode

	Screenshot []byte
}

// Kiosk mode defines mode of operations
type KioskMode string

var (
	// KioskModeDirect - use webview directly to render the pages
	KioskModeDirect KioskMode = "direct"
	// KioskModeProxy - render pages via built-in proxy
	KioskModeProxy KioskMode = "proxy"
)

// PowerState defines screen power state
type PowerState int

const (
	PowerStateOn PowerState = iota
	PowerStateOff
	PowerStateUnknown
)
