package entity

import (
	"math"
	"time"

	"github.com/gofrs/uuid"

	app "pricing/internal/pkg/pricingerror"
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

func (e *Trip) OldVersion() int {
	return e.version.Old
}

func (e *Trip) Version() int {
	if e.version.new != nil {
		return *e.version.new
	}
	return e.version.Old
}

func NewTrip(rentalID, userID, scooterID uuid.UUID, plan string, rateMinor, advanceMinor int64, rentalVersion int, at time.Time) (*Trip, error) {
	if rentalID == uuid.Nil || userID == uuid.Nil || scooterID == uuid.Nil {
		return nil, app.NewInvalidArgumentErr("Некорректные идентификаторы поездки")
	}
	switch plan {
	case "BASE", "PREMIUM", "ULTRA":
	default:
		return nil, app.NewBusinessErr("Неизвестный тариф")
	}
	if rateMinor <= 0 || advanceMinor < 0 {
		return nil, app.NewBusinessErr("Некорректная стоимость тарифа")
	}
	if rentalVersion != 1 || at.IsZero() {
		return nil, app.NewInvalidArgumentErr("Некорректное событие бронирования")
	}
	return &Trip{
		rentalID:      rentalID,
		userID:        userID,
		scooterID:     scooterID,
		plan:          plan,
		rateMinor:     rateMinor,
		advanceMinor:  advanceMinor,
		status:        Reserved,
		rentalVersion: rentalVersion,
		reservedAt:    at,
		version:       version{Old: 1},
	}, nil
}

// CheckEvent защищает снимок от несовместимых повторов и пропуска предшественника.
func (e *Trip) CheckEvent(userID, scooterID uuid.UUID, rentalVersion int) (bool, error) {
	if e.userID != userID || e.scooterID != scooterID {
		return false, app.NewBusinessErr("Событие противоречит данным поездки")
	}
	if rentalVersion <= e.rentalVersion {
		return false, nil
	}
	if rentalVersion != e.rentalVersion+1 {
		return false, app.NewConflictErr("Предыдущее событие аренды ещё не обработано")
	}
	return true, nil
}

func (e *Trip) Start(version int, at time.Time) error {
	if e.status != Reserved || version != e.rentalVersion+1 {
		return app.NewConflictErr("Поездка ещё не готова к начислению аванса")
	}
	if at.Before(e.reservedAt) || at.IsZero() {
		return app.NewBusinessErr("Некорректное время начала поездки")
	}
	e.status = Active
	e.startedAt = at
	e.chargedMinor = e.advanceMinor
	e.rentalVersion = version
	e.version.inc()
	return nil
}

func (e *Trip) Complete(version int, startedAt, finishedAt time.Time) error {
	if e.status != Active || version != e.rentalVersion+1 {
		return app.NewConflictErr("Начало поездки ещё не обработано")
	}
	if !startedAt.Equal(e.startedAt) || finishedAt.Before(startedAt) {
		return app.NewBusinessErr("Некорректный период поездки")
	}
	duration := finishedAt.Sub(startedAt)
	if !startedAt.Add(duration).Equal(finishedAt) {
		return app.NewBusinessErr("Длительность поездки превышает допустимый диапазон")
	}
	minutes := int64(duration / time.Minute)
	if duration%time.Minute != 0 {
		minutes++
	}
	if minutes == 0 {
		minutes = 1
	}
	if e.rateMinor <= 0 || minutes > math.MaxInt64/e.rateMinor {
		return app.NewBusinessErr("Стоимость поездки превышает допустимый диапазон")
	}
	total := minutes * e.rateMinor
	e.status = Completed
	e.finishedAt = finishedAt
	e.billedMinutes = minutes
	e.totalMinor = total
	e.adjustmentMinor = total - e.advanceMinor
	e.chargedMinor = total
	e.rentalVersion = version
	e.version.inc()
	return nil
}

func (e *Trip) Cancel(version int) error {
	if e.status != Reserved || version != e.rentalVersion+1 {
		return app.NewConflictErr("Отмена противоречит состоянию поездки")
	}
	e.status = Canceled
	e.rentalVersion = version
	e.version.inc()
	return nil
}

func (e *Trip) CheckOwner(userID uuid.UUID) error {
	if e.userID != userID {
		return app.New(app.WithCode(int(app.PermissionDenied)), app.WithMessage("Расчёт принадлежит другому пользователю"))
	}
	return nil
}
