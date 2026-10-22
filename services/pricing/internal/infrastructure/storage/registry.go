package storage

import (
	"pricing/internal/infrastructure/storage/pricing"
	"pricing/internal/infrastructure/storage/tariff"
	"pricing/internal/pkg/transaction/wrapper"

	"github.com/jackc/pgx/v5/pgxpool"
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
