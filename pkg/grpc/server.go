package grpc

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/mjudeikis/unikiosk/pkg/config"
	"github.com/mjudeikis/unikiosk/pkg/grpc/impl"
	"github.com/mjudeikis/unikiosk/pkg/grpc/service"
	"github.com/mjudeikis/unikiosk/pkg/queue"
	"github.com/mjudeikis/unikiosk/pkg/util/recover"
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

func New(log *zap.Logger, config *config.Config, queue queue.Queue) (*GRPCServer, error) {
	listener, err := net.Listen("tcp", config.GRPCServerURI)
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
	kioskServiceImpl := impl.NewKioskServiceGrpcImpl(log, queue)
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

	s.log.Info("GRPC Server will now listen", zap.String("url", s.config.GRPCServerURI))
	return s.server.Serve(s.listener)

}

//func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
//	meta, ok := metadata.FromIncomingContext(ctx)
//	if !ok {
//		return nil, status.Errorf(codes.Unauthenticated, "missing context metadata")
//	}
//	if len(meta["token"]) != 1 {
//		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
//	}
//	if meta["token"][0] != "valid-token" {
//		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
//	}
//
//	return handler(ctx, req)
//}
