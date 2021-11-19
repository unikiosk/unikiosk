package lorca

import (
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/api"
)

/*
#cgo darwin LDFLAGS: -framework CoreGraphics -L/usr/include/X11 -lextensions
#cgo linux pkg-config: x11 xkbcommon x11-xcb
#if defined(__APPLE__)
#include <CoreGraphics/CGDisplayConfiguration.h>
int display_width() {
	return CGDisplayPixelsWide(CGMainDisplayID());
}
int display_height() {
	return CGDisplayPixelsHigh(CGMainDisplayID());
}
#elif defined(_WIN32)
#include <wtypes.h>
int display_width() {
	RECT desktop;
	const HWND hDesktop = GetDesktopWindow();
	GetWindowRect(hDesktop, &desktop);
	return desktop.right;
}
int display_height() {
	RECT desktop;
	const HWND hDesktop = GetDesktopWindow();
	GetWindowRect(hDesktop, &desktop);
	return desktop.bottom;
}
#else
#include <X11/Xlib.h>
int display_width() {
	Display* d = XOpenDisplay(NULL);
	Screen*  s = DefaultScreenOfDisplay(d);
	return s->width;
}
int display_height() {
	Display* d = XOpenDisplay(NULL);
	Screen*  s = DefaultScreenOfDisplay(d);
	return s->height;
}
#endif
*/
import "C"

func (k *kiosk) getCurrentState() api.KioskState {
	state, err := k.store.Get(stateKey)
	if err != nil || state == nil {
		k.log.Info("no state found - start fresh")
		s := api.KioskState{
			Content: k.config.DefaultWebServerURL,
			SizeW:   int64(C.display_width()),
			SizeH:   int64(C.display_height()),
			Title:   "UniKiosk",
		}
		err = k.store.Persist(stateKey, s)
		if err != nil {
			k.log.Warn("failed to persist store, will not recover after restart", zap.Error(err))
		}
		return s
	} else {
		return *state
	}
}
