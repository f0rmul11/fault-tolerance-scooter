package dao

import (
	"scooter/internal/domain/entity"
	"scooter/internal/pkg/xo"
)

type Scooter struct {
	xo.Scooter
}

func (row Scooter) ConvertTo() *entity.Scooter {
	return entity.With(
		row.ID,
		entity.WithNumber(row.Number.String),
		entity.WithBrand(row.Brand.String),
		entity.WithIsActive(row.IsActive.Bool),
		entity.WithOwner(row.Owner.String),
		entity.WithRiddenSeconds(row.RiddenSeconds.Int64),
		entity.WithUpdatedAt(row.UpdatedAt),
		entity.WithVersion(int(row.Version.Int64)),
	)
}
