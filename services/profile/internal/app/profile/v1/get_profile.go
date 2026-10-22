package v1

import (
	"context"
	"profile/internal/application/usecase/profile/get"
	pb "profile/internal/pkg/pb/profile/profile/v1"
	"time"

	"github.com/gofrs/uuid"
)

func (i *Implementation) GetProfile(ctx context.Context, request *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	result, err := i.usecases.GetProfile.Execute(ctx, get.Input{
		ID: uuid.FromStringOrNil(request.GetUserId()),
	})
	if err != nil {
		return nil, err
	}
	return &pb.GetProfileResponse{
		Profile: &pb.User{
			Id:        result.ID().String(),
			Name:      result.Name(),
			Email:     result.Email(),
			Plan:      pb.SubscriptionPlan(pb.SubscriptionPlan_value["SUBSCRIPTION_PLAN_"+string(result.Plan())]),
			Version:   int64(result.Version()),
			UpdatedAt: result.UpdatedAt().Format(time.RFC3339Nano),
		},
	}, nil
}
