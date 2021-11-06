package service

import (
	"context"
	"sync/atomic"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/grpc"
	"github.com/unikiosk/unikiosk/pkg/manager"
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
}

func New(log *zap.Logger, config *config.Config) (*ServiceManager, error) {

	queue, err := queue.New(1)
	if err != nil {
		return nil, err
	}

	subsystemsReady := atomic.Value{}
	subsystemsReady.Store(false)

	manager, err := manager.New(log, config, queue, &subsystemsReady)
	if err != nil {
		return nil, err
	}

	web, err := web.New(log, config, &subsystemsReady)
	if err != nil {
		return nil, err
	}

	grpc, err := grpc.New(log, config, queue)
	if err != nil {
		return nil, err
	}

	return &ServiceManager{
		log:    log,
		config: config,

		ready: subsystemsReady,

		manager: manager,
		web:     web,
		grpc:    grpc,
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

	// manager must run in the main thread! can't be in separete go routine
	s.manager.Run(ctx)

	return g.Wait()
}
