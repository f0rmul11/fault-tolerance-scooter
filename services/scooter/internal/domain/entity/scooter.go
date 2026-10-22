package entity

import (
	app "scooter/internal/pkg/scootererror"
	"strings"
	"time"

	"github.com/samber/mo"
)

func (e *Scooter) Update(number, brand mo.Option[string], charge mo.Option[int], ridden mo.Option[int64], expectedVersion int, at time.Time) error {
	if expectedVersion != e.Version() {
		return app.NewConflictErr("Самокат уже изменён; перечитайте его")
	}
	nextNumber, nextBrand, nextCharge, nextRidden := e.number, e.brand, e.chargePercent, e.riddenSeconds
	if value, ok := number.Get(); ok {
		nextNumber = strings.TrimSpace(value)
	}
	if value, ok := brand.Get(); ok {
		nextBrand = strings.TrimSpace(value)
	}
	if value, ok := charge.Get(); ok {
		nextCharge = value
	}
	if value, ok := ridden.Get(); ok {
		nextRidden = value
	}
	if nextNumber == "" || len(nextNumber) > 40 {
		return app.NewInvalidArgumentErr("Номер самоката должен содержать от 1 до 40 байт")
	}
	if nextBrand == "" || len(nextBrand) > 100 {
		return app.NewInvalidArgumentErr("Марка должна содержать от 1 до 100 байт")
	}
	if nextCharge < 0 || nextCharge > 100 {
		return app.NewInvalidArgumentErr("Заряд должен находиться в диапазоне 0–100 процентов")
	}
	if nextRidden < e.riddenSeconds || nextRidden < 0 {
		return app.NewBusinessErr("Счётчик наезда нельзя уменьшить")
	}
	e.number, e.brand, e.chargePercent, e.riddenSeconds = nextNumber, nextBrand, nextCharge, nextRidden
	e.updatedAt = at.UTC().Truncate(time.Microsecond)
	e.version.inc()
	return nil
}
