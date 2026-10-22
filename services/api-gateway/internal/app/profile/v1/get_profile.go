package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/profile/v1"
	external "api-gateway/internal/pkg/pb/external/profile/profile/v1"
	"context"
)

func (i *Implementation) GetProfile(ctx context.Context, request *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	response, err := i.external.GetProfile(ctx, &external.GetProfileRequest{
		UserId: request.GetUserId(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.GetProfileResponse{
		Profile: &pb.User{
			Id:        response.GetProfile().GetId(),
			Name:      response.GetProfile().GetName(),
			Email:     response.GetProfile().GetEmail(),
			Plan:      pb.SubscriptionPlan(response.GetProfile().GetPlan()),
			Version:   response.GetProfile().GetVersion(),
			UpdatedAt: response.GetProfile().GetUpdatedAt(),
		},
	}, nil
}
