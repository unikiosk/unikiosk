package impl

import (
	"context"

	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/grpc/models"
	"github.com/unikiosk/unikiosk/pkg/grpc/service"
	apimodels "github.com/unikiosk/unikiosk/pkg/models"
	"github.com/unikiosk/unikiosk/pkg/queue"
)

//KioskServiceGrpcImpl is a implementation of KioskService Grpc Service.
type KioskServiceGrpcImpl struct {
	log   *zap.Logger
	queue queue.Queue
}

//NewKioskServiceGrpcImpl returns the pointer to the implementation.
func NewKioskServiceGrpcImpl(log *zap.Logger, queue queue.Queue) *KioskServiceGrpcImpl {
	return &KioskServiceGrpcImpl{
		log:   log,
		queue: queue,
	}
}

// StartOrUpdate will start or update running kioks session
func (s *KioskServiceGrpcImpl) StartOrUpdate(ctx context.Context, in *models.KioskState) (*service.StartKioskResponse, error) {
	s.queue.Emit(apimodels.ProtoToKioskState(in))

	return &service.StartKioskResponse{
		State: in,
		Error: nil,
	}, nil
}
