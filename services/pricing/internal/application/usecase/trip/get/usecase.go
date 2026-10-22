package get

import (
	"context"

	"github.com/gofrs/uuid"

	"pricing/internal/domain/entity"
	app "pricing/internal/pkg/pricingerror"
)

type Storage interface {
	Get(context.Context, uuid.UUID) (*entity.Trip, error)
}

type UseCase struct {
	storage Storage
}

func New(storage Storage) *UseCase {
	return &UseCase{storage: storage}
}

func (s *UseCase) Execute(ctx context.Context, input Input) (*entity.Trip, error) {
	if input.RentalID == uuid.Nil || input.UserID == uuid.Nil {
		return nil, app.NewInvalidArgumentErr("Идентификаторы должны быть UUID")
	}
	object, err := s.storage.Get(ctx, input.RentalID)
	if err != nil {
		return nil, err
	}
	if err := object.CheckOwner(input.UserID); err != nil {
		return nil, err
	}
	return object, nil
}
