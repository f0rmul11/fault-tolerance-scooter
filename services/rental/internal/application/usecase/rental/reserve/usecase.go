package reserve

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"rental/internal/domain/entity"
	scooterdto "rental/internal/dto/scooter"
	"rental/internal/pkg/event"
	app "rental/internal/pkg/rentalerror"
)

type Storage interface {
	FindKey(context.Context, uuid.UUID, uuid.UUID) (*entity.Rental, error)
}

type Reservation interface {
	Reserve(context.Context, *entity.Rental) error
}

type Scooter interface {
	Get(context.Context, uuid.UUID) (scooterdto.Scooter, error)
}

type UseCase struct {
	storage     Storage
	reservation Reservation
	scooter     Scooter
	flusher     event.Flusher
}

func New(storage Storage, reservation Reservation, scooter Scooter, flusher event.Flusher) *UseCase {
	return &UseCase{
		storage:     storage,
		reservation: reservation,
		scooter:     scooter,
		flusher:     flusher,
	}
}

// Execute возвращает сохранённый повтор до любых обращений к каталогу.
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

	// Каталог нужен только для новой брони, после проверки идемпотентности.
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

	// Домен проверяет условия бронирования и формирует событие в контексте операции.
	_, ctx = event.WithContext(ctx, s.flusher)
	rental, err := entity.NewRental(
		ctx,
		id,
		input.UserID,
		input.ScooterID,
		input.IdempotencyKey,
		input.RequestID,
		scooter.Number,
		scooter.Brand,
		scooter.IsActive,
		time.Now(),
	)
	if err != nil {
		return nil, err
	}
	if err := s.reservation.Reserve(ctx, rental); err != nil {
		return nil, err
	}
	return rental, nil
}
