package impl

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/api"
	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/eventer"
	"github.com/unikiosk/unikiosk/pkg/grpc/models"
	"github.com/unikiosk/unikiosk/pkg/grpc/service"
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
func (s *KioskServiceGrpcImpl) StartOrUpdate(ctx context.Context, in *models.KioskRequest) (*service.KioskResponse, error) {
	payload := api.ProtoKioskRequestToModels(in)

	switch s.config.KioskMode {
	// direct mode allows static content and url
	case api.KioskModeDirect:
		// load static file
		if strings.HasPrefix(payload.Content, api.StaticFilePrefix) {
			// update webview only
			callback, err := s.events.Emit(&eventer.EventWrapper{
				Payload: api.Event{
					Type:      api.EventTypeWebViewUpdate,
					KioskMode: api.KioskModeDirect,
					Request:   api.ProtoKioskRequestToModels(in),
				},
			})
			if err != nil {
				return nil, err
			}
			if callback != nil {
				return &service.KioskResponse{
					State: &models.KioskRespose{
						Content:    callback.Payload.Response.Content,
						Title:      callback.Payload.Response.Title,
						SizeW:      callback.Payload.Response.SizeW,
						SizeH:      callback.Payload.Response.SizeH,
						PowerState: models.EnumScreenPowerState(callback.Payload.Response.ScreenPowerState),
					},
					Error: nil,
				}, nil
			}
		} else { //load url
			// update proxy, and proxy will update webview
			// update webview only
			callback, err := s.events.Emit(&eventer.EventWrapper{
				Payload: api.Event{
					Type:      api.EventTypeWebViewUpdate,
					KioskMode: api.KioskModeDirect,
					Request:   api.ProtoKioskRequestToModels(in),
				},
			})
			if err != nil {
				return nil, err
			}
			if callback != nil {
				return &service.KioskResponse{
					State: &models.KioskRespose{
						Content:    callback.Payload.Response.Content,
						Title:      callback.Payload.Response.Title,
						SizeW:      callback.Payload.Response.SizeW,
						SizeH:      callback.Payload.Response.SizeH,
						PowerState: models.EnumScreenPowerState(callback.Payload.Response.ScreenPowerState),
					},
					Error: nil,
				}, nil
			}
		}
	case api.KioskModeProxy:
		// update proxy, and proxy will update webview
		callback, err := s.events.Emit(&eventer.EventWrapper{
			Payload: api.Event{
				Type:      api.EventTypeProxyUpdate,
				KioskMode: api.KioskModeProxy,
				Request:   api.ProtoKioskRequestToModels(in),
			},
		})
		if err != nil {
			return nil, err
		}
		if callback != nil {
			return &service.KioskResponse{
				State: &models.KioskRespose{
					Content:    callback.Payload.Response.Content,
					Title:      callback.Payload.Response.Title,
					SizeW:      callback.Payload.Response.SizeW,
					SizeH:      callback.Payload.Response.SizeH,
					PowerState: models.EnumScreenPowerState(callback.Payload.Response.ScreenPowerState),
				},
				Error: nil,
			}, nil
		}

	default:
		return nil, fmt.Errorf("unsupported update with mode %s", string(s.config.KioskMode))
	}
	return nil, fmt.Errorf("unknown error")
}

// powerOnOrOff will start or update running kiosk session
func (s *KioskServiceGrpcImpl) PowerOnOrOff(ctx context.Context, in *models.KioskRequest) (*service.KioskResponse, error) {
	// update proxy, and proxy will update webview
	callback, err := s.events.Emit(&eventer.EventWrapper{
		Payload: api.Event{
			Type:    api.EventTypePowerAction,
			Request: api.ProtoKioskRequestToModels(in),
		},
	})
	if err != nil {
		return nil, err
	}
	if callback != nil {
		return &service.KioskResponse{
			State: &models.KioskRespose{
				Content:    callback.Payload.Response.Content,
				Title:      callback.Payload.Response.Title,
				SizeW:      callback.Payload.Response.SizeW,
				SizeH:      callback.Payload.Response.SizeH,
				PowerState: models.EnumScreenPowerState(callback.Payload.Response.ScreenPowerState),
			},
			Error: nil,
		}, nil
	}

	return nil, fmt.Errorf("unknown error")
}

// powerOnOrOff will start or update running kiosk session
func (s *KioskServiceGrpcImpl) Screenshot(ctx context.Context, in *models.KioskRequest) (*service.KioskResponse, error) {
	// update proxy, and proxy will update webview
	callback, err := s.events.Emit(&eventer.EventWrapper{
		Payload: api.Event{
			Type:    api.EventTypePowerAction,
			Request: api.ProtoKioskRequestToModels(in),
		},
	})
	if err != nil {
		return nil, err
	}
	if callback != nil {
		for response := range callback.Callback {
			return &service.KioskResponse{
				State: &models.KioskRespose{
					Content:    response.Payload.Response.Content,
					Title:      response.Payload.Response.Title,
					SizeW:      response.Payload.Response.SizeW,
					SizeH:      callback.Payload.Response.SizeH,
					PowerState: models.EnumScreenPowerState(callback.Payload.Response.ScreenPowerState),
					Screenshot: callback.Payload.Response.Screenshot,
				},
				Error: nil,
			}, nil
		}
	}

	return nil, fmt.Errorf("unknown error")
}
