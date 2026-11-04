package v1

import (
	"context"

	"github.com/gofrs/uuid"

	pb "pricing/internal/pkg/pb/pricing/pricing/v1"
)

func (i *Implementation) GetCompletedTripsCount(ctx context.Context, request *pb.GetCompletedTripsCountRequest) (*pb.GetCompletedTripsCountResponse, error) {
	count, err := i.usecases.GetCompletedTripsCount.Execute(ctx, uuid.FromStringOrNil(request.GetUserId()))
	if err != nil {
		return nil, err
	}
	return &pb.GetCompletedTripsCountResponse{Count: count}, nil
}
