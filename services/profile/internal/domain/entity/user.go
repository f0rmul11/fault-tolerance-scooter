package entity

import (
	"net/mail"
	"strings"
	"time"

	"github.com/gofrs/uuid"

	app "profile/internal/pkg/profileerror"
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

func (e *User) OldVersion() int {
	return e.version.Old
}

func (e *User) Version() int {
	if e.version.new != nil {
		return *e.version.new
	}
	return e.version.Old
}

// ExpectVersion задаёт условие записи; сравнение выполняется только в SQL UPDATE.
func (e *User) ExpectVersion(expected int) {
	e.version.expected = &expected
}

// ExpectedVersion возвращает ожидаемую клиентом версию или версию прочитанного снимка.
func (e *User) ExpectedVersion() int {
	if e.version.expected != nil {
		return *e.version.expected
	}
	return e.version.Old
}

// Rename изменяет только имя пользователя.
func (e *User) Rename(name string, at time.Time) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 200 {
		return app.NewInvalidArgumentErr("Имя должно содержать от 1 до 200 байт")
	}
	if name == e.name {
		return nil
	}
	e.name = name
	e.changed(at)
	return nil
}

// ChangeEmail проверяет адрес до изменения профиля.
func (e *User) ChangeEmail(email string, at time.Time) error {
	email = strings.TrimSpace(email)
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 320 {
		return app.NewInvalidArgumentErr("Некорректный email")
	}
	if email == e.email {
		return nil
	}
	e.email = email
	e.changed(at)
	return nil
}

// ChangeSubscription меняет подписку, сохраняя остальные данные пользователя.
func (e *User) ChangeSubscription(plan Plan, at time.Time) error {
	switch plan {
	case Base, Premium, Ultra:
	default:
		return app.NewInvalidArgumentErr("Неизвестная подписка")
	}
	if plan == e.plan {
		return nil
	}
	e.plan = plan
	e.changed(at)
	return nil
}

// changed объединяет изменения полей в одну новую версию сохраняемого снимка.
func (e *User) changed(at time.Time) {
	e.updatedAt = at.UTC().Truncate(time.Microsecond)
	e.version.inc()
}
