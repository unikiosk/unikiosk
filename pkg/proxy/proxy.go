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
	log    *zap.Logger
	config *config.Config
	server *http.Server

	events eventer.Eventer

	ready atomic.Value
	lock  sync.RWMutex

	u *url.URL
}

func New(ctx context.Context, log *zap.Logger, config *config.Config, events eventer.Eventer) (*proxy, error) {
	// TODO: Inject user navigate value here
	u, err := url.Parse(config.DefaultWebServerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse upstream address: %w", err)
	}

	p := proxy{
		log:    log,
		config: config,
		events: events,
		u:      u,

		ready: atomic.Value{},
		lock:  sync.RWMutex{},
	}

	h, err := p.getProxyHandler(u)
	if err != nil {
		return nil, err
	}

	p.server = &http.Server{
		Addr:    p.config.ProxyServerAddr,
		Handler: h,
	}

	go func() {
		err := p.runProxyReloader(ctx)
		if err != nil {
			p.log.Debug("failed proxy reload", zap.Error(err))
		}
	}()

	return &p, nil
}

func (p *proxy) getProxyHandler(u *url.URL) (http.Handler, error) {
	rp := httputil.NewSingleHostReverseProxy(u)
	if u.Scheme == "https" {
		// Set a custom DialTLS to access the TLS connection state
		rp.Transport = &http.Transport{DialTLS: dialTLS}
	}

	// Change req.Host so example.com host check is passed
	director := rp.Director
	rp.Director = func(req *http.Request) {
		director(req)
		req.Host = req.URL.Host
	}

	r := mux.NewRouter()
	r.Use(handlers.CompressHandler)
	if len(p.config.ProxyHeaders) > 0 {
		r.Use(p.headers())
	}

	r.PathPrefix("/").Handler(rp)
	return r, nil
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

	cert.VerifyHostname(host)

	return tlsConn, nil
}

func (p *proxy) headers() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for k, v := range p.config.ProxyHeaders {
				w.Header().Set(k, v)
			}

			h.ServeHTTP(w, r)
		})
	}
}

func (p *proxy) Run(ctx context.Context) error {
	p.ready.Store(true)
	p.log.Info("Proxy running", zap.String("listening", p.config.ProxyServerAddr), zap.String("destination", p.u.String()))

	for {
		// lock while we run
		p.lock.Lock()
		p.ready.Store(true)

		err := p.server.ListenAndServe()
		if err != nil {
			p.log.Debug("proxy failed. Restarting", zap.Error(err))
			time.Sleep(time.Second)
		}
		// once failed release lock sp we can hot swap the server
		p.lock.Unlock()
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

func (p *proxy) runProxyReloader(ctx context.Context) error {
	listener, err := p.events.Subscribe(ctx)
	if err != nil {
		return err
	}

	for event := range listener {
		// act only on requests to reload webview
		if event.Type == models.EventTypeProxyUpdate && p.config.KioskMode == models.KioskModeProxy {
			// create new server object and override existing one
			u, err := url.Parse(event.Payload.Content)
			if err != nil {
				return fmt.Errorf("failed to parse upstream address: %w", err)
			}
			p.u = u

			h, err := p.getProxyHandler(u)
			if err != nil {
				return err
			}

			server := &http.Server{
				Addr:    p.server.Addr,
				Handler: h,
			}

			err = p.Stop(ctx)
			if err != nil {
				return err
			}

			p.lock.Lock()
			p.server = server
			p.lock.Unlock()

			// wait until proxy restarts. Restart is done by service package
			// once restared trigger webview reload
			for !p.ready.Load().(bool) {
				p.log.Debug("wait for proxy reload")
				// wait for reload
			}

			// override webview back to proxy as we might be in fire serve mode
			p.events.Emit(&models.Event{
				Type:      models.EventTypeWebViewUpdate,
				KioskMode: models.KioskModeProxy,
			})
		}
	}
	return nil

}
