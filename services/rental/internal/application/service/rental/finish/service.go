package finish

import (
	"context"
	"log/slog"

	"rental/internal/domain/entity"
	"rental/internal/pkg/event"
	"rental/internal/pkg/transaction"
)

type Storage interface {
	Save(context.Context, ...*entity.Rental) error
}

type Service struct {
	storage Storage
}

func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Finish сохраняет готовую сущность и события из контекста одной транзакцией.
func (s *Service) Finish(ctx context.Context, rental *entity.Rental) error {
	err := transaction.Exec(ctx, func(txCtx context.Context) error {
		if err := s.storage.Save(txCtx, rental); err != nil {
			return err
		}
		return event.Extract(txCtx).Flush(txCtx)
	})
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "Аренда сохранена",
		"operation", "finish_rental",
		"rental_id", rental.ID().String(),
		"request_id", rental.RequestID().String(),
		"status", rental.Status(),
		"version", rental.Version(),
	)
	return nil
}
