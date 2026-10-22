package entity

import (
	"time"

	"github.com/gofrs/uuid"
)

type Rental struct {
	id             uuid.UUID
	userID         uuid.UUID
	scooterID      uuid.UUID
	idempotencyKey uuid.UUID
	requestID      uuid.UUID
	scooterNumber  string
	scooterBrand   string
	status         Status
	reservedAt     time.Time
	startedAt      time.Time
	finishedAt     time.Time
	updatedAt      time.Time
	version        version
}

func (e *Rental) ID() uuid.UUID {
	return e.id
}

func (e *Rental) UserID() uuid.UUID {
	return e.userID
}

func (e *Rental) ScooterID() uuid.UUID {
	return e.scooterID
}

func (e *Rental) IdempotencyKey() uuid.UUID {
	return e.idempotencyKey
}

func (e *Rental) RequestID() uuid.UUID {
	return e.requestID
}

func (e *Rental) ScooterNumber() string {
	return e.scooterNumber
}

func (e *Rental) ScooterBrand() string {
	return e.scooterBrand
}

func (e *Rental) Status() Status {
	return e.status
}

func (e *Rental) ReservedAt() time.Time {
	return e.reservedAt
}

func (e *Rental) StartedAt() time.Time {
	return e.startedAt
}

func (e *Rental) FinishedAt() time.Time {
	return e.finishedAt
}

func (e *Rental) UpdatedAt() time.Time {
	return e.updatedAt
}
