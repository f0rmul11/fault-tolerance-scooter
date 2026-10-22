package entity

import (
	"time"

	"github.com/gofrs/uuid"
)

type User struct {
	id        uuid.UUID
	name      string
	email     string
	plan      Plan
	updatedAt time.Time
	version   version
}

func (e *User) ID() uuid.UUID {
	return e.id
}

func (e *User) Name() string {
	return e.name
}

func (e *User) Email() string {
	return e.email
}

func (e *User) Plan() Plan {
	return e.plan
}

func (e *User) UpdatedAt() time.Time {
	return e.updatedAt
}
