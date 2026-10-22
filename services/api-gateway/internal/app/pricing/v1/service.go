package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/pricing/v1"
	external "api-gateway/internal/pkg/pb/external/pricing/pricing/v1"
)

type Implementation struct {
	pb.UnimplementedPricingServiceServer
	external external.PricingServiceClient
}

func NewPricingService(external external.PricingServiceClient) *Implementation {
	return &Implementation{external: external}
}
