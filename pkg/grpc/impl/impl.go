package impl

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/eventer"
	"github.com/unikiosk/unikiosk/pkg/grpc/models"
	"github.com/unikiosk/unikiosk/pkg/grpc/service"
	apimodels "github.com/unikiosk/unikiosk/pkg/models"
)

//KioskServiceGrpcImpl is a implementation of KioskService GRPC Service.
type KioskServiceGrpcImpl struct {
	log    *zap.Logger
	events eventer.Eventer
}

//NewKioskServiceGrpcImpl returns the pointer to the implementation.
func NewKioskServiceGrpcImpl(log *zap.Logger, events eventer.Eventer) *KioskServiceGrpcImpl {
	return &KioskServiceGrpcImpl{
		log:    log,
		events: events,
	}
}

// StartOrUpdate will start or update running kiosk session
func (s *KioskServiceGrpcImpl) StartOrUpdate(ctx context.Context, in *models.KioskState) (*service.StartKioskResponse, error) {
	payload := apimodels.ProtoToKioskState(in)

	// file load request
	if strings.HasPrefix(payload.Content, apimodels.StaticFilePrefix) {
		// update webview only
		s.events.Emit(&apimodels.Event{
			Type:    apimodels.EventTypeWebViewUpdate,
			Payload: apimodels.ProtoToKioskState(in),
		})
	} else {
		// update proxy, and proxy will update webview
		s.events.Emit(&apimodels.Event{
			Type:    apimodels.EventTypeProxyUpdate,
			Payload: apimodels.ProtoToKioskState(in),
		})
	}

	return &service.StartKioskResponse{
		State: in,
		Error: nil,
	}, nil
}
