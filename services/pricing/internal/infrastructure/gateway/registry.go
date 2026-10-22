package gateway

import (
	"pricing/config"
	"pricing/internal/infrastructure/gateway/profile"

	"google.golang.org/grpc"
)

type Registry struct {
	Profile *profile.Client
}

func NewRegistry(connections map[string]grpc.ClientConnInterface) *Registry {
	return &Registry{
		Profile: profile.NewClient(profile.NewExternalClient(connections[config.ProfileService])),
	}
}
