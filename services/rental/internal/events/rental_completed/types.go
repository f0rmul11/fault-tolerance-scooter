package rental_completed

// Payload — снимок аренды; общие идентификаторы события находятся в envelope.
type Payload struct {
	UserID        string `json:"user_id"`
	ScooterID     string `json:"scooter_id"`
	ScooterNumber string `json:"scooter_number"`
	ScooterBrand  string `json:"scooter_brand"`
	Status        string `json:"status"`
	ReservedAt    string `json:"reserved_at"`
	StartedAt     string `json:"started_at,omitempty"`
	FinishedAt    string `json:"finished_at,omitempty"`
}
