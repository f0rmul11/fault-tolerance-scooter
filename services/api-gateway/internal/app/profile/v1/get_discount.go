package v1

import (
	"context"

	pb "api-gateway/internal/pkg/pb/api-gateway/profile/v1"
	external "api-gateway/internal/pkg/pb/external/profile/profile/v1"
)

func (i *Implementation) GetDiscount(ctx context.Context, request *pb.GetDiscountRequest) (*pb.GetDiscountResponse, error) {
	result, err := i.external.GetDiscount(ctx, &external.GetDiscountRequest{UserId: request.GetUserId()})
	if err != nil {
		return nil, err
	}
	return &pb.GetDiscountResponse{
		DiscountPercent:     result.GetDiscountPercent(),
		CompletedTripsCount: result.CompletedTripsCount,
		Degraded:            result.GetDegraded(),
		Source:              result.GetSource(),
	}, nil
}
