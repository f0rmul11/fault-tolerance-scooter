package start

import (
	"context"
	"pricing/internal/domain/entity"
	app "pricing/internal/pkg/pricingerror"

	"github.com/gofrs/uuid"
)

type Storage interface {
	Get(context.Context, uuid.UUID) (*entity.Trip, error)
	Save(context.Context, ...*entity.Trip) error
}
type UseCase struct {
	storage Storage
}

func New(storage Storage) *UseCase {
	return &UseCase{storage: storage}
}
func (s *UseCase) Execute(ctx context.Context, input Input) error {
	if input.RentalVersion < 2 || input.RentalID == uuid.Nil || input.UserID == uuid.Nil || input.ScooterID == uuid.Nil {
		return app.NewInvalidArgumentErr("Некорректное событие поездки")
	}
	trip, err := s.storage.Get(ctx, input.RentalID)
	if err != nil {
		return err
	}
	apply, err := trip.CheckEvent(input.UserID, input.ScooterID, input.RentalVersion)
	if err != nil || !apply {
		return err
	}
	if err := trip.Start(input.RentalVersion, input.StartedAt); err != nil {
		return err
	}
	return s.storage.Save(ctx, trip)
}
