package dao

import (
	"rental/internal/domain/entity"
	"rental/internal/pkg/xo"
)

type Rental struct {
	xo.Rental
}

func (row Rental) ConvertTo() *entity.Rental {
	return entity.With(
		row.ID,
		entity.WithUserID(row.UserID.UUID),
		entity.WithScooterID(row.ScooterID.UUID),
		entity.WithIdempotencyKey(row.IdempotencyKey.UUID),
		entity.WithRequestID(row.RequestID.UUID),
		entity.WithScooterNumber(row.ScooterNumber.String),
		entity.WithScooterBrand(row.ScooterBrand.String),
		entity.WithStatus(entity.Status(row.Status.String)),
		entity.WithReservedAt(row.ReservedAt.Time),
		entity.WithStartedAt(row.StartedAt.Time),
		entity.WithFinishedAt(row.FinishedAt.Time),
		entity.WithUpdatedAt(row.UpdatedAt.Time),
		entity.WithVersion(int(row.Version.Int64)),
	)
}
