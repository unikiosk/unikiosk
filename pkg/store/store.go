package store

import (
	"github.com/unikiosk/unikiosk/pkg/models"
)

type Store interface {
	Get(keys string) (*models.KioskState, error)
	Persist(key string, in models.KioskState) error
}
