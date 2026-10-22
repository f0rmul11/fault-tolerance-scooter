package v1

import (
	"context"
	"profile/internal/application/usecase/profile/list"
	pb "profile/internal/pkg/pb/profile/profile/v1"
	"time"
)

func (i *Implementation) ListProfiles(ctx context.Context, request *pb.ListProfilesRequest) (*pb.ListProfilesResponse, error) {
	result, err := i.usecases.ListProfiles.Execute(ctx, list.Input{
		Limit:  int(request.GetLimit()),
		Offset: int(request.GetOffset()),
	})
	if err != nil {
		return nil, err
	}
	items := make([]*pb.User, 0, len(result))
	for _, object := range result {
		items = append(items, &pb.User{
			Id:        object.ID().String(),
			Name:      object.Name(),
			Email:     object.Email(),
			Plan:      pb.SubscriptionPlan(pb.SubscriptionPlan_value["SUBSCRIPTION_PLAN_"+string(object.Plan())]),
			Version:   int64(object.Version()),
			UpdatedAt: object.UpdatedAt().Format(time.RFC3339Nano),
		})
	}
	return &pb.ListProfilesResponse{
		Profiles: items,
	}, nil
}
