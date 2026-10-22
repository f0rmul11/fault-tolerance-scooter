package start

import (
	"time"

	"github.com/gofrs/uuid"
)

type Input struct {
	RentalID      uuid.UUID
	UserID        uuid.UUID
	ScooterID     uuid.UUID
	RentalVersion int
	ReservedAt    time.Time
	StartedAt     time.Time
	FinishedAt    time.Time
}
