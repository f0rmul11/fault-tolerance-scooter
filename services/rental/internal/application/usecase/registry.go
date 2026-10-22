package usecase

import (
	"rental/internal/application/service"
	"rental/internal/application/usecase/rental/cancel"
	"rental/internal/application/usecase/rental/finish"
	"rental/internal/application/usecase/rental/get"
	"rental/internal/application/usecase/rental/list"
	"rental/internal/application/usecase/rental/reserve"
	"rental/internal/application/usecase/rental/start"
	"rental/internal/infrastructure/gateway"
	"rental/internal/infrastructure/storage"
	"rental/internal/pkg/event"
)

type Registry struct {
	ReserveRental *reserve.UseCase
	StartRental   *start.UseCase
	FinishRental  *finish.UseCase
	CancelRental  *cancel.UseCase
	GetRental     *get.UseCase
	ListRentals   *list.UseCase
}

func NewRegistry(storage *storage.Registry, services *service.Registry, gateways *gateway.Registry, flusher event.Flusher) *Registry {
	return &Registry{
		ReserveRental: reserve.New(
			storage.Rentals,
			services.ReserveRental,
			gateways.Scooter,
			flusher,
		),
		StartRental:  start.New(storage.Rentals, services.StartRental, flusher),
		FinishRental: finish.New(storage.Rentals, services.FinishRental, flusher),
		CancelRental: cancel.New(storage.Rentals, services.CancelRental, flusher),
		GetRental:    get.New(storage.Rentals),
		ListRentals:  list.New(storage.Rentals),
	}
}
