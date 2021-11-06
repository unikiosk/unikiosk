package store

import (
	"github.com/unikiosk/unikiosk/pkg/models"
)

type Store interface {
	Get() (*models.KioskState, error)
	Persist(in models.KioskState) error
}
