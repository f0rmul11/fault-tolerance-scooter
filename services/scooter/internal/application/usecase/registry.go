package usecase

import (
	"scooter/internal/application/usecase/scooter/availability"
	"scooter/internal/application/usecase/scooter/get"
	"scooter/internal/application/usecase/scooter/list"
	"scooter/internal/application/usecase/scooter/owner"
	"scooter/internal/infrastructure/storage"
)

type Registry struct {
	GetScooter                *get.UseCase
	ListScooters              *list.UseCase
	UpdateScooterAvailability *availability.UseCase
	UpdateScooterOwner        *owner.UseCase
}

func NewRegistry(storage *storage.Registry) *Registry {
	return &Registry{
		GetScooter:                get.New(storage.Scooters),
		ListScooters:              list.New(storage.Scooters),
		UpdateScooterAvailability: availability.New(storage.Scooters),
		UpdateScooterOwner:        owner.New(storage.Scooters),
	}
}
