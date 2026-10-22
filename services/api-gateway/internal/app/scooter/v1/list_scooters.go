package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/scooter/v1"
	external "api-gateway/internal/pkg/pb/external/scooter/scooter/v1"
	"context"

	"github.com/samber/lo"
)

func (i *Implementation) ListScooters(ctx context.Context, request *pb.ListScootersRequest) (*pb.ListScootersResponse, error) {
	response, err := i.external.ListScooters(ctx, &external.ListScootersRequest{
		Limit:  request.GetLimit(),
		Offset: request.GetOffset(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.ListScootersResponse{
		Scooters: lo.Map(response.GetScooters(), func(item *external.Scooter, _ int) *pb.Scooter {
			return &pb.Scooter{
				Id:            item.GetId(),
				Number:        item.GetNumber(),
				Brand:         item.GetBrand(),
				IsActive:      item.GetIsActive(),
				Owner:         item.GetOwner(),
				RiddenSeconds: item.GetRiddenSeconds(),
				Version:       item.GetVersion(),
				UpdatedAt:     item.GetUpdatedAt(),
			}
		}),
	}, nil
}
