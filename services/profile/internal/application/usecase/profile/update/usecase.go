package update

import (
	"context"
	"profile/internal/domain/entity"
	app "profile/internal/pkg/profileerror"
	"profile/internal/pkg/transaction"
	"time"

	"github.com/gofrs/uuid"
)

type Storage interface {
	Get(context.Context, uuid.UUID) (*entity.User, error)
	Save(context.Context, ...*entity.User) error
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
	if input.ExpectedVersion <= 0 {
		return nil, app.NewInvalidArgumentErr("expected_version должна быть положительной")
	}
	object, err := s.storage.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if err := object.Update(input.Name, input.Email, input.Plan, input.ExpectedVersion, time.Now()); err != nil {
		return nil, err
	}
	err = transaction.Exec(ctx, func(txCtx context.Context) error {
		return s.storage.Save(txCtx, object)
	})
	if err != nil {
		return nil, err
	}
	return object, nil
}
