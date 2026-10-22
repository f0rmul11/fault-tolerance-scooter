package get

import (
	"context"

	"github.com/gofrs/uuid"

	"scooter/internal/domain/entity"
	app "scooter/internal/pkg/scootererror"
)

type Storage interface {
	Get(context.Context, uuid.UUID) (*entity.Scooter, error)
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
	return s.storage.Get(ctx, input.ID)
}
