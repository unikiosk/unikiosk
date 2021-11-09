package disk

import (
	"encoding/json"

	"github.com/peterbourgon/diskv"
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/api"
	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/store"
)

var _ store.Store = &DiskStore{}

type DiskStore struct {
	log    *zap.Logger
	config *config.Config

	store *diskv.Diskv
}

func New(log *zap.Logger, config *config.Config) (*DiskStore, error) {
	d := diskv.New(diskv.Options{
		BasePath:     config.StateDir,
		Transform:    func(s string) []string { return []string{} },
		CacheSizeMax: 1024 * 1024,
	})

	return &DiskStore{
		log:    log,
		config: config,
		store:  d,
	}, nil
}

func (s *DiskStore) Get(key string) (*api.KioskState, error) {
	data, err := s.store.Read(key)
	if err != nil {
		return nil, err
	}
	var r api.KioskState
	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *DiskStore) Persist(key string, in api.KioskState) error {
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return s.store.Write(key, data)
}
