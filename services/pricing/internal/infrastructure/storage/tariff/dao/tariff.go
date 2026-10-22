package dao

import (
	"pricing/internal/dto/tariff"
	"pricing/internal/pkg/xo"
)

type Tariff struct {
	xo.Tariff
}

func (row Tariff) ConvertTo() tariff.Tariff {
	return tariff.Tariff{
		Plan:         row.Plan,
		RateMinor:    row.RateMinor.Int64,
		AdvanceMinor: row.AdvanceMinor.Int64,
	}
}
