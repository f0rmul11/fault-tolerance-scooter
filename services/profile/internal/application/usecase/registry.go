package usecase

import (
	"profile/internal/application/usecase/profile/get"
	"profile/internal/application/usecase/profile/list"
	"profile/internal/application/usecase/profile/update"
	"profile/internal/infrastructure/storage"
)

type Registry struct {
	GetProfile    *get.UseCase
	ListProfiles  *list.UseCase
	UpdateProfile *update.UseCase
}

func NewRegistry(storage *storage.Registry) *Registry {
	return &Registry{
		GetProfile:    get.New(storage.Profiles),
		ListProfiles:  list.New(storage.Profiles),
		UpdateProfile: update.New(storage.Profiles),
	}
}
