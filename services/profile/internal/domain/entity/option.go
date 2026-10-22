package entity

import (
	"time"

	"github.com/gofrs/uuid"
)

type Option func(*User)

func With(id uuid.UUID, opts ...Option) *User {
	e := &User{id: id}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func WithName(value string) Option {
	return func(e *User) {
		e.name = value
	}
}

func WithEmail(value string) Option {
	return func(e *User) {
		e.email = value
	}
}

func WithPlan(value Plan) Option {
	return func(e *User) {
		e.plan = value
	}
}

func WithUpdatedAt(value time.Time) Option {
	return func(e *User) {
		e.updatedAt = value
	}
}

func WithVersion(value int) Option {
	return func(e *User) {
		e.version = version{Old: value}
	}
}
