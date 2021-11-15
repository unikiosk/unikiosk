package service

import (
	"context"
	"sync/atomic"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/eventer"
	"github.com/unikiosk/unikiosk/pkg/grpc"
	"github.com/unikiosk/unikiosk/pkg/proxy"
	"github.com/unikiosk/unikiosk/pkg/store/disk"
	"github.com/unikiosk/unikiosk/pkg/util/recover"
	"github.com/unikiosk/unikiosk/pkg/web"
	"github.com/unikiosk/unikiosk/pkg/webview"
)

type Service interface {
	Run(ctx context.Context) error
}

type ServiceManager struct {
	log    *zap.Logger
	config *config.Config

	ready atomic.Value

	webview webview.Kiosk
	web     web.Interface
	grpc    grpc.Server
	proxy   proxy.Proxy
}

func New(ctx context.Context, log *zap.Logger, config *config.Config) (*ServiceManager, error) {

	events := eventer.New(ctx, log)

	store, err := disk.New(log, config)
	if err != nil {
		return nil, err
	}

	webview, err := webview.New(log.Named("webview"), config, events, store)
	if err != nil {
		return nil, err
	}

	web, err := web.New(log.Named("webserver"), config)
	if err != nil {
		return nil, err
	}

	grpc, err := grpc.New(log.Named("grpc-api"), config, events)
	if err != nil {
		return nil, err
	}

	proxy, err := proxy.New(ctx, log.Named("proxy"), config, events, store)
	if err != nil {
		return nil, err
	}

	return &ServiceManager{
		log:    log,
		config: config,

		webview: webview,
		web:     web,
		grpc:    grpc,
		proxy:   proxy,
	}, nil
}

func (s *ServiceManager) Run(ctx context.Context) error {

	g := &errgroup.Group{}

	// due to nature of go routines execution we use atomic.Ready to start webview only when all other subsystems are ready
	g.Go(func() error {
		defer recover.Panic(s.log)
		return s.grpc.Run(ctx)
	})

	g.Go(func() error {
		defer recover.Panic(s.log)
		return s.web.Run(ctx)
	})
	g.Go(func() error {
		defer recover.Panic(s.log)
		return s.proxy.Run(ctx)
	})

	// webview must run in the main thread! can't be in separete go routine
	s.webview.Run(ctx)
	defer s.webview.Close()

	return g.Wait()
}
