package storage

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"scooter/internal/infrastructure/storage/scooter"
	"scooter/internal/pkg/transaction/wrapper"
)

type Registry struct {
	Scooters *scooter.Storage
}

func NewRegistry(pool *pgxpool.Pool) *Registry {
	db := wrapper.NewDatabase(pool)
	return &Registry{
		Scooters: scooter.New(db),
	}
}
