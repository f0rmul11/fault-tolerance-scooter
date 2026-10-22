package storage

import (
	"profile/internal/infrastructure/storage/profile"
	"profile/internal/pkg/transaction/wrapper"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Registry struct {
	Profiles *profile.Storage
}

func NewRegistry(pool *pgxpool.Pool) *Registry {
	db := wrapper.NewDatabase(pool)
	return &Registry{
		Profiles: profile.New(db),
	}
}
