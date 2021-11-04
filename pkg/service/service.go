package service

import (
	"context"

	"github.com/mjudeikis/unikiosk/pkg/config"
	"github.com/mjudeikis/unikiosk/pkg/grpc"
	"github.com/mjudeikis/unikiosk/pkg/manager"
	"github.com/mjudeikis/unikiosk/pkg/queue"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	Run(ctx context.Context) error
}

type ServiceManager struct {
	log    *zap.Logger
	config *config.Config

	manager manager.Kiosk
	grpc    grpc.Server
}

func New(log *zap.Logger, config *config.Config) (*ServiceManager, error) {

	queue, err := queue.New(1)
	if err != nil {
		return nil, err
	}

	manager := manager.New(log, config, queue)
	grpc, err := grpc.New(log, config, queue)
	if err != nil {
		return nil, err
	}

	return &ServiceManager{
		log:    log,
		config: config,

		manager: manager,
		grpc:    grpc,
	}, nil
}

func (s *ServiceManager) Run(ctx context.Context) error {

	g := &errgroup.Group{}

	g.Go(func() error {
		return s.grpc.Run(ctx)
	})

	// manager must run in the main thread! can't be in separete go routine
	s.manager.Run(ctx)

	return g.Wait()
}
