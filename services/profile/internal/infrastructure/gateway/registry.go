package gateway

import (
	"google.golang.org/grpc"

	"profile/config"
	"profile/internal/infrastructure/gateway/pricing"
)

type Registry struct {
	Pricing *pricing.Client
}

func NewRegistry(conn map[string]grpc.ClientConnInterface) *Registry {
	return &Registry{
		Pricing: pricing.NewClient(pricing.NewExternalClient(conn[config.PricingService])),
	}
}
