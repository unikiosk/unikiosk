package store

import (
	"github.com/mjudeikis/unikiosk/pkg/models"
)

type Store interface {
	Get() (*models.KioskState, error)
	Persist(in models.KioskState) error
}
