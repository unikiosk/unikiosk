package lorca

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image/png"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/zserge/lorca"
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/api"
	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/eventer"
	"github.com/unikiosk/unikiosk/pkg/store"
	"github.com/unikiosk/unikiosk/pkg/util/recover"
	"github.com/unikiosk/unikiosk/pkg/util/shell"
)

var stateKey = "lorca"

type kiosk struct {
	log    *zap.Logger
	config *config.Config

	events eventer.Eventer
	l      lorca.UI
	store  store.Store
}

type Kiosk interface {
	Run(ctx context.Context) error
	PowerOff() error
	PowerOn() error
	Screenshot() ([]byte, error)
	Close() error
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

func (k *kiosk) Close() error {
	return k.l.Close()
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
	_, sErr, err = shell.Exec("xset -dpms")
	if err != nil {
		return err
	}
	if sErr != "" {
		return fmt.Errorf(sErr)
	}

	return err
}

func (k *kiosk) Screenshot() ([]byte, error) {
	n := screenshot.NumActiveDisplays()
	// TODO: add support for more than 1 display
	if n == 0 {
		return nil, fmt.Errorf("no screen found")
	}

	bounds := screenshot.GetDisplayBounds(0)

	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, err
	}

	var image []byte
	buf := bytes.NewBuffer(image)

	err = png.Encode(buf, img)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (k *kiosk) startOrRestore() error {
	k.log.Info("set proxy", zap.String("http", k.config.DefaultHTTPProxyURL), zap.String("https", k.config.DefaultHTTPSProxyURL))

	state := k.getCurrentState()

	sandbox := "--no-sandbox" // mandatory
	fullScreen := "--start-fullscreen"
	proxyHTTPS := fmt.Sprintf("--proxy-server=%s", k.config.DefaultHTTPSProxyURL)

	ui, err := lorca.New("https://grafana.webrelay.io/playlists/play/1?kiosk", "", int(state.SizeW), int(state.SizeH), sandbox, fullScreen, proxyHTTPS)
	if err != nil {
		return err
	}
	k.l = ui

	<-k.l.Done()

	return nil
}

func (k *kiosk) Run(ctx context.Context) error {
	k.log.Info("start lorca manager")

	// because webview must run in main thread, we run dispatcher as separete thread.
	// dispatcher responsible for acting to grpc calls and updating the ser
	go k.runDispatcher(ctx)

	// HACK: Webview is not loading page on start due to some race condition. Still need to get to the bottom of it
	// We emit event to re-load after 2s of the startup hope we will succeed :/
	go func() {
		defer recover.Panic(k.log)

		time.Sleep(time.Second * 5)
		state := k.getCurrentState()
		k.log.Info("emit", zap.String("content", state.Content))
		_, _ = k.events.Emit(&eventer.EventWrapper{
			Payload: api.Event{
				Type: api.EventTypeWebViewUpdate,
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
		// TODO: Add context
		err := k.startOrRestore()
		if err != nil {
			k.log.Error("startOrRestore failed", zap.Error(err))
			time.Sleep(time.Second * 1)
		}

	}
}

func (k *kiosk) runDispatcher(ctx context.Context) {
	defer recover.Panic(k.log)
	listener := k.events.Subscribe(ctx)

	for event := range listener {
		err := k.dispatch(ctx, event)
		if err != nil {
			k.log.Error("dispatch error", zap.Error(err))
		}
	}
}

func (k *kiosk) dispatch(ctx context.Context, event *eventer.EventWrapper) error {
	e := event.Payload
	if e.Type != api.EventTypeWebViewUpdate && e.Type != api.EventTypeAction {
		return nil
	}

	callback := event.Callback

	hash := k.getURLHash(e.Request.Content)

	// act only on requests to reload webview
	if e.Type == api.EventTypeWebViewUpdate {
		k.log.Info("webview reload")

		k.updateState(ctx, e.Request, hash)
	}

	// if power action event
	if e.Type == api.EventTypeAction {
		k.log.Info("execute action", zap.String("type", e.Request.Action.String()))
		switch e.Request.Action {
		case api.ScreenActionPowerOff:
			err := k.PowerOff()
			if err != nil {
				return err
			}
			k.updateState(ctx, e.Request, "")
		case api.ScreenActionPowerOn:
			err := k.PowerOn()
			if err != nil {
				return err
			}
			k.updateState(ctx, e.Request, "")
		case api.ScreenActionScreenShot:
			screen, err := k.Screenshot()
			if err != nil {
				return err
			}
			err = k.updateLastScreenshot(ctx, screen)
			if err != nil {
				return err
			}
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
				Screenshot:       state.Screenshot,
			},
		},
	}
	callback <- result
	return nil
}

func (k *kiosk) getURLHash(in string) string {
	h := sha256.New()
	h.Write([]byte(in))

	return string(h.Sum(nil))
}

func (k *kiosk) updateState(ctx context.Context, in api.KioskRequest, urlHash string) {
	state := k.getCurrentState()

	// Dispatch is async, so we need to persist inside of it :/ this is not ideal as context are mixed
	if in.Content != "" && urlHash != state.ContentHash {

		k.l.Load(in.Content)

		state.Content = in.Content
		state.ContentHash = urlHash
	}

	if in.Action.String() == api.ScreenActionPowerOff.String() {
		state.ScreenPowerState = api.ScreenPowerStateOff
	}
	if in.Action.String() == api.ScreenActionPowerOn.String() {
		state.ScreenPowerState = api.ScreenPowerStateOn
	}

	err := k.store.Persist(stateKey, state)
	if err != nil {
		k.log.Warn("failed to persist store, will not recover after restart", zap.Error(err))
	}
}

func (k *kiosk) updateLastScreenshot(ctx context.Context, screen []byte) error {
	state := k.getCurrentState()
	state.Screenshot = screen

	err := k.store.Persist(stateKey, state)
	if err != nil {
		k.log.Warn("failed to persist store, will not recover after restart", zap.Error(err))
	}
	return err
}
