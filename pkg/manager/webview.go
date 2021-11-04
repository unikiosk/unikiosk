package manager

/*
#cgo darwin LDFLAGS: -framework CoreGraphics
#cgo linux pkg-config: x11
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

import (
	"context"
	"sync"

	"github.com/mjudeikis/unikiosk/pkg/config"
	"github.com/mjudeikis/unikiosk/pkg/models"
	"github.com/mjudeikis/unikiosk/pkg/queue"
	"github.com/webview/webview"
	"go.uber.org/zap"
)

type kiosk struct {
	log    *zap.Logger
	config *config.Config

	queue queue.Queue
	w     webview.WebView

	lock  sync.Mutex
	state models.KioskState
}

type Kiosk interface {
	Run(ctx context.Context) error
}

func New(log *zap.Logger, config *config.Config, queue queue.Queue) *kiosk {

	s := models.KioskState{
		Content: "",
		SizeW:   int(C.display_width()),
		SizeH:   int(C.display_height()),
		Title:   "Synpse.net",
	}

	return &kiosk{
		log:    log,
		config: config,
		queue:  queue,
		//w is initiated in startOrRestore

		lock:  sync.Mutex{},
		state: s,
	}
}

func (k *kiosk) startOrRestore() error {
	w := webview.New(true)
	k.w = w

	k.w.SetSize(k.state.SizeW, k.state.SizeH, webview.HintNone)

	k.w.Dispatch(func() {

		files := []string{
			"assets/index.html",
		}

		for _, file := range files {
			d, err := assets.ReadFile(file)
			if err != nil {
				k.log.Error("fails to open", zap.Error(err))
			}

			w.Navigate(`data:text/html,
    ` + string(d) + `
    `)
		}
	})

	k.w.Run()
	defer k.w.Destroy()

	return nil
}

func (k *kiosk) Run(ctx context.Context) error {
	k.log.Info("start webview manager")

	// because webview must run in main thread, we run dispatcher as separete thread.
	// dispatcher responsible for acting to grpc calls and updating the ser
	go k.runDispatcher(ctx)

	for {
		k.startOrRestore()
	}

}

func (k *kiosk) runDispatcher(ctx context.Context) error {
	listener := k.queue.Listen()

	for event := range listener {
		k.updateState(ctx, event)
	}
	return nil
}

func (k *kiosk) updateState(ctx context.Context, state models.KioskState) {
	k.lock.Lock()
	defer k.lock.Unlock()

	if state.Content != "" && k.state.Content != state.Content {
		k.w.Navigate(state.Content)
		k.state.Content = state.Content
	}

	if state.Title != "" && k.state.Title != state.Title {
		k.w.Navigate(state.Title)
		k.state.Title = state.Title
	}

	var changed bool
	if state.SizeW != 0 && k.state.SizeW != state.SizeW {
		k.state.SizeW = state.SizeW
	}
	if state.SizeH != 0 && k.state.SizeH != state.SizeH {
		k.state.SizeH = state.SizeH
	}
	if changed {
		k.w.SetSize(k.state.SizeW, k.state.SizeH, webview.HintNone)
	}
}
