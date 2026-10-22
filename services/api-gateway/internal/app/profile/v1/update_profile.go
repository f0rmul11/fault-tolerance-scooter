package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/profile/v1"
	external "api-gateway/internal/pkg/pb/external/profile/profile/v1"
	"context"

	"github.com/samber/lo"
)

func (i *Implementation) UpdateProfile(ctx context.Context, request *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	var plan *external.SubscriptionPlan
	if request.Plan != nil {
		plan = lo.ToPtr(external.SubscriptionPlan(*request.Plan))
	}
	response, err := i.external.UpdateProfile(ctx, &external.UpdateProfileRequest{
		UserId:          request.GetUserId(),
		ExpectedVersion: request.GetExpectedVersion(),
		Name:            request.Name,
		Email:           request.Email,
		Plan:            plan,
	})
	if err != nil {
		return nil, err
	}
	return &pb.UpdateProfileResponse{
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
