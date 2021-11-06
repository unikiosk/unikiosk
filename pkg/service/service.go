package service

import (
	"context"

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

	manager manager.Kiosk
	web     web.Interface
	grpc    grpc.Server
}

func New(log *zap.Logger, config *config.Config) (*ServiceManager, error) {

	queue, err := queue.New(1)
	if err != nil {
		return nil, err
	}

	manager, err := manager.New(log, config, queue)
	if err != nil {
		return nil, err
	}

	web, err := web.New(log, config)
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

		manager: manager,
		web:     web,
		grpc:    grpc,
	}, nil
}

func (s *ServiceManager) Run(ctx context.Context) error {

	g := &errgroup.Group{}

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
