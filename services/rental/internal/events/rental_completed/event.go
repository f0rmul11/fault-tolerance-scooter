package rental_completed

import (
	"context"
	"encoding/json"
	"rental/config"
	"rental/internal/domain/entity"
	"rental/internal/events"
	"rental/internal/pkg/event"
	"rental/internal/pkg/pipe"
	app "rental/internal/pkg/rentalerror"
	"time"

	"github.com/gofrs/uuid"
)

const Type = "rental.completed.v1"

func New(rental *entity.Rental) pipe.Func[event.Events] {
	return func(ctx context.Context, batch event.Events) (event.Events, error) {
		if rental.Status() != entity.Completed {
			return nil, app.NewInternalError("Состояние аренды не соответствует событию", nil)
		}
		correlationID, err := uuid.NewV4()
		if err != nil {
			return nil, err
		}
		payload := Payload{
			UserID:        rental.UserID().String(),
			ScooterID:     rental.ScooterID().String(),
			ScooterNumber: rental.ScooterNumber(),
			ScooterBrand:  rental.ScooterBrand(),
			Status:        string(rental.Status()),
			ReservedAt:    rental.ReservedAt().Format(time.RFC3339Nano),
		}
		if !rental.StartedAt().IsZero() {
			payload.StartedAt = rental.StartedAt().Format(time.RFC3339Nano)
		}
		if !rental.FinishedAt().IsZero() {
			payload.FinishedAt = rental.FinishedAt().Format(time.RFC3339Nano)
		}
		baseEvent := events.Base[Payload]{
			CorrelationID: correlationID.String(),
			EventType:     Type,
			RentalID:      rental.ID().String(),
			RentalVersion: rental.Version(),
			OccurredAt:    rental.UpdatedAt().Format(time.RFC3339Nano),
			RequestID:     rental.RequestID().String(),
			Payload:       payload,
		}
		body, err := json.Marshal(baseEvent)
		if err != nil {
			return nil, err
		}
		headers, err := json.Marshal(map[string]string{
			"x-correlation-id": correlationID.String(),
			"x-app-name":       "rental",
			"x-event-type":     Type,
		})
		if err != nil {
			return nil, err
		}
		return append(batch, event.Event{
			EntityID: rental.ID().String(),
			Key:      event.Raw(rental.ID().String()),
			Body:     body,
			Headers:  headers,
			Schema:   config.RentalEventsTopic,
		}), nil
	}
}
