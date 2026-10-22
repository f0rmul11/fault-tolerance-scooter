package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/rental/v1"
	external "api-gateway/internal/pkg/pb/external/rental/rental/v1"
)

type Implementation struct {
	pb.UnimplementedRentalServiceServer
	external external.RentalServiceClient
}

func NewRentalService(external external.RentalServiceClient) *Implementation {
	return &Implementation{external: external}
}
