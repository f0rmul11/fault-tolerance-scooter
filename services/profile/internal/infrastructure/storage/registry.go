package storage

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"profile/internal/infrastructure/storage/profile"
	"profile/internal/pkg/transaction/wrapper"
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
