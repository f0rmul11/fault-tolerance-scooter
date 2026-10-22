package reserve

import (
	"time"

	"github.com/gofrs/uuid"
)

// Input содержит данные конкретного события после проверки транспортного envelope.
type Input struct {
	RentalID      uuid.UUID
	UserID        uuid.UUID
	ScooterID     uuid.UUID
	RentalVersion int
	ReservedAt    time.Time
}
