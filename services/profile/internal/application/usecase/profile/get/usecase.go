package get

import (
	"context"
	"profile/internal/domain/entity"
	app "profile/internal/pkg/profileerror"

	"github.com/gofrs/uuid"
)

type Storage interface {
	Get(context.Context, uuid.UUID) (*entity.User, error)
}
type UseCase struct {
	storage Storage
}

func New(storage Storage) *UseCase {
	return &UseCase{storage: storage}
}
func (s *UseCase) Execute(ctx context.Context, input Input) (*entity.User, error) {
	if input.ID == uuid.Nil {
		return nil, app.NewInvalidArgumentErr("Идентификатор должен быть UUID")
	}
	return s.storage.Get(ctx, input.ID)
}
