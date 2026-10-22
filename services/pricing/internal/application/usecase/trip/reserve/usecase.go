package reserve

import (
	"context"
	"pricing/internal/domain/entity"
	"pricing/internal/dto/tariff"
	app "pricing/internal/pkg/pricingerror"

	"github.com/gofrs/uuid"
)

type Storage interface {
	Get(context.Context, uuid.UUID) (*entity.Trip, error)
	Create(context.Context, *entity.Trip) error
}
type Profile interface {
	Plan(context.Context, uuid.UUID) (string, error)
}
type Tariffs interface {
	Get(context.Context, string) (tariff.Tariff, error)
}
type UseCase struct {
	storage Storage
	profile Profile
	tariffs Tariffs
}

func New(storage Storage, profile Profile, tariffs Tariffs) *UseCase {
	return &UseCase{
		storage: storage,
		profile: profile,
		tariffs: tariffs,
	}
}
func (s *UseCase) Execute(ctx context.Context, input Input) error {
	if input.RentalVersion != 1 || input.RentalID == uuid.Nil || input.UserID == uuid.Nil || input.ScooterID == uuid.Nil {
		return app.NewInvalidArgumentErr("Некорректное событие бронирования")
	}
	saved, err := s.storage.Get(ctx, input.RentalID)
	if err == nil {
		_, err = saved.CheckEvent(input.UserID, input.ScooterID, input.RentalVersion)
		return err
	}
	if !app.IsCode(err, app.NotFound) {
		return err
	}
	plan, err := s.profile.Plan(ctx, input.UserID)
	if err != nil {
		return err
	}
	price, err := s.tariffs.Get(ctx, plan)
	if err != nil {
		return err
	}
	trip, err := entity.NewTrip(input.RentalID, input.UserID, input.ScooterID,
		price.Plan, price.RateMinor, price.AdvanceMinor, input.RentalVersion, input.ReservedAt)
	if err != nil {
		return err
	}
	return s.storage.Create(ctx, trip)
}
