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

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/eventer"
	"github.com/unikiosk/unikiosk/pkg/models"
)

type Proxy interface {
	Run(ctx context.Context) error
}

type proxy struct {
	config *config.Config
	server *http.Server

	events eventer.Eventer

	ready atomic.Value

	targetURLMu *sync.RWMutex
	targetURL   *url.URL

	log *zap.Logger
}

func New(ctx context.Context, log *zap.Logger, config *config.Config, events eventer.Eventer) (*proxy, error) {
	// TODO: Inject user navigate value here
	defaultURL, err := url.Parse(config.DefaultWebServerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse upstream address: %w", err)
	}

	p := proxy{
		log:       log,
		config:    config,
		events:    events,
		targetURL: defaultURL,

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

	fmt.Println("target: ", req.URL.String())
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
	listener, err := p.events.Subscribe(ctx)
	if err != nil {
		return err
	}

	for event := range listener {
		// act only on requests to reload webview
		if p.config.KioskMode == models.KioskModeProxy &&
			event.Type == models.EventTypeProxyUpdate {
			// create new server object and override existing one
			u, err := url.Parse(event.Payload.Content)
			if err != nil {
				p.log.Error("failed to parse target URL",
					zap.String("url", event.Payload.Content),
					zap.Error(err),
				)
				continue
			}
			p.targetURLMu.Lock()
			p.targetURL = u
			p.targetURLMu.Unlock()

			// override webview back to proxy as we might be in file serve mode
			p.events.Emit(&models.Event{
				Type:      models.EventTypeWebViewUpdate,
				KioskMode: models.KioskModeProxy,
			})
		}
	}
	return nil

}
