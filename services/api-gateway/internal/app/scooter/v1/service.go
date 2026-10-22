package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/scooter/v1"
	external "api-gateway/internal/pkg/pb/external/scooter/scooter/v1"
)

type Implementation struct {
	pb.UnimplementedScooterServiceServer
	external external.ScooterServiceClient
}

func NewScooterService(external external.ScooterServiceClient) *Implementation {
	return &Implementation{external: external}
}
