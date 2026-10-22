package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/profile/v1"
	external "api-gateway/internal/pkg/pb/external/profile/profile/v1"
)

type Implementation struct {
	pb.UnimplementedProfileServiceServer
	external external.ProfileServiceClient
}

func NewProfileService(external external.ProfileServiceClient) *Implementation {
	return &Implementation{external: external}
}
