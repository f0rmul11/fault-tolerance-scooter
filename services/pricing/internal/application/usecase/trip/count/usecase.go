package count

import (
	"context"

	"github.com/gofrs/uuid"

	app "pricing/internal/pkg/pricingerror"
)

type Storage interface {
	CountCompleted(context.Context, uuid.UUID) (int64, error)
}

type UseCase struct {
	storage Storage
}

func New(storage Storage) *UseCase {
	return &UseCase{storage: storage}
}

func (s *UseCase) Execute(ctx context.Context, userID uuid.UUID) (int64, error) {
	if userID == uuid.Nil {
		return 0, app.NewInvalidArgumentErr("Идентификатор должен быть UUID")
	}
	return s.storage.CountCompleted(ctx, userID)
}
