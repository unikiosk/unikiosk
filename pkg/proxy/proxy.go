package proxy

import (
	"context"

	"github.com/elazarl/goproxy"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/unikiosk/unikiosk/pkg/config"
)

type Proxy interface {
	Run(ctx context.Context) error
}

type proxy struct {
	config     *config.Config
	proxyHTTP  *goproxy.ProxyHttpServer
	proxyHTTPS *goproxy.ProxyHttpServer

	log *zap.Logger
}

func New(ctx context.Context, log *zap.Logger, config *config.Config) (*proxy, error) {

	p := proxy{
		log:    log,
		config: config,
	}

	proxyHTTP := goproxy.NewProxyHttpServer()
	proxyHTTP.Verbose = true
	p.proxyHTTP = proxyHTTP

	proxyHTTPS := goproxy.NewProxyHttpServer()
	proxyHTTPS.Verbose = true
	p.proxyHTTPS = proxyHTTPS

	err := p.setCertificate()
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (p *proxy) Run(ctx context.Context) error {
	p.log.Debug("Proxy server starting up", zap.String("http", p.config.ProxyHTTPServerAddr))

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		return p.runHTTP(ctx)
	})

	g.Go(func() error {
		return p.runHTTPS(ctx)
	})

	return g.Wait()
}

func (p *proxy) Stop(ctx context.Context) error {
	p.log.Info("stopping proxy")
	// TODO: implement
	return nil
}
