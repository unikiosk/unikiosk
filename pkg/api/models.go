package api

import "github.com/unikiosk/unikiosk/pkg/grpc/models"

const StaticFilePrefix = "data:text/html"

// KioskState respresent current Kiosk state
type KioskState struct {
	Content string
	// ContentHash is hash of content before it was modified
	ContentHash      string
	Title            string
	SizeW            int64
	SizeH            int64
	ScreenPowerState ScreenPowerState
	KioskMode        KioskMode

	Screenshot []byte
}

// KioskRequest represents request to interact with kiosk. It should match protoc spec
type KioskRequest struct {
	Content string
	Title   string
	SizeW   int64
	SizeH   int64
	Action  ScreenAction
}

func ProtoKioskRequestToModels(in *models.KioskRequest) KioskRequest {
	return KioskRequest{
		Content: in.Content,
		Title:   in.Title,
		SizeW:   in.SizeW,
		SizeH:   in.SizeH,
		Action:  ScreenAction(in.Action),
	}
}

// KioskResponse represents response payload for the api
type KioskResponse struct {
	Content          string
	Title            string
	SizeW            int64
	SizeH            int64
	ScreenPowerState ScreenPowerState
	KioskMode        KioskMode
	Screenshot       []byte
}

// Kiosk mode defines mode of operations
type KioskMode string

var (
	// KioskModeDirect - use webview directly to render the pages
	KioskModeDirect KioskMode = "direct"
	// KioskModeProxy - render pages via built-in proxy
	KioskModeProxy KioskMode = "proxy"
)

type ScreenPowerState int64

const (
	ScreenPowerStateOn ScreenPowerState = iota
	ScreenPowerStateOff
	ScreenPowerStateUnknown
)
