package update

import (
	"context"

	"profile/internal/domain/entity"
	"profile/internal/pkg/transaction"
)

type Storage interface {
	Save(context.Context, ...*entity.User) error
	SaveSubscriptions(context.Context, ...*entity.User) error
}

type Service struct {
	storage Storage
}

func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Update сохраняет готовый профиль и подписку одной транзакцией.
func (s *Service) Update(ctx context.Context, object *entity.User) error {
	return transaction.Exec(ctx, func(txCtx context.Context) error {
		if err := s.storage.Save(txCtx, object); err != nil {
			return err
		}
		return s.storage.SaveSubscriptions(txCtx, object)
	})
}
