package events

import (
	"time"

	"github.com/gofrs/uuid"
)

// Rental — снимок перехода. Последующие изменения сущности не меняют событие.
type Rental struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	ScooterID     uuid.UUID
	RequestID     uuid.UUID
	ScooterNumber string
	ScooterBrand  string
	Status        string
	Version       int
	ReservedAt    time.Time
	StartedAt     time.Time
	FinishedAt    time.Time
	UpdatedAt     time.Time
}
