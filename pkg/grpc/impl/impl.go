package impl

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/eventer"
	"github.com/unikiosk/unikiosk/pkg/grpc/models"
	"github.com/unikiosk/unikiosk/pkg/grpc/service"
	apimodels "github.com/unikiosk/unikiosk/pkg/models"
)

//KioskServiceGrpcImpl is a implementation of KioskService GRPC Service.
type KioskServiceGrpcImpl struct {
	log    *zap.Logger
	config *config.Config
	events eventer.Eventer
}

//NewKioskServiceGrpcImpl returns the pointer to the implementation.
func NewKioskServiceGrpcImpl(log *zap.Logger, config *config.Config, events eventer.Eventer) *KioskServiceGrpcImpl {
	return &KioskServiceGrpcImpl{
		log:    log,
		config: config,
		events: events,
	}
}

// StartOrUpdate will start or update running kiosk session
func (s *KioskServiceGrpcImpl) StartOrUpdate(ctx context.Context, in *models.KioskState) (*service.StartKioskResponse, error) {
	payload := apimodels.ProtoToKioskState(in)

	switch s.config.KioskMode {
	// direct mode allows static content and url
	case apimodels.KioskModeDirect:
		// load static file
		if strings.HasPrefix(payload.Content, apimodels.StaticFilePrefix) {
			// update webview only
			s.events.Emit(&apimodels.Event{
				Type:      apimodels.EventTypeWebViewUpdate,
				KioskMode: apimodels.KioskModeDirect,
				Payload:   apimodels.ProtoToKioskState(in),
			})
			// load url
		} else {
			// update proxy, and proxy will update webview
			s.events.Emit(&apimodels.Event{
				Type:      apimodels.EventTypeWebViewUpdate,
				KioskMode: apimodels.KioskModeDirect,
				Payload:   apimodels.ProtoToKioskState(in),
			})
		}
	case apimodels.KioskModeProxy:
		// update proxy, and proxy will update webview
		s.events.Emit(&apimodels.Event{
			Type:      apimodels.EventTypeProxyUpdate,
			KioskMode: apimodels.KioskModeProxy,
			Payload:   apimodels.ProtoToKioskState(in),
		})
	default:
		return nil, fmt.Errorf("unsupported update with mode %s", string(s.config.KioskMode))
	}

	return &service.StartKioskResponse{
		State: in,
		Error: nil,
	}, nil
}
