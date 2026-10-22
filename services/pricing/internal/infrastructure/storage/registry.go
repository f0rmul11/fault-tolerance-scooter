package storage

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"pricing/internal/infrastructure/storage/pricing"
	"pricing/internal/infrastructure/storage/tariff"
	"pricing/internal/pkg/transaction/wrapper"
)

type Registry struct {
	Trips   *pricing.Storage
	Tariffs *tariff.Storage
}

func NewRegistry(pool *pgxpool.Pool) *Registry {
	db := wrapper.NewDatabase(pool)
	return &Registry{
		Trips:   pricing.New(db),
		Tariffs: tariff.New(db),
	}
}
