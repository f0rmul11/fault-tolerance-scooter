package v1

import (
	"context"

	pb "api-gateway/internal/pkg/pb/api-gateway/scooter/v1"
	external "api-gateway/internal/pkg/pb/external/scooter/scooter/v1"
)

func (i *Implementation) UpdateScooterAvailability(ctx context.Context, request *pb.UpdateScooterAvailabilityRequest) (*pb.UpdateScooterAvailabilityResponse, error) {
	response, err := i.external.UpdateScooterAvailability(ctx, &external.UpdateScooterAvailabilityRequest{
		ScooterId:       request.GetScooterId(),
		ExpectedVersion: request.GetExpectedVersion(),
		IsActive:        request.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return &pb.UpdateScooterAvailabilityResponse{
		Scooter: &pb.Scooter{
			Id:            response.GetScooter().GetId(),
			Number:        response.GetScooter().GetNumber(),
			Brand:         response.GetScooter().GetBrand(),
			IsActive:      response.GetScooter().GetIsActive(),
			Owner:         response.GetScooter().GetOwner(),
			RiddenSeconds: response.GetScooter().GetRiddenSeconds(),
			Version:       response.GetScooter().GetVersion(),
			UpdatedAt:     response.GetScooter().GetUpdatedAt(),
		},
	}, nil
}
