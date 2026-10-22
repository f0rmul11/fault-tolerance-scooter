package service

import (
	"profile/internal/application/service/profile/update"
	"profile/internal/infrastructure/storage"
)

type Registry struct {
	UpdateProfile *update.Service
}

func NewRegistry(storage *storage.Registry) *Registry {
	return &Registry{
		UpdateProfile: update.New(storage.Profiles),
	}
}
