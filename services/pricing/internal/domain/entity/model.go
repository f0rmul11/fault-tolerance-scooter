package entity

import (
	"time"

	"github.com/gofrs/uuid"
)

type Trip struct {
	rentalID        uuid.UUID
	userID          uuid.UUID
	scooterID       uuid.UUID
	plan            string
	rateMinor       int64
	advanceMinor    int64
	chargedMinor    int64
	totalMinor      int64
	adjustmentMinor int64
	billedMinutes   int64
	status          Status
	rentalVersion   int
	reservedAt      time.Time
	startedAt       time.Time
	finishedAt      time.Time
	version         version
}

func (e *Trip) RentalID() uuid.UUID {
	return e.rentalID
}

func (e *Trip) UserID() uuid.UUID {
	return e.userID
}

func (e *Trip) ScooterID() uuid.UUID {
	return e.scooterID
}

func (e *Trip) Plan() string {
	return e.plan
}

func (e *Trip) RateMinor() int64 {
	return e.rateMinor
}

func (e *Trip) AdvanceMinor() int64 {
	return e.advanceMinor
}

func (e *Trip) ChargedMinor() int64 {
	return e.chargedMinor
}

func (e *Trip) TotalMinor() int64 {
	return e.totalMinor
}

func (e *Trip) AdjustmentMinor() int64 {
	return e.adjustmentMinor
}

func (e *Trip) BilledMinutes() int64 {
	return e.billedMinutes
}

func (e *Trip) Status() Status {
	return e.status
}

func (e *Trip) RentalVersion() int {
	return e.rentalVersion
}

func (e *Trip) ReservedAt() time.Time {
	return e.reservedAt
}

func (e *Trip) StartedAt() time.Time {
	return e.startedAt
}

func (e *Trip) FinishedAt() time.Time {
	return e.finishedAt
}
