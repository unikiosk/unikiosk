package firefox

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image/png"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kbinani/screenshot"
	gofirefox "github.com/unikiosk/go-firefox"
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/api"
	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/eventer"
	"github.com/unikiosk/unikiosk/pkg/store"
	"github.com/unikiosk/unikiosk/pkg/util/recover"
	"github.com/unikiosk/unikiosk/pkg/util/shell"
)

var stateKey = "gofirefox"

type kiosk struct {
	log    *zap.Logger
	config *config.Config

	events  eventer.Eventer
	l       gofirefox.UI
	started atomic.Value
	store   store.Store
}

type Kiosk interface {
	Run(ctx context.Context) error
	Stop() error

	PowerOff() error
	PowerOn() error
	Screenshot() ([]byte, error)
}

func New(log *zap.Logger, config *config.Config, events eventer.Eventer, store store.Store) (*kiosk, error) {
	k := &kiosk{
		log:     log,
		config:  config,
		events:  events,
		store:   store,
		started: atomic.Value{},
		//w is initiated in startOrRestore
	}

	k.started.Store(false)

	// empty get to we set it on the first run
	// Id we don't have state - bootstrap with defaults
	state, err := k.store.Get(stateKey)
	if err != nil || state == nil {
		k.log.Info("no state found, new start")
		w, h := getScreenSize()
		s := api.KioskState{
			Content: k.config.DefaultWebServerURL,
			SizeW:   w,
			SizeH:   h,
			Title:   "UniKiosk",
		}
		err = k.store.Persist(stateKey, s)
		if err != nil {
			k.log.Warn("failed to persist store, will not recover after restart", zap.Error(err))
		}
	}

	return k, nil
}

func (k *kiosk) Run(ctx context.Context) error {
	k.log.Info("start firefox manager")

	// because lorca must run in main thread, we run dispatcher as separete thread.
	// dispatcher responsible for acting to grpc calls and updating the ser
	go k.runDispatcher(ctx)

	for {
		// TODO: Add context
		err := k.startOrRecover(ctx)
		if err != nil {
			k.log.Error("startOrRecover failed", zap.Error(err))
			time.Sleep(time.Second * 1)
		}

	}
}

func (k *kiosk) Stop() error {
	return k.l.Stop()
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

func (k *kiosk) startOrRecover(ctx context.Context) error {
	state, err := k.store.Get(stateKey)
	if err != nil {
		return fmt.Errorf("failed to get state: %s", err)
	}

	k.log.Info("set proxy", zap.String("http", k.config.ProxyHTTPServerAddr), zap.String("https", k.config.ProxyHTTPServerAddr))

	//user_pref("network.proxy.http" “10.10.10.200”);
	//user_pref("network.proxy.http_port”, 8080);
	//user_pref("network.proxy.ssl”, “10.10.10.200”);
	//user_pref("network.proxy.ssl_port”, 8080);
	//user_pref("network.proxy.type", 1);
	proxy := []string{
		"user_pref(\"network.proxy.type\", 1);",
		"user_pref(\"network.proxy.http\", \"localhost\");",
		fmt.Sprintf("user_pref(\"network.proxy.http_port\", %s);", strings.Replace(k.config.ProxyHTTPServerAddr, ":", "", -1)),
		"user_pref(\"network.proxy.ssl\", \"localhost\");",
		fmt.Sprintf("user_pref(\"network.proxy.ssl_port\", %s);", strings.Replace(k.config.ProxyHTTPSServerAddr, ":", "", -1)),
		"user_pref(\"security.insecure_field_warning.contextual.enabled\", False);",
		"user_pref(\"security.insecure_password.ui.enabled\", False);",
		"user_pref(\"dom.security.https_only_mode\", False);",
		"user_pref(\"security.enterprise_roots.enabled\", True);",
	}
	flags := []string{}
	ui, err := gofirefox.New(state.Content, flags, proxy)
	if err != nil {
		return fmt.Errorf("failed to start lorca: %s", err)
	}

	k.l = ui
	k.started.Store(true)

	return k.l.Run(ctx)
}

func (k *kiosk) runDispatcher(ctx context.Context) {
	defer recover.Panic(k.log)
	listener := k.events.Subscribe(ctx)

	for {
		if k.started.Load().(bool) {
			break
		}
	}

	for event := range listener {
		err := k.handle(ctx, event)
		if err != nil {
			k.log.Error("dispatch error", zap.Error(err))
		}
	}
}

// handle will handle events intended for lorca
func (k *kiosk) handle(ctx context.Context, event *eventer.EventWrapper) error {
	e := event.Payload

	callback := event.Callback
	hash := k.getURLHash(e.Request.Content)

	k.log.Info("execute action", zap.String("type", e.Request.Action.String()))
	switch e.Request.Action {
	case api.ScreenActionPowerOff:
		k.log.Info("lorca power off")
		err := k.PowerOff()
		if err != nil {
			return err
		}
		k.updateState(ctx, e.Request, "")
	case api.ScreenActionPowerOn:
		k.log.Info("lorca power on")
		err := k.PowerOn()
		if err != nil {
			return err
		}
		k.updateState(ctx, e.Request, hash)
	case api.ScreenActionScreenShot:
		k.log.Info("lorca screenShot")
		screen, err := k.Screenshot()
		if err != nil {
			return err
		}
		err = k.updateLastScreenshot(ctx, screen)
		if err != nil {
			return err
		}
	case api.ScreenActionUpdate:
		k.log.Info("lorca update")
		err := k.updateState(ctx, e.Request, hash)
		if err != nil {
			return err
		}
	}

	// once steps has been handled - get state and return
	state, err := k.store.Get(stateKey)
	if err != nil {
		return err
	}

	result := &eventer.EventWrapper{
		Payload: api.Event{
			Response: api.KioskResponse{
				Content:    state.Content,
				Title:      state.Title,
				SizeW:      state.SizeW,
				SizeH:      state.SizeH,
				PowerState: state.PowerState,
				KioskMode:  state.KioskMode,
				Screenshot: state.Screenshot,
			},
		},
	}

	err = k.updateState(ctx, e.Request, hash)
	if err != nil {
		k.log.Error("error while updating the state", zap.Error(err))
	}

	callback <- result
	return nil
}

func (k *kiosk) getURLHash(in string) string {
	h := sha256.New()
	h.Write([]byte(in))

	return string(h.Sum(nil))
}

func (k *kiosk) updateState(ctx context.Context, in api.KioskRequest, urlHash string) error {
	state, err := k.store.Get(stateKey)
	if err != nil {
		k.log.Error("failed to get state", zap.Error(err))
		return err
	}

	// Dispatch is async, so we need to persist inside of it :/ this is not ideal as context are mixed
	if in.Content != "" && urlHash != state.ContentHash {
		k.l.Load(in.Content)

		state.Content = in.Content
		state.ContentHash = urlHash
	}
	if state.SizeW != in.SizeW || state.SizeH != in.SizeH {
		state.SizeW = in.SizeW
		state.SizeH = in.SizeH
	}

	if in.Action.String() == api.ScreenActionPowerOff.String() {
		state.PowerState = api.PowerStateOff
	}
	if in.Action.String() == api.ScreenActionPowerOn.String() {
		state.PowerState = api.PowerStateOn
	}

	err = k.store.Persist(stateKey, *state)
	if err != nil {
		k.log.Warn("failed to persist store, will not recover after restart", zap.Error(err))
		return err
	}
	return nil
}

func (k *kiosk) updateLastScreenshot(ctx context.Context, screen []byte) error {
	state, err := k.store.Get(stateKey)
	if err != nil {
		return err
	}
	state.Screenshot = screen

	err = k.store.Persist(stateKey, *state)
	if err != nil {
		k.log.Warn("failed to persist store, will not recover after restart", zap.Error(err))
	}
	return err
}
