package update

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"profile/internal/domain/entity"
	app "profile/internal/pkg/profileerror"
)

type Storage interface {
	Get(context.Context, uuid.UUID) (*entity.User, error)
}

type Updater interface {
	Update(context.Context, *entity.User) error
}

type UseCase struct {
	storage Storage
	updater Updater
}

func New(storage Storage, updater Updater) *UseCase {
	return &UseCase{
		storage: storage,
		updater: updater,
	}
}

// Execute читает снимок и применяет только переданные поля доменными методами.
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

	// Изменяем копию: отказ валидации или записи не портит загруженный снимок.
	updated := *object
	updated.ExpectVersion(input.ExpectedVersion)
	at := time.Now()

	if value, ok := input.Name.Get(); ok {
		if err := updated.Rename(value, at); err != nil {
			return nil, err
		}
	}
	if value, ok := input.Email.Get(); ok {
		if err := updated.ChangeEmail(value, at); err != nil {
			return nil, err
		}
	}
	if value, ok := input.Plan.Get(); ok {
		if err := updated.ChangeSubscription(value, at); err != nil {
			return nil, err
		}
	}

	// Даже пустой PATCH проходит CAS в БД, сохраняя текущую версию состояния.
	if err := s.updater.Update(ctx, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}
