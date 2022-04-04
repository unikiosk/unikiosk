package firefox

import (
	"github.com/kbinani/screenshot"
)

// TODO: Support multiple displays
func getScreenSize() (int, int) {
	bounds := screenshot.GetDisplayBounds(0)
	x, y := bounds.Dx(), bounds.Dy()

	return x, y
}
