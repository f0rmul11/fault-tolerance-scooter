package dao

import (
	"scooter/internal/domain/entity"
	"scooter/internal/pkg/xo"
)

type Scooter struct {
	xo.Scooter
}

func (row Scooter) ConvertTo() *entity.Scooter {
	return entity.With(row.ID,
		entity.WithNumber(row.Number.String),
		entity.WithBrand(row.Brand.String),
		entity.WithChargePercent(int(row.ChargePercent.Int64)),
		entity.WithRiddenSeconds(row.RiddenSeconds.Int64),
		entity.WithUpdatedAt(row.UpdatedAt),
		entity.WithVersion(int(row.Version.Int64)),
	)
}
