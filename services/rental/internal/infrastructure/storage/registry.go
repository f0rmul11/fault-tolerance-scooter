package storage

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"rental/internal/infrastructure/storage/rental"
	"rental/internal/pkg/transaction/wrapper"
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
