package cancel

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"rental/internal/domain/entity"
	"rental/internal/pkg/event"
	app "rental/internal/pkg/rentalerror"
)

type Storage interface {
	Get(context.Context, uuid.UUID) (*entity.Rental, error)
}

type Transition interface {
	Cancel(context.Context, *entity.Rental) error
}

type UseCase struct {
	storage    Storage
	transition Transition
	flusher    event.Flusher
}

func New(storage Storage, transition Transition, flusher event.Flusher) *UseCase {
	return &UseCase{
		storage:    storage,
		transition: transition,
		flusher:    flusher,
	}
}

// Execute проверяет владение и применяет доменный переход вместе с его событием.
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

	_, ctx = event.WithContext(ctx, s.flusher)
	updated := *rental
	changed, err := updated.Cancel(ctx, input.RequestID, time.Now())
	if err != nil {
		return nil, err
	}
	if !changed {
		return rental, nil
	}

	if err := s.transition.Cancel(ctx, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}
