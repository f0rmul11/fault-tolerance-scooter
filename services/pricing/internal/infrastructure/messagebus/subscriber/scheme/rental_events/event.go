package rental_events

const correlationIDHeader = "x-correlation-id"

type BaseEvent struct {
	EventType     string `json:"event_type"`
	CorrelationID string `json:"correlation_id"`
	RentalID      string `json:"rental_id"`
}
