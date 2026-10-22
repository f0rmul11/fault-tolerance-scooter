package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/profile/v1"
	external "api-gateway/internal/pkg/pb/external/profile/profile/v1"
	"context"

	"github.com/samber/lo"
)

func (i *Implementation) ListProfiles(ctx context.Context, request *pb.ListProfilesRequest) (*pb.ListProfilesResponse, error) {
	response, err := i.external.ListProfiles(ctx, &external.ListProfilesRequest{
		Limit:  request.GetLimit(),
		Offset: request.GetOffset(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.ListProfilesResponse{
		Profiles: lo.Map(response.GetProfiles(), func(item *external.User, _ int) *pb.User {
			return &pb.User{
				Id:        item.GetId(),
				Name:      item.GetName(),
				Email:     item.GetEmail(),
				Plan:      pb.SubscriptionPlan(item.GetPlan()),
				Version:   item.GetVersion(),
				UpdatedAt: item.GetUpdatedAt(),
			}
		}),
	}, nil
}
