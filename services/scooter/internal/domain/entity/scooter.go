package entity

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid"

	app "scooter/internal/pkg/scootererror"
)

type Scooter struct {
	id            uuid.UUID
	number        string
	brand         string
	isActive      bool
	owner         string
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

func (e *Scooter) IsActive() bool {
	return e.isActive
}

func (e *Scooter) Owner() string {
	return e.owner
}

func (e *Scooter) RiddenSeconds() int64 {
	return e.riddenSeconds
}

func (e *Scooter) UpdatedAt() time.Time {
	return e.updatedAt
}

func (e *Scooter) OldVersion() int {
	return e.version.Old
}

func (e *Scooter) Version() int {
	if e.version.new != nil {
		return *e.version.new
	}
	return e.version.Old
}

// ExpectVersion задаёт условие записи; сравнение выполняется только в SQL UPDATE.
func (e *Scooter) ExpectVersion(expected int) {
	e.version.expected = &expected
}

// ExpectedVersion возвращает ожидаемую клиентом версию или версию прочитанного снимка.
func (e *Scooter) ExpectedVersion() int {
	if e.version.expected != nil {
		return *e.version.expected
	}
	return e.version.Old
}

// SetActive разрешает или запрещает новые бронирования.
func (e *Scooter) SetActive(isActive bool, at time.Time) {
	if isActive == e.isActive {
		return
	}
	e.isActive = isActive
	e.changed(at)
}

// ChangeOwner меняет название юридического лица — владельца самоката.
func (e *Scooter) ChangeOwner(owner string, at time.Time) error {
	owner = strings.TrimSpace(owner)
	if owner == "" || utf8.RuneCountInString(owner) > 200 {
		return app.NewInvalidArgumentErr("Название владельца должно содержать от 1 до 200 символов")
	}
	if owner == e.owner {
		return nil
	}
	e.owner = owner
	e.changed(at)
	return nil
}

// changed объединяет изменения полей в одну новую версию сохраняемого снимка.
func (e *Scooter) changed(at time.Time) {
	e.updatedAt = at.UTC().Truncate(time.Microsecond)
	e.version.inc()
}
