package rental_events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

type Envelope struct {
	CorrelationID string    `json:"correlation_id"`
	EventType     string    `json:"event_type"`
	RentalID      string    `json:"rental_id"`
	RentalVersion int       `json:"rental_version"`
	OccurredAt    time.Time `json:"occurred_at"`
	RequestID     string    `json:"request_id"`
	Payload       Payload   `json:"payload"`
}
type Payload struct {
	UserID        string    `json:"user_id"`
	ScooterID     string    `json:"scooter_id"`
	ScooterNumber string    `json:"scooter_number"`
	ScooterBrand  string    `json:"scooter_brand"`
	Status        string    `json:"status"`
	ReservedAt    time.Time `json:"reserved_at"`
	StartedAt     time.Time `json:"started_at"`
	FinishedAt    time.Time `json:"finished_at"`
}

func Deserialize(body []byte, eventType string) (Envelope, error) {
	var e Envelope
	if err := json.Unmarshal(body, &e); err != nil {
		return e, err
	}
	for _, id := range []string{
		e.CorrelationID,
		e.RentalID,
		e.RequestID,
		e.Payload.UserID,
		e.Payload.ScooterID,
	} {
		if uuid.FromStringOrNil(id) == uuid.Nil {
			return e, fmt.Errorf("событие содержит некорректный UUID")
		}
	}
	if e.EventType != eventType || e.OccurredAt.IsZero() || e.Payload.ReservedAt.IsZero() {
		return e, fmt.Errorf("неполное событие аренды")
	}
	switch eventType {
	case "rental.reserved.v1":
		if e.RentalVersion != 1 || e.Payload.Status != "RESERVED" || !e.Payload.StartedAt.IsZero() || !e.Payload.FinishedAt.IsZero() {
			return e, fmt.Errorf("некорректное событие бронирования")
		}
	case "rental.started.v1":
		if e.RentalVersion != 2 || e.Payload.Status != "ACTIVE" || e.Payload.StartedAt.IsZero() || !e.Payload.FinishedAt.IsZero() {
			return e, fmt.Errorf("некорректное событие начала поездки")
		}
	case "rental.completed.v1":
		if e.RentalVersion != 3 || e.Payload.Status != "COMPLETED" || e.Payload.StartedAt.IsZero() || e.Payload.FinishedAt.IsZero() {
			return e, fmt.Errorf("некорректное событие завершения поездки")
		}
	case "rental.canceled.v1":
		if e.RentalVersion != 2 || e.Payload.Status != "CANCELED" || !e.Payload.StartedAt.IsZero() || !e.Payload.FinishedAt.IsZero() {
			return e, fmt.Errorf("некорректное событие отмены")
		}
	default:
		return e, fmt.Errorf("неизвестный тип события аренды")
	}
	return e, nil
}
