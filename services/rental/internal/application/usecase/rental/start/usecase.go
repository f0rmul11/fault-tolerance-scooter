package start

import (
	"context"
	"log/slog"
	"rental/internal/domain/entity"
	"rental/internal/events/rental_started"
	"rental/internal/pkg/event"
	app "rental/internal/pkg/rentalerror"
	"rental/internal/pkg/transaction"
	"time"

	"github.com/gofrs/uuid"
)

type Storage interface {
	Get(context.Context, uuid.UUID) (*entity.Rental, error)
	Save(context.Context, ...*entity.Rental) error
}
type UseCase struct {
	storage Storage
	flusher event.Flusher
}

func New(storage Storage, flusher event.Flusher) *UseCase {
	return &UseCase{
		storage: storage,
		flusher: flusher,
	}
}
func (s *UseCase) Execute(ctx context.Context, input Input) (*entity.Rental, error) {
	if input.RentalID == uuid.Nil || input.RequestID == uuid.Nil {
		return nil, app.NewInvalidArgumentErr("rental_id и request_id должны быть UUID")
	}
	rental, err := s.storage.Get(ctx, input.RentalID)
	if err != nil {
		return nil, err
	}
	if err := rental.CheckOwner(input.UserID); err != nil {
		return nil, err
	}
	changed, err := rental.Start(input.RequestID, time.Now())
	if err != nil {
		return nil, err
	}
	if !changed {
		return rental, nil
	}
	buffer := event.New(s.flusher)
	buffer.Add(rental_started.New(rental))
	err = transaction.Exec(ctx, func(txCtx context.Context) error {
		if err := s.storage.Save(txCtx, rental); err != nil {
			return err
		}
		return buffer.Flush(txCtx)
	})
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "Аренда сохранена",
		"operation", "start_rental",
		"rental_id", rental.ID().String(),
		"request_id", rental.RequestID().String(),
		"status", rental.Status(),
		"version", rental.Version(),
	)
	return rental, nil
}
