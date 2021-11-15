package webview

import (
	"context"
	"crypto/sha256"
	"fmt"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/webview/webview"
	gowebview "github.com/webview/webview"
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/api"
	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/eventer"
	"github.com/unikiosk/unikiosk/pkg/store"
	"github.com/unikiosk/unikiosk/pkg/util/recover"
	"github.com/unikiosk/unikiosk/pkg/util/shell"
)

var webViewStateKey = "webview"

type kiosk struct {
	log    *zap.Logger
	config *config.Config

	events eventer.Eventer
	w      gowebview.WebView
	store  store.Store

	lock sync.Mutex
}

type Kiosk interface {
	Run(ctx context.Context) error
	PowerOff() error
	PowerOn() error
	Screenshot() error
	Close()
}

func New(log *zap.Logger, config *config.Config, events eventer.Eventer, store store.Store) (*kiosk, error) {
	k := &kiosk{
		log:    log,
		config: config,
		events: events,
		store:  store,
		//w is initiated in startOrRestore
	}

	// empty get to we set it on the first run
	_ = k.getCurrentState()
	return k, nil
}

func (k *kiosk) Close() {
	k.w.Destroy()
}

// TODO: Power off/on should be better done via CGO
// https://stackoverflow.com/questions/60477195/turning-off-monitor-in-c
// But somehow I didn't managed to get right bindings.

// PowerOff - powers off the screen
func (k *kiosk) PowerOff() error {
	// xset -display :0.0 dpms force off
	_, _, err := shell.Exec("xset -display :0.0 dpms force off")
	return err
}

// PowerOn - powers on the screen
func (k *kiosk) PowerOn() error {
	k.log.Debug("execute powerOn")
	// xset -display :0.0 dpms force off
	_, sErr, err := shell.Exec("xset -display :0.0 dpms force on")
	if err != nil {
		return err
	}
	if sErr != "" {
		return fmt.Errorf(sErr)
	}
	// this prevents blanking of the screen after it gets on
	_, sErr, err = shell.Exec("xset -display :0.0 s off")
	if err != nil {
		return err
	}
	if sErr != "" {
		return fmt.Errorf(sErr)
	}
	_, sErr, err = shell.Exec("xset -display :0.0 s noblank")
	if err != nil {
		return err
	}
	if sErr != "" {
		return fmt.Errorf(sErr)
	}

	return err
}

func (k *kiosk) Screenshot() error {
	n := screenshot.NumActiveDisplays()

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)

		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			panic(err)
		}
		fileName := fmt.Sprintf(filepath.Join(k.config.StateDir, "%d_%dx%d.png"), i, bounds.Dx(), bounds.Dy())
		file, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer file.Close()
		png.Encode(file, img)

		fmt.Printf("#%d : %v \"%s\"\n", i, bounds, fileName)
	}

	return nil
}

func (k *kiosk) startOrRestore() error {
	w := webview.New(true)
	k.w = w

	state := k.getCurrentState()

	k.w.SetSize(int(state.SizeW), int(state.SizeH), webview.HintNone)

	k.w.Dispatch(func() {
		contentLog := state.Content
		if len(state.Content) > 50 {
			contentLog = state.Content[:50]
		}

		k.log.Info("open", zap.String("content", contentLog))
		w.Navigate(state.Content)
	})

	k.w.Run()

	return nil
}

func (k *kiosk) Run(ctx context.Context) error {
	k.log.Info("start webview manager")

	// because webview must run in main thread, we run dispatcher as separete thread.
	// dispatcher responsible for acting to grpc calls and updating the ser
	go k.runDispatcher(ctx)

	// HACK: Webview is not loading page on start due to some race condition. Still need to get to the bottom of it
	// We emit event to re-load after 2s of the startup hope we will succeed :/
	go func() {
		defer recover.Panic(k.log)

		time.Sleep(time.Second * 2)
		state := k.getCurrentState()
		k.log.Info("emit", zap.String("content", state.Content))
		_, _ = k.events.Emit(&eventer.EventWrapper{
			Payload: api.Event{
				Type:      api.EventTypeWebViewUpdate,
				KioskMode: api.KioskModeDirect,
				Request: api.KioskRequest{
					Content: state.Content,
					Title:   state.Title,
					SizeW:   state.SizeW,
					SizeH:   state.SizeH,
				},
			},
		})
	}()

	for {
		k.startOrRestore()
	}
}

func (k *kiosk) runDispatcher(ctx context.Context) error {
	defer recover.Panic(k.log)
	listener := k.events.Subscribe(ctx)

	for event := range listener {
		e := event.Payload
		callback := event.Callback

		hash := k.getURLHash(e.Request.Content)

		// act only on requests to reload webview
		if e.Type == api.EventTypeWebViewUpdate && e.KioskMode == api.KioskModeDirect {
			k.log.Info("direct webview reload")

			k.updateState(ctx, e.Request, hash)
		}
		// act only on requests to reload webview iin proxy mode
		if e.Type == api.EventTypeWebViewUpdate && e.KioskMode == api.KioskModeProxy {
			k.log.Info("proxy webview reload")
			url, err := k.getProxyURL(event.Payload.Request)
			if err != nil {
				return err
			}
			e.Request.Content = url

			k.updateState(ctx, e.Request, hash)
		}
		// if power action event
		if e.Type == api.EventTypePowerAction {
			k.log.Info("power action", zap.String("action", e.Request.Action.String()))
			if e.Request.Action == api.ScreenActionPowerOff {
				k.PowerOff()
				k.updateState(ctx, e.Request, "")
			}
			if e.Request.Action == api.ScreenActionPowerOn {
				k.PowerOn()
				k.updateState(ctx, e.Request, "")
			}
		}

		state := k.getCurrentState()

		result := &eventer.EventWrapper{
			Payload: api.Event{
				Response: api.KioskResponse{
					Content:          state.Content,
					Title:            state.Title,
					SizeW:            state.SizeW,
					SizeH:            state.SizeH,
					ScreenPowerState: state.ScreenPowerState,
					KioskMode:        state.KioskMode,
				},
			},
		}

		callback <- result
	}
	return nil
}

func (k *kiosk) getProxyURL(in api.KioskRequest) (string, error) {
	// replace original URL with proxy address while preserving path
	// in addition tls is handled in proxy, so drop https
	targetUrl, err := url.Parse(in.Content)
	if err != nil {
		return "", err
	}
	proxyUrl, err := url.Parse(k.config.DefaultProxyURL)
	if err != nil {
		return "", err
	}

	targetUrl.Host = proxyUrl.Host
	return strings.Replace(targetUrl.String(), "https", "http", 1), nil
}

func (k *kiosk) getURLHash(in string) string {
	h := sha256.New()
	h.Write([]byte(in))

	return string(h.Sum(nil))
}

func (k *kiosk) updateState(ctx context.Context, in api.KioskRequest, urlHash string) {
	state := k.getCurrentState()

	// Dispatch is async, so we need to persist inside of it :/ this is not ideal as context are mixed
	k.w.Dispatch(func() {
		if in.Content != "" && urlHash != state.ContentHash {
			k.log.Info("refresh webview", zap.String("incoming", in.Content), zap.String("current", state.Content))
			k.w.Navigate(in.Content)

			state.Content = in.Content
			state.ContentHash = urlHash
		}

		if in.Title != "" && state.Title != in.Title {
			k.w.SetTitle(in.Title)
			state.Title = in.Title
		}

		var changed bool
		if in.SizeW != 0 && state.SizeW != in.SizeW {
			state.SizeW = in.SizeW
			changed = true
		}
		if in.SizeH != 0 && state.SizeH != in.SizeH {
			state.SizeH = in.SizeH
			changed = true
		}
		if changed {
			k.w.SetSize(int(state.SizeW), int(state.SizeH), webview.HintNone)
		}

		if in.Action.String() == api.ScreenActionPowerOff.String() {
			state.ScreenPowerState = api.ScreenPowerStateOff
		}
		if in.Action.String() == api.ScreenActionPowerOn.String() {
			state.ScreenPowerState = api.ScreenPowerStateOn
		}

		err := k.store.Persist(webViewStateKey, state)
		if err != nil {
			k.log.Warn("failed to persist store, will not recover after restart", zap.Error(err))
		}
	})
}
