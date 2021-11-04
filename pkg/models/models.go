package models

import "github.com/mjudeikis/unikiosk/pkg/grpc/models"

type KioskState struct {
	Content string
	Title   string
	SizeW   int
	SizeH   int

	State State
}

type State int64

const (
	StateRunning State = iota
	StateStopped
	StateUnknown
)

func (s State) String() string {
	switch s {
	case StateRunning:
		return "running"
	case StateStopped:
		return "stopped"
	}
	return "unknown"
}

func ProtoToKioskState(in *models.KioskState) KioskState {
	return KioskState{
		Content: in.Content,
		Title:   in.Title,
		SizeW:   int(in.SizeW),
		SizeH:   int(in.SizeH),
		State:   State(in.State),
	}
}
