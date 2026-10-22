package entity

import (
	"time"

	"github.com/gofrs/uuid"
)

type Option func(*Rental)

func With(id uuid.UUID, opts ...Option) *Rental {
	e := &Rental{id: id}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func WithUserID(value uuid.UUID) Option {
	return func(e *Rental) {
		e.userID = value
	}
}

func WithScooterID(value uuid.UUID) Option {
	return func(e *Rental) {
		e.scooterID = value
	}
}

func WithIdempotencyKey(value uuid.UUID) Option {
	return func(e *Rental) {
		e.idempotencyKey = value
	}
}

func WithRequestID(value uuid.UUID) Option {
	return func(e *Rental) {
		e.requestID = value
	}
}

func WithScooterNumber(value string) Option {
	return func(e *Rental) {
		e.scooterNumber = value
	}
}

func WithScooterBrand(value string) Option {
	return func(e *Rental) {
		e.scooterBrand = value
	}
}

func WithStatus(value Status) Option {
	return func(e *Rental) {
		e.status = value
	}
}

func WithReservedAt(value time.Time) Option {
	return func(e *Rental) {
		e.reservedAt = value
	}
}

func WithStartedAt(value time.Time) Option {
	return func(e *Rental) {
		e.startedAt = value
	}
}

func WithFinishedAt(value time.Time) Option {
	return func(e *Rental) {
		e.finishedAt = value
	}
}

func WithUpdatedAt(value time.Time) Option {
	return func(e *Rental) {
		e.updatedAt = value
	}
}

func WithVersion(value int) Option {
	return func(e *Rental) {
		e.version = version{Old: value}
	}
}
