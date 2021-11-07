package service

import (
	"context"
	"sync/atomic"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/grpc"
	"github.com/unikiosk/unikiosk/pkg/manager"
	"github.com/unikiosk/unikiosk/pkg/proxy"
	"github.com/unikiosk/unikiosk/pkg/queue"
	"github.com/unikiosk/unikiosk/pkg/web"
)

type Service interface {
	Run(ctx context.Context) error
}

type ServiceManager struct {
	log    *zap.Logger
	config *config.Config

	ready atomic.Value

	manager manager.Kiosk
	web     web.Interface
	grpc    grpc.Server
	proxy   proxy.Proxy
}

func New(log *zap.Logger, config *config.Config) (*ServiceManager, error) {

	queue, err := queue.New(1)
	if err != nil {
		return nil, err
	}

	manager, err := manager.New(log.Named("webview"), config, queue)
	if err != nil {
		return nil, err
	}

	web, err := web.New(log.Named("webserver"), config)
	if err != nil {
		return nil, err
	}

	grpc, err := grpc.New(log.Named("grpc-api"), config, queue)
	if err != nil {
		return nil, err
	}

	proxy, err := proxy.New(log.Named("proxy"), config)
	if err != nil {
		return nil, err
	}

	return &ServiceManager{
		log:    log,
		config: config,

		manager: manager,
		web:     web,
		grpc:    grpc,
		proxy:   proxy,
	}, nil
}

func (s *ServiceManager) Run(ctx context.Context) error {

	g := &errgroup.Group{}

	// due to nature of go routines execution we use atomic.Ready to start webview only when all other subsystems are ready
	g.Go(func() error {
		return s.grpc.Run(ctx)
	})

	g.Go(func() error {
		return s.web.Run(ctx)
	})
	g.Go(func() error {
		return s.proxy.Run(ctx)
	})

	// manager must run in the main thread! can't be in separete go routine
	s.manager.Run(ctx)

	return g.Wait()
}
