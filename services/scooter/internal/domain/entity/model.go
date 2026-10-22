package entity

import (
	"time"

	"github.com/gofrs/uuid"
)

type Scooter struct {
	id            uuid.UUID
	number        string
	brand         string
	chargePercent int
	riddenSeconds int64
	updatedAt     time.Time
	version       version
}

func (e *Scooter) ID() uuid.UUID {
	return e.id
}

func (e *Scooter) Number() string {
	return e.number
}

func (e *Scooter) Brand() string {
	return e.brand
}

func (e *Scooter) ChargePercent() int {
	return e.chargePercent
}

func (e *Scooter) RiddenSeconds() int64 {
	return e.riddenSeconds
}

func (e *Scooter) UpdatedAt() time.Time {
	return e.updatedAt
}
