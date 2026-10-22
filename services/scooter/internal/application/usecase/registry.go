package usecase

import (
	"scooter/internal/application/usecase/scooter/get"
	"scooter/internal/application/usecase/scooter/list"
	"scooter/internal/application/usecase/scooter/update"
	"scooter/internal/infrastructure/storage"
)

type Registry struct {
	GetScooter    *get.UseCase
	ListScooters  *list.UseCase
	UpdateScooter *update.UseCase
}

func NewRegistry(storage *storage.Registry) *Registry {
	return &Registry{
		GetScooter:    get.New(storage.Scooters),
		ListScooters:  list.New(storage.Scooters),
		UpdateScooter: update.New(storage.Scooters),
	}
}
