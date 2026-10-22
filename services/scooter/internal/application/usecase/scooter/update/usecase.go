package update

import (
	"context"
	"scooter/internal/domain/entity"
	app "scooter/internal/pkg/scootererror"
	"time"

	"github.com/gofrs/uuid"
)

type Storage interface {
	Get(context.Context, uuid.UUID) (*entity.Scooter, error)
	Save(context.Context, ...*entity.Scooter) error
}
type UseCase struct {
	storage Storage
}

func New(storage Storage) *UseCase {
	return &UseCase{storage: storage}
}
func (s *UseCase) Execute(ctx context.Context, input Input) (*entity.Scooter, error) {
	if input.ID == uuid.Nil {
		return nil, app.NewInvalidArgumentErr("Идентификатор должен быть UUID")
	}
	if input.ExpectedVersion <= 0 {
		return nil, app.NewInvalidArgumentErr("expected_version должна быть положительной")
	}
	object, err := s.storage.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if err := object.Update(input.Number, input.Brand, input.ChargePercent, input.RiddenSeconds, input.ExpectedVersion, time.Now()); err != nil {
		return nil, err
	}
	err = s.storage.Save(ctx, object)
	if err != nil {
		return nil, err
	}
	return object, nil
}
