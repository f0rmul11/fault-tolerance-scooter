package reserve

import (
	"context"
	"log/slog"
	"rental/internal/domain/entity"
	scooterdto "rental/internal/dto/scooter"
	"rental/internal/events/rental_reserved"
	"rental/internal/pkg/event"
	app "rental/internal/pkg/rentalerror"
	"rental/internal/pkg/transaction"
	"time"

	"github.com/gofrs/uuid"
)

type Storage interface {
	FindKey(context.Context, uuid.UUID, uuid.UUID) (*entity.Rental, error)
	Create(context.Context, *entity.Rental) error
}
type Scooter interface {
	Get(context.Context, uuid.UUID) (scooterdto.Scooter, error)
}
type UseCase struct {
	storage Storage
	scooter Scooter
	flusher event.Flusher
}

func New(storage Storage, scooter Scooter, flusher event.Flusher) *UseCase {
	return &UseCase{
		storage: storage,
		scooter: scooter,
		flusher: flusher,
	}
}
func (s *UseCase) Execute(ctx context.Context, input Input) (*entity.Rental, error) {
	if input.UserID == uuid.Nil || input.ScooterID == uuid.Nil || input.IdempotencyKey == uuid.Nil || input.RequestID == uuid.Nil {
		return nil, app.NewInvalidArgumentErr("Идентификаторы запроса должны быть UUID")
	}
	saved, err := s.storage.FindKey(ctx, input.UserID, input.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if saved != nil {
		if err := saved.MatchReservation(input.ScooterID); err != nil {
			return nil, err
		}
		return saved, nil
	}
	scooter, err := s.scooter.Get(ctx, input.ScooterID)
	if err != nil {
		return nil, err
	}
	if scooter.ID != input.ScooterID {
		return nil, app.NewInternalError("Каталог вернул другой самокат", nil)
	}
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	rental, err := entity.NewRental(id, input.UserID, input.ScooterID, input.IdempotencyKey, input.RequestID,
		scooter.Number, scooter.Brand, scooter.ChargePercent, time.Now())
	if err != nil {
		return nil, err
	}
	buffer := event.New(s.flusher)
	buffer.Add(rental_reserved.New(rental))
	err = transaction.Exec(ctx, func(txCtx context.Context) error {
		if err := s.storage.Create(txCtx, rental); err != nil {
			return err
		}
		return buffer.Flush(txCtx)
	})
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "Аренда сохранена",
		"operation", "reserve_rental",
		"rental_id", rental.ID().String(),
		"request_id", rental.RequestID().String(),
		"status", rental.Status(),
		"version", rental.Version(),
	)
	return rental, nil
}
