package models

import "github.com/unikiosk/unikiosk/pkg/grpc/models"

const StaticFilePrefix = "data:text/html"

type KioskState struct {
	Content string
	Title   string
	SizeW   int
	SizeH   int
}

type State int64

func ProtoToKioskState(in *models.KioskState) KioskState {
	return KioskState{
		Content: in.Content,
		Title:   in.Title,
		SizeW:   int(in.SizeW),
		SizeH:   int(in.SizeH),
	}
}

type Event struct {
	Type    EventType
	Payload KioskState
}

type EventType int64

const (
	// EventTypeProxyUpdate will require chaning proxy destination and reload. And trigger reload on webview.
	EventTypeProxyUpdate EventType = iota
	// EventTypeWebViewUpdate - will need only Webview reload.
	// In the future we might want to server this from application bundle using our webserver, so flow stays the same
	EventTypeWebViewUpdate
)
