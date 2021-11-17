package impl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

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
	// update webview only
	callback, err := s.events.Emit(&eventer.EventWrapper{
		Payload: api.Event{
			Type:    api.EventTypeProxyUpdate,
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
func (s *KioskServiceGrpcImpl) PowerOnOrOff(ctx context.Context, in *models.KioskRequest) (*service.KioskResponse, error) {
	// update proxy, and proxy will update webview
	callback, err := s.events.Emit(&eventer.EventWrapper{
		Payload: api.Event{
			Type:    api.EventTypeAction,
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
				Screenshot: callback.Payload.Response.Screenshot,
			},
			Error: nil,
		}, nil
	}

	return nil, fmt.Errorf("unknown error")
}

// Screenshot stream screenshot image
func (s *KioskServiceGrpcImpl) Screenshot(in *models.KioskRequest, srv service.KioskService_ScreenshotServer) error {
	s.log.Debug("screenshot taken")
	// update proxy, and proxy will update webview
	callback, err := s.events.Emit(&eventer.EventWrapper{
		Payload: api.Event{
			Type:    api.EventTypeAction,
			Request: api.ProtoKioskRequestToModels(in),
		},
	})
	if err != nil {
		return err
	}

	screenshot := callback.Payload.Response.Screenshot
	screenshotBuf := bytes.NewBuffer(screenshot)

	bufferSize := 64 * 1024 //64KiB
	buff := make([]byte, bufferSize)
	for {
		bytesRead, err := screenshotBuf.Read(buff)
		if err != nil {
			if err != io.EOF {
				err := srv.Send(&service.KioskScreenshootResponse{
					Error: &models.Error{
						Code:    "0",
						Message: err.Error(),
					},
				})
				return err
			}
			break
		}
		err = srv.Send(&service.KioskScreenshootResponse{
			Screenshot: &models.KioskScreenshotRespose{
				Screenshot: buff[:bytesRead],
			},
		})
		if err != nil {
			log.Println("error while sending chunk:", err)
			err = srv.Send(&service.KioskScreenshootResponse{
				Error: &models.Error{
					Code:    "0",
					Message: err.Error(),
				},
			})
			return err
		}
	}

	return nil
}
