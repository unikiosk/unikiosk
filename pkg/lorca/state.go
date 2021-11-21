package lorca

import (
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/api"
)

func (k *kiosk) getCurrentState() api.KioskState {
	state, err := k.store.Get(stateKey)
	if err != nil || state == nil {
		k.log.Info("no state found - start fresh")
		s := api.KioskState{
			Content: k.config.DefaultWebServerURL,
			//SizeW:   int64(C.display_width()),
			//SizeH:   int64(C.display_height()),
			Title: "UniKiosk",
		}
		err = k.store.Persist(stateKey, s)
		if err != nil {
			k.log.Warn("failed to persist store, will not recover after restart", zap.Error(err))
		}
		return s
	} else {
		return *state
	}
}
