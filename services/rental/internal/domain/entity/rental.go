package entity

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"rental/internal/events"
	"rental/internal/events/rental_canceled"
	"rental/internal/events/rental_completed"
	"rental/internal/events/rental_reserved"
	"rental/internal/events/rental_started"
	"rental/internal/pkg/event"
	app "rental/internal/pkg/rentalerror"
)

type Rental struct {
	id             uuid.UUID
	userID         uuid.UUID
	scooterID      uuid.UUID
	idempotencyKey uuid.UUID
	requestID      uuid.UUID
	scooterNumber  string
	scooterBrand   string
	status         Status
	reservedAt     time.Time
	startedAt      time.Time
	finishedAt     time.Time
	updatedAt      time.Time
	version        version
}

func (e *Rental) ID() uuid.UUID {
	return e.id
}

func (e *Rental) UserID() uuid.UUID {
	return e.userID
}

func (e *Rental) ScooterID() uuid.UUID {
	return e.scooterID
}

func (e *Rental) IdempotencyKey() uuid.UUID {
	return e.idempotencyKey
}

func (e *Rental) RequestID() uuid.UUID {
	return e.requestID
}

func (e *Rental) ScooterNumber() string {
	return e.scooterNumber
}

func (e *Rental) ScooterBrand() string {
	return e.scooterBrand
}

func (e *Rental) Status() Status {
	return e.status
}

func (e *Rental) ReservedAt() time.Time {
	return e.reservedAt
}

func (e *Rental) StartedAt() time.Time {
	return e.startedAt
}

func (e *Rental) FinishedAt() time.Time {
	return e.finishedAt
}

func (e *Rental) UpdatedAt() time.Time {
	return e.updatedAt
}

func (e *Rental) OldVersion() int {
	return e.version.Old
}

func (e *Rental) Version() int {
	if e.version.new != nil {
		return *e.version.new
	}
	return e.version.Old
}

// NewRental создаёт бронь и добавляет событие в буфер текущей операции.
func NewRental(ctx context.Context, id, userID, scooterID, key, requestID uuid.UUID, number, brand string, isActive bool, at time.Time) (*Rental, error) {
	if id == uuid.Nil || userID == uuid.Nil || scooterID == uuid.Nil {
		return nil, app.NewInvalidArgumentErr("Обязательные ID должны быть UUID")
	}
	if key == uuid.Nil || requestID == uuid.Nil {
		return nil, app.NewInvalidArgumentErr("Ключ идемпотентности и request_id должны быть UUID")
	}
	if !isActive {
		return nil, app.NewBusinessErr("Самокат недоступен для бронирования")
	}
	if number == "" || brand == "" {
		return nil, app.NewBusinessErr("Каталог вернул некорректные данные самоката")
	}
	if at.IsZero() {
		return nil, app.NewInvalidArgumentErr("Время бронирования не задано")
	}
	at = at.UTC().Truncate(time.Microsecond)
	rental := &Rental{
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
	}
	event.Add(ctx, rental_reserved.New(rental.eventSnapshot()))
	return rental, nil
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

// Start фиксирует переход и его событие только при изменении состояния.
func (e *Rental) Start(ctx context.Context, requestID uuid.UUID, at time.Time) (bool, error) {
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
	event.Add(ctx, rental_started.New(e.eventSnapshot()))
	return true, nil
}

// Finish фиксирует переход и его событие только при изменении состояния.
func (e *Rental) Finish(ctx context.Context, requestID uuid.UUID, at time.Time) (bool, error) {
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
	event.Add(ctx, rental_completed.New(e.eventSnapshot()))
	return true, nil
}

// Cancel фиксирует переход и его событие только при изменении состояния.
func (e *Rental) Cancel(ctx context.Context, requestID uuid.UUID, at time.Time) (bool, error) {
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
	event.Add(ctx, rental_canceled.New(e.eventSnapshot()))
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

func (e *Rental) eventSnapshot() events.Rental {
	return events.Rental{
		ID:            e.id,
		UserID:        e.userID,
		ScooterID:     e.scooterID,
		RequestID:     e.requestID,
		ScooterNumber: e.scooterNumber,
		ScooterBrand:  e.scooterBrand,
		Status:        string(e.status),
		Version:       e.Version(),
		ReservedAt:    e.reservedAt,
		StartedAt:     e.startedAt,
		FinishedAt:    e.finishedAt,
		UpdatedAt:     e.updatedAt,
	}
}
