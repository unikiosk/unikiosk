package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/config"
)

type Proxy interface {
	Run(ctx context.Context) error
}

type proxy struct {
	log    *zap.Logger
	config *config.Config
	server *http.Server

	u *url.URL
}

func New(log *zap.Logger, config *config.Config) (*proxy, error) {
	// TODO: Inject user navigate value here
	u, err := url.Parse(config.DefaultURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse upstream address: %w", err)
	}

	p := proxy{
		log:    log,
		config: config,
		u:      u,
	}

	rp := httputil.NewSingleHostReverseProxy(u)

	r := mux.NewRouter()
	r.Use(handlers.CompressHandler)
	if len(config.ProxyHeaders) > 0 {
		r.Use(p.headers())
	}

	r.PathPrefix("/").Handler(rp)

	p.server = &http.Server{
		Addr:    p.config.ProxyServerURI,
		Handler: r,
	}

	return &p, nil
}

func (p *proxy) headers() func(http.Handler) http.Handler {
	fmt.Println("setting up headers", p.config.ProxyHeaders)
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
	go func() {
		<-ctx.Done()
		p.server.Close()
	}()

	p.log.Info("Proxy running", zap.String("listening", p.config.ProxyServerURI), zap.String("destination", p.u.String()))
	return p.server.ListenAndServe()
}
