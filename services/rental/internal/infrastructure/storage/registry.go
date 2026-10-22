package storage

import (
	"rental/internal/infrastructure/storage/rental"
	"rental/internal/pkg/transaction/wrapper"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Registry struct {
	Rentals *rental.Storage
}

func NewRegistry(pool *pgxpool.Pool) *Registry {
	db := wrapper.NewDatabase(pool)
	return &Registry{
		Rentals: rental.New(db),
	}
}
