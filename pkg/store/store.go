package store

import (
	"github.com/unikiosk/unikiosk/pkg/api"
)

type Store interface {
	Get(keys string) (*api.KioskState, error)
	Persist(key string, in api.KioskState) error
}
