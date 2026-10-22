package usecase

import (
	"pricing/internal/application/usecase/tariff/list"
	"pricing/internal/application/usecase/trip/cancel"
	"pricing/internal/application/usecase/trip/complete"
	"pricing/internal/application/usecase/trip/get"
	"pricing/internal/application/usecase/trip/reserve"
	"pricing/internal/application/usecase/trip/start"
	"pricing/internal/infrastructure/gateway"
	"pricing/internal/infrastructure/storage"
)

type Registry struct {
	ReserveTrip  *reserve.UseCase
	StartTrip    *start.UseCase
	CompleteTrip *complete.UseCase
	CancelTrip   *cancel.UseCase
	GetTrip      *get.UseCase
	ListTariffs  *list.UseCase
}

func NewRegistry(storage *storage.Registry, gateways *gateway.Registry) *Registry {
	return &Registry{
		ReserveTrip: reserve.New(
			storage.Trips,
			gateways.Profile,
			storage.Tariffs,
		),
		StartTrip:    start.New(storage.Trips),
		CompleteTrip: complete.New(storage.Trips),
		CancelTrip:   cancel.New(storage.Trips),
		GetTrip:      get.New(storage.Trips),
		ListTariffs:  list.New(storage.Tariffs),
	}
}
