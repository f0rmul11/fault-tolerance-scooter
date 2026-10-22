package gateway

import (
	"rental/config"
	"rental/internal/infrastructure/gateway/scooter"

	"google.golang.org/grpc"
)

type Registry struct {
	Scooter *scooter.Client
}

func NewRegistry(connections map[string]grpc.ClientConnInterface) *Registry {
	return &Registry{
		Scooter: scooter.NewClient(scooter.NewExternalClient(connections[config.ScooterService])),
	}
}
