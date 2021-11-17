package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/api"
	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/eventer"
	"github.com/unikiosk/unikiosk/pkg/store"
	"github.com/unikiosk/unikiosk/pkg/util/recover"
)

var proxyStateKey = "proxy"

type Proxy interface {
	Run(ctx context.Context) error
}

type proxy struct {
	config *config.Config
	server *http.Server

	store  store.Store
	events eventer.Eventer

	ready atomic.Value

	targetURLMu *sync.RWMutex
	targetURL   *url.URL

	log *zap.Logger
}

func New(ctx context.Context, log *zap.Logger, config *config.Config, events eventer.Eventer, store store.Store) (*proxy, error) {
	// check if we have state, if not - default
	var u string
	state, err := store.Get(proxyStateKey)
	if err != nil || state == nil {
		log.Info("no state found - start fresh")
		u = config.DefaultWebServerURL
	} else {
		u = state.Content
	}

	defaultURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("failed to parse upstream address: %w", err)
	}

	p := proxy{
		log:       log,
		config:    config,
		events:    events,
		targetURL: defaultURL,
		store:     store,

		ready:       atomic.Value{},
		targetURLMu: &sync.RWMutex{},
	}

	h, err := p.getProxyHandler(defaultURL)
	if err != nil {
		return nil, err
	}

	p.server = &http.Server{
		Addr:    p.config.ProxyServerAddr,
		Handler: h,
	}

	go func() {
		defer recover.Panic(p.log)

		err := p.runSync(ctx)
		if err != nil {
			p.log.Debug("failed proxy reload", zap.Error(err))
		}
	}()

	return &p, nil
}

func (p *proxy) getProxyHandler(u *url.URL) (http.Handler, error) {
	rp := &httputil.ReverseProxy{
		Transport: &http.Transport{DialTLS: dialTLS},
	}

	if u.Scheme == "https" {
		// Set a custom DialTLS to access the TLS connection state
		rp.Transport = &http.Transport{DialTLS: dialTLS}
	}

	rp.Director = p.Director

	r := mux.NewRouter()
	r.Use(handlers.CompressHandler)

	r.PathPrefix("/").Handler(rp)
	return r, nil
}

func (p *proxy) Director(req *http.Request) {
	spew.Dump(req.Host)
	for k, v := range p.config.ProxyHeaders {
		req.Header.Set(k, v)
	}

	p.targetURLMu.RLock()
	defer p.targetURLMu.RUnlock()

	req.URL.Scheme = p.targetURL.Scheme
	req.URL.Host = p.targetURL.Host
	req.Host = p.targetURL.Host

	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
}

func dialTLS(network, addr string) (net.Conn, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	cfg := &tls.Config{ServerName: host}

	tlsConn := tls.Client(conn, cfg)
	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	cs := tlsConn.ConnectionState()
	cert := cs.PeerCertificates[0]

	_ = cert.VerifyHostname(host)

	return tlsConn, nil
}

func (p *proxy) Run(ctx context.Context) error {
	p.ready.Store(true)
	p.log.Info("Proxy running",
		zap.String("listening", p.config.ProxyServerAddr),
		zap.String("destination", p.targetURL.String()),
	)

	for {
		// lock while we run
		p.ready.Store(true)

		err := p.server.ListenAndServe()
		if err != nil {
			p.log.Debug("proxy failed. Restarting", zap.Error(err))
			time.Sleep(time.Second)
		}
		// once failed release lock sp we can hot swap the server
		p.log.Debug("proxy restarting", zap.Error(err))
	}
}

func (p *proxy) Stop(ctx context.Context) error {
	p.log.Info("stopping proxy")
	p.ready.Store(false)
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return p.server.Shutdown(ctxTimeout)
}

func (p *proxy) runSync(ctx context.Context) error {
	listener := p.events.Subscribe(ctx)

	for event := range listener {
		p.log.Debug("reload proxy")
		if event.Payload.Type != api.EventTypeProxyUpdate {
			continue
		}

		e := event.Payload
		callback := event.Callback

		// create new server object and override existing one
		u, err := url.Parse(e.Request.Content)
		if err != nil {
			p.log.Error("failed to parse target URL",
				zap.String("url", e.Request.Content),
				zap.Error(err),
			)
			continue
		}
		p.targetURLMu.Lock()
		p.targetURL = u
		p.targetURLMu.Unlock()

		err = p.store.Persist(proxyStateKey, api.KioskState{
			Content: u.String(),
		})
		if err != nil {
			p.log.Warn("failed to persist proxy state, will fails to recover")
		}

		// override webview back to proxy as we might be in file serve mode
		_, err = p.events.Emit(&eventer.EventWrapper{
			Payload: api.Event{
				Type: api.EventTypeWebViewUpdate,
				Request: api.KioskRequest{
					Content: p.targetURL.String(),
				},
			},
			// pipe in callback
			Callback: callback,
		})
		if err != nil {
			p.log.Error("failed to emit webview update event",
				zap.Error(err),
			)
		}
	}
	return nil

}
