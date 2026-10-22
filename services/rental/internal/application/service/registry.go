package service

import (
	"rental/internal/application/service/rental/cancel"
	"rental/internal/application/service/rental/finish"
	"rental/internal/application/service/rental/reserve"
	"rental/internal/application/service/rental/start"
	"rental/internal/infrastructure/storage"
)

type Registry struct {
	ReserveRental *reserve.Service
	StartRental   *start.Service
	FinishRental  *finish.Service
	CancelRental  *cancel.Service
}

func NewRegistry(storage *storage.Registry) *Registry {
	return &Registry{
		ReserveRental: reserve.New(storage.Rentals),
		StartRental:   start.New(storage.Rentals),
		FinishRental:  finish.New(storage.Rentals),
		CancelRental:  cancel.New(storage.Rentals),
	}
}
