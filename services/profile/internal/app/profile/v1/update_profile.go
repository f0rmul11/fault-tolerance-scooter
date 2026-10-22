package v1

import (
	"context"
	"profile/internal/application/usecase/profile/update"
	"profile/internal/domain/entity"
	pb "profile/internal/pkg/pb/profile/profile/v1"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/samber/mo"
)

func (i *Implementation) UpdateProfile(ctx context.Context, request *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	var name mo.Option[string]
	if request.Name != nil {
		name = mo.Some(*request.Name)
	}
	var email mo.Option[string]
	if request.Email != nil {
		email = mo.Some(*request.Email)
	}
	var plan mo.Option[entity.Plan]
	if request.Plan != nil {
		plan = mo.Some(entity.Plan(strings.TrimPrefix(request.GetPlan().String(), "SUBSCRIPTION_PLAN_")))
	}
	result, err := i.usecases.UpdateProfile.Execute(ctx, update.Input{
		ID:              uuid.FromStringOrNil(request.GetUserId()),
		ExpectedVersion: int(request.GetExpectedVersion()),
		Name:            name,
		Email:           email,
		Plan:            plan,
	})
	if err != nil {
		return nil, err
	}
	return &pb.UpdateProfileResponse{
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
