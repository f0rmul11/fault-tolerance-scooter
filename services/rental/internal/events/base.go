package events

// Base — envelope интеграционного события аренды. ID события стабилен при retry relay.
type Base[T any] struct {
	CorrelationID string `json:"correlation_id"`
	EventType     string `json:"event_type"`
	RentalID      string `json:"rental_id"`
	RentalVersion int    `json:"rental_version"`
	OccurredAt    string `json:"occurred_at"`
	RequestID     string `json:"request_id"`
	Payload       T      `json:"payload"`
}
