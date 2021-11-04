package impl

import (
	"context"

	"github.com/mjudeikis/unikiosk/pkg/grpc/models"
	"github.com/mjudeikis/unikiosk/pkg/grpc/service"
	apimodels "github.com/mjudeikis/unikiosk/pkg/models"
	"github.com/mjudeikis/unikiosk/pkg/queue"
	"go.uber.org/zap"
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
	s.log.Debug("Received request for start kiosk " + in.Url)

	s.queue.Emit(apimodels.ProtoToKioskState(in))

	return &service.StartKioskResponse{
		State: in,
		Error: nil,
	}, nil
}
