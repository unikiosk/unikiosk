package models

import "github.com/unikiosk/unikiosk/pkg/grpc/models"

const StaticFilePrefix = "data:text/html"

type KioskState struct {
	Content string
	Title   string
	SizeW   int
	SizeH   int
	Action  ScreenAction
}

type State int64

func ProtoToKioskState(in *models.KioskState) KioskState {
	return KioskState{
		Content: in.Content,
		Title:   in.Title,
		SizeW:   int(in.SizeW),
		SizeH:   int(in.SizeH),
		Action:  ScreenAction(in.Action),
	}
}

type Event struct {
	Type      EventType
	KioskMode KioskMode
	Payload   KioskState
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

type KioskMode string

var (
	// KioskModeDirect - use webview directly to render the pages
	KioskModeDirect KioskMode = "direct"
	// KioskModeProxy - render pages via built-in proxy
	KioskModeProxy KioskMode = "proxy"
)

type ScreenAction string

var (
	ScreenActionStart    ScreenAction = "start"
	ScreenActionUpdate   ScreenAction = "update"
	ScreenActionStop     ScreenAction = "stop"
	ScreenActionPowerOff ScreenAction = "poweroff"
	ScreenActionPowerOn  ScreenAction = "poweron"
)

func ProtoToScreenAction(in *models.EnumScreenAction) ScreenAction {
	switch int(*in) {
	case 0:
		return ScreenActionStart
	case 1:
		return ScreenActionUpdate
	case 2:
		return ScreenActionStop
	case 3:
		return ScreenActionPowerOff
	case 4:
		return ScreenActionPowerOn
	default:
		return ScreenActionUpdate // update is default....
	}
}
