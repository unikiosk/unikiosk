package api

type ScreenAction int

const (
	ScreenActionStart ScreenAction = iota
	ScreenActionUpdate
	ScreenActionStop
	ScreenActionPowerOff
	ScreenActionPowerOn
	ScreenActionScreenShot
)

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
