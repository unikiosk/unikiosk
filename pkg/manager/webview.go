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
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/webview/webview"
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/models"
	"github.com/unikiosk/unikiosk/pkg/queue"
	"github.com/unikiosk/unikiosk/pkg/store"
	"github.com/unikiosk/unikiosk/pkg/store/disk"
)

type kiosk struct {
	log    *zap.Logger
	config *config.Config

	isReady *atomic.Value

	queue queue.Queue
	w     webview.WebView
	store store.Store

	lock  sync.Mutex
	state models.KioskState
}

type Kiosk interface {
	Run(ctx context.Context) error
}

func New(log *zap.Logger, config *config.Config, queue queue.Queue, isReady *atomic.Value) (*kiosk, error) {
	store, err := disk.New(log, config)
	if err != nil {
		return nil, err
	}

	var s models.KioskState
	state, err := store.Get()
	fmt.Println(state)
	if err != nil || state == nil {
		log.Info("no state found - start fresh")
		s = models.KioskState{
			Content: config.DefaultURI,
			SizeW:   int(C.display_width()),
			SizeH:   int(C.display_height()),
			Title:   "UniKiosk",
		}
	} else {
		s = *state
	}

	err = store.Persist(s)
	if err != nil {
		log.Warn("failed to persist store, will not recover after restart", zap.Error(err))
		return nil, err
	}

	return &kiosk{
		log:     log,
		config:  config,
		queue:   queue,
		store:   store,
		isReady: isReady,
		//w is initiated in startOrRestore

		lock:  sync.Mutex{},
		state: s,
	}, nil
}

func (k *kiosk) startOrRestore() error {
	w := webview.New(true)
	k.w = w

	k.w.SetSize(k.state.SizeW, k.state.SizeH, webview.HintNone)

	k.w.Dispatch(func() {
		contentLog := k.state.Content
		if len(k.state.Content) > 50 {
			contentLog = k.state.Content[:50]
		}

		k.log.Info("open", zap.String("content", contentLog))
		w.Navigate(k.state.Content)
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

	for !k.isReady.Load().(bool) {
		k.log.Debug("waiting for subsystems to report ready")
	}

	// HACK: Webview is not loading page on start due to some race condition. Still need to get to the bottom of it
	// We emit event to re-load after 2s of the startup hope we will succeed :/
	go func() {
		time.Sleep(time.Second * 2)
		k.log.Info("emit", zap.String("content", k.state.Content))
		k.queue.Emit(models.KioskState{
			Content: k.state.Content + "?hack",
		})
	}()

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

	err := k.store.Persist(k.state)
	if err != nil {
		k.log.Warn("failed to persist store, will not recover after restart", zap.Error(err))
	}

}
