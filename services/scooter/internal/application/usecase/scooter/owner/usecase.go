package owner

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"scooter/internal/domain/entity"
	app "scooter/internal/pkg/scootererror"
)

type Storage interface {
	Get(context.Context, uuid.UUID) (*entity.Scooter, error)
	Save(context.Context, ...*entity.Scooter) error
}

type UseCase struct {
	storage Storage
}

func New(storage Storage) *UseCase {
	return &UseCase{
		storage: storage,
	}
}

// Execute изменяет одно поле; ожидаемую версию проверяет SQL UPDATE.
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
	object.ExpectVersion(input.ExpectedVersion)
	if err := object.ChangeOwner(input.Owner, time.Now()); err != nil {
		return nil, err
	}

	// Повтор прежнего значения тоже проверяет версию в БД, но не увеличивает её.
	if err := s.storage.Save(ctx, object); err != nil {
		return nil, err
	}
	return object, nil
}
