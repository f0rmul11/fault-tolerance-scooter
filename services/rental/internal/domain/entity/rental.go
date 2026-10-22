package entity

import (
	app "rental/internal/pkg/rentalerror"
	"time"

	"github.com/gofrs/uuid"
)

func NewRental(id, userID, scooterID, key, requestID uuid.UUID, number, brand string, charge int, at time.Time) (*Rental, error) {
	if id == uuid.Nil || userID == uuid.Nil || scooterID == uuid.Nil {
		return nil, app.NewInvalidArgumentErr("Обязательные ID должны быть UUID")
	}
	if key == uuid.Nil || requestID == uuid.Nil {
		return nil, app.NewInvalidArgumentErr("Ключ идемпотентности и request_id должны быть UUID")
	}
	if charge < 20 {
		return nil, app.NewBusinessErr("Заряда самоката недостаточно для бронирования")
	}
	if charge > 100 || number == "" || brand == "" {
		return nil, app.NewBusinessErr("Каталог вернул некорректные данные самоката")
	}
	if at.IsZero() {
		return nil, app.NewInvalidArgumentErr("Время бронирования не задано")
	}
	at = at.UTC().Truncate(time.Microsecond)
	return &Rental{
		id:             id,
		userID:         userID,
		scooterID:      scooterID,
		idempotencyKey: key,
		requestID:      requestID,
		scooterNumber:  number,
		scooterBrand:   brand,
		status:         Reserved,
		reservedAt:     at,
		updatedAt:      at,
		version:        version{Old: 1},
	}, nil
}

func (e *Rental) CheckOwner(userID uuid.UUID) error {
	if userID == uuid.Nil {
		return app.NewInvalidArgumentErr("ID пользователя должен быть UUID")
	}
	if e.userID != userID {
		return app.New(app.WithCode(int(app.PermissionDenied)), app.WithMessage("Аренда принадлежит другому пользователю"))
	}
	return nil
}

func (e *Rental) MatchReservation(scooterID uuid.UUID) error {
	if e.scooterID != scooterID {
		return app.NewConflictErr("Ключ идемпотентности уже использован для другого самоката")
	}
	return nil
}

func (e *Rental) Start(requestID uuid.UUID, at time.Time) (bool, error) {
	if e.status == Active {
		return false, nil
	}
	if e.status != Reserved {
		return false, app.NewBusinessErr("Начать можно только забронированную поездку")
	}
	if err := e.validateAction(requestID, at); err != nil {
		return false, err
	}
	e.status = Active
	e.startedAt = at.UTC().Truncate(time.Microsecond)
	e.changed(requestID, at)
	return true, nil
}

func (e *Rental) Finish(requestID uuid.UUID, at time.Time) (bool, error) {
	if e.status == Completed {
		return false, nil
	}
	if e.status != Active {
		return false, app.NewBusinessErr("Завершить можно только начатую поездку")
	}
	if err := e.validateAction(requestID, at); err != nil {
		return false, err
	}
	e.status = Completed
	e.finishedAt = at.UTC().Truncate(time.Microsecond)
	e.changed(requestID, at)
	return true, nil
}

func (e *Rental) Cancel(requestID uuid.UUID, at time.Time) (bool, error) {
	if e.status == Canceled {
		return false, nil
	}
	if e.status != Reserved {
		return false, app.NewBusinessErr("Отменить можно только бронь до начала поездки")
	}
	if err := e.validateAction(requestID, at); err != nil {
		return false, err
	}
	e.status = Canceled
	e.changed(requestID, at)
	return true, nil
}

func (e *Rental) validateAction(requestID uuid.UUID, at time.Time) error {
	if requestID == uuid.Nil {
		return app.NewInvalidArgumentErr("request_id должен быть UUID")
	}
	if at.Before(e.updatedAt) || at.IsZero() {
		return app.NewBusinessErr("Время действия раньше предыдущего изменения аренды")
	}
	return nil
}

func (e *Rental) changed(requestID uuid.UUID, at time.Time) {
	e.requestID = requestID
	e.updatedAt = at.UTC().Truncate(time.Microsecond)
	e.version.inc()
}
