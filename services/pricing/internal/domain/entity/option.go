package entity

import (
	"time"

	"github.com/gofrs/uuid"
)

type Option func(*Trip)

func With(id uuid.UUID, opts ...Option) *Trip {
	e := &Trip{rentalID: id}
	for _, opt := range opts {
		opt(e)
	}
	return e
}
func WithUserID(value uuid.UUID) Option {
	return func(e *Trip) {
		e.userID = value
	}
}
func WithScooterID(value uuid.UUID) Option {
	return func(e *Trip) {
		e.scooterID = value
	}
}
func WithPlan(value string) Option {
	return func(e *Trip) {
		e.plan = value
	}
}
func WithRateMinor(value int64) Option {
	return func(e *Trip) {
		e.rateMinor = value
	}
}
func WithAdvanceMinor(value int64) Option {
	return func(e *Trip) {
		e.advanceMinor = value
	}
}
func WithChargedMinor(value int64) Option {
	return func(e *Trip) {
		e.chargedMinor = value
	}
}
func WithTotalMinor(value int64) Option {
	return func(e *Trip) {
		e.totalMinor = value
	}
}
func WithAdjustmentMinor(value int64) Option {
	return func(e *Trip) {
		e.adjustmentMinor = value
	}
}
func WithBilledMinutes(value int64) Option {
	return func(e *Trip) {
		e.billedMinutes = value
	}
}
func WithStatus(value Status) Option {
	return func(e *Trip) {
		e.status = value
	}
}
func WithRentalVersion(value int) Option {
	return func(e *Trip) {
		e.rentalVersion = value
	}
}
func WithReservedAt(value time.Time) Option {
	return func(e *Trip) {
		e.reservedAt = value
	}
}
func WithStartedAt(value time.Time) Option {
	return func(e *Trip) {
		e.startedAt = value
	}
}
func WithFinishedAt(value time.Time) Option {
	return func(e *Trip) {
		e.finishedAt = value
	}
}
func WithVersion(value int) Option {
	return func(e *Trip) {
		e.version = version{Old: value}
	}
}
