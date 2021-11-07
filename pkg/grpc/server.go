package grpc

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/eventer"
	"github.com/unikiosk/unikiosk/pkg/grpc/impl"
	"github.com/unikiosk/unikiosk/pkg/grpc/service"
	"github.com/unikiosk/unikiosk/pkg/util/recover"
)

type Server interface {
	Run(ctx context.Context) error
}

type GRPCServer struct {
	log    *zap.Logger
	config *config.Config

	listener net.Listener
	server   *grpc.Server
}

func New(log *zap.Logger, config *config.Config, events eventer.Eventer) (*GRPCServer, error) {
	listener, err := net.Listen("tcp", config.GRPCServerAddr)
	if err != nil {
		return nil, err
	}

	// server accepts token authentication:
	// "token" - user or sa token

	opts := []grpc.ServerOption{
		//grpc.UnaryInterceptor(AuthInterceptor),
	}
	server := grpc.NewServer(opts...)

	// register kiosk service
	kioskServiceImpl := impl.NewKioskServiceGrpcImpl(log, events)
	service.RegisterKioskServiceServer(server, kioskServiceImpl)

	return &GRPCServer{
		log:      log,
		config:   config,
		server:   server,
		listener: listener,
	}, nil
}

func (s *GRPCServer) Run(ctx context.Context) error {
	if s.server == nil {
		return fmt.Errorf("GRPC server initiation failed")
	}
	s.log.Info("Starting GRPC Service")
	go func() {
		defer recover.Panic(s.log)

		<-ctx.Done()
		s.server.Stop()

	}()

	s.log.Info("GRPC Server will now listen", zap.String("url", s.config.GRPCServerAddr))
	return s.server.Serve(s.listener)

}
