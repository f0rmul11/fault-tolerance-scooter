package entity

import (
	"net/mail"
	app "profile/internal/pkg/profileerror"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/samber/mo"
)

// Update проверяет весь набор изменений до изменения агрегата.
func (e *User) Update(name, email mo.Option[string], plan mo.Option[Plan], expectedVersion int, at time.Time) error {
	if e.id == uuid.Nil {
		return app.NewInvalidArgumentErr("Некорректный ID пользователя")
	}
	if expectedVersion != e.Version() {
		return app.NewConflictErr("Профиль уже изменён; перечитайте его")
	}
	nextName, nextEmail, nextPlan := e.name, e.email, e.plan
	if value, ok := name.Get(); ok {
		nextName = strings.TrimSpace(value)
	}
	if value, ok := email.Get(); ok {
		nextEmail = strings.TrimSpace(value)
	}
	if value, ok := plan.Get(); ok {
		nextPlan = value
	}
	if nextName == "" || len(nextName) > 200 {
		return app.NewInvalidArgumentErr("Имя должно содержать от 1 до 200 байт")
	}
	address, err := mail.ParseAddress(nextEmail)
	if err != nil || address.Address != nextEmail || len(nextEmail) > 320 {
		return app.NewInvalidArgumentErr("Некорректный email")
	}
	switch nextPlan {
	case Base, Premium, Ultra:
	default:
		return app.NewInvalidArgumentErr("Неизвестная подписка")
	}
	e.name, e.email, e.plan = nextName, nextEmail, nextPlan
	e.updatedAt = at.UTC().Truncate(time.Microsecond)
	e.version.inc()
	return nil
}
