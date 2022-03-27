package api

import (
	"fmt"
	"strings"
)

type ScreenAction int

const (
	ScreenActionStart ScreenAction = iota
	ScreenActionUpdate
	ScreenActionStop
	ScreenActionPowerOff
	ScreenActionPowerOn
	ScreenActionScreenShot

	ScreenActionUnknown
)

func StringToAction(a string) (ScreenAction, error) {
	switch strings.ToLower(a) {
	case "start":
		return ScreenActionStart, nil
	case "update":
		return ScreenActionUpdate, nil
	case "stop":
		return ScreenActionStop, nil
	case "poweroff":
		return ScreenActionPowerOff, nil
	case "poweron":
		return ScreenActionPowerOn, nil
	case "screenshot":
		return ScreenActionScreenShot, nil
	default:
		return ScreenActionUnknown, fmt.Errorf("unknown action")
	}
}

func (s ScreenAction) String() string {
	switch s {
	case ScreenActionStart:
		return "start"
	case ScreenActionUpdate:
		return "update"
	case ScreenActionStop:
		return "stop"
	case ScreenActionPowerOff:
		return "poweroff"
	case ScreenActionPowerOn:
		return "poweron"
	case ScreenActionScreenShot:
		return "screenshot"
	default:
		return "unknown"
	}
}
