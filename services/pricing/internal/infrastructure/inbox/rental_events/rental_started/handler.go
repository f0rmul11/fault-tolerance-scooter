package rental_started

import (
	"context"
	"pricing/internal/application/usecase/trip/start"
	"pricing/internal/infrastructure/inbox/rental_events"

	"github.com/gofrs/uuid"
)

const eventType = "rental.started.v1"

type Handler struct {
	usecase *start.UseCase
}

func NewHandler(usecase *start.UseCase) *Handler {
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
	return h.usecase.Execute(ctx, start.Input{
		RentalID:      uuid.FromStringOrNil(e.RentalID),
		UserID:        uuid.FromStringOrNil(e.Payload.UserID),
		ScooterID:     uuid.FromStringOrNil(e.Payload.ScooterID),
		RentalVersion: e.RentalVersion,
		StartedAt:     e.Payload.StartedAt,
	})
}
