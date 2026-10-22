package rental_canceled

import (
	"context"
	"pricing/internal/application/usecase/trip/cancel"
	"pricing/internal/infrastructure/inbox/rental_events"

	"github.com/gofrs/uuid"
)

const eventType = "rental.canceled.v1"

type Handler struct {
	usecase *cancel.UseCase
}

func NewHandler(usecase *cancel.UseCase) *Handler {
	return &Handler{usecase: usecase}
}
func (h *Handler) EventType() string {
	return eventType
}
func (h *Handler) HandleEvent(ctx context.Context, body []byte) error {
	e, err := rental_events.Deserialize(body, eventType)
	if err != nil {
		return err
	}
	return h.usecase.Execute(ctx, cancel.Input{
		RentalID:      uuid.FromStringOrNil(e.RentalID),
		UserID:        uuid.FromStringOrNil(e.Payload.UserID),
		ScooterID:     uuid.FromStringOrNil(e.Payload.ScooterID),
		RentalVersion: e.RentalVersion,
		ReservedAt:    e.Payload.ReservedAt,
		StartedAt:     e.Payload.StartedAt,
		FinishedAt:    e.Payload.FinishedAt,
	})
}
