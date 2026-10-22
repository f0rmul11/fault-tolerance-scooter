package dao

import (
	"pricing/internal/domain/entity"
	"pricing/internal/pkg/xo"
)

type Trip struct {
	xo.Trip
}

func (row Trip) ConvertTo() *entity.Trip {
	return entity.With(row.RentalID,
		entity.WithUserID(row.UserID.UUID),
		entity.WithScooterID(row.ScooterID.UUID),
		entity.WithPlan(row.Plan.String),
		entity.WithRateMinor(row.RateMinor.Int64),
		entity.WithAdvanceMinor(row.AdvanceMinor.Int64),
		entity.WithChargedMinor(row.ChargedMinor.Int64),
		entity.WithTotalMinor(row.TotalMinor.Int64),
		entity.WithAdjustmentMinor(row.AdjustmentMinor.Int64),
		entity.WithBilledMinutes(row.BilledMinutes.Int64),
		entity.WithStatus(entity.Status(row.Status.String)),
		entity.WithRentalVersion(int(row.RentalVersion.Int64)),
		entity.WithReservedAt(row.ReservedAt.Time),
		entity.WithStartedAt(row.StartedAt.Time),
		entity.WithFinishedAt(row.FinishedAt.Time),
		entity.WithVersion(int(row.Version.Int64)),
	)
}
