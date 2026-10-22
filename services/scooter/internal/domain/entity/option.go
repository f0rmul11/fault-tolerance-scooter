package entity

import (
	"time"

	"github.com/gofrs/uuid"
)

type Option func(*Scooter)

func With(id uuid.UUID, opts ...Option) *Scooter {
	e := &Scooter{id: id}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func WithNumber(value string) Option {
	return func(e *Scooter) {
		e.number = value
	}
}

func WithBrand(value string) Option {
	return func(e *Scooter) {
		e.brand = value
	}
}

func WithIsActive(value bool) Option {
	return func(e *Scooter) {
		e.isActive = value
	}
}

func WithOwner(value string) Option {
	return func(e *Scooter) {
		e.owner = value
	}
}

func WithRiddenSeconds(value int64) Option {
	return func(e *Scooter) {
		e.riddenSeconds = value
	}
}

func WithUpdatedAt(value time.Time) Option {
	return func(e *Scooter) {
		e.updatedAt = value
	}
}

func WithVersion(value int) Option {
	return func(e *Scooter) {
		e.version = version{Old: value}
	}
}
