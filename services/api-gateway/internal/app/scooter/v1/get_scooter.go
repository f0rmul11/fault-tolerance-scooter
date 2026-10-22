package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/scooter/v1"
	external "api-gateway/internal/pkg/pb/external/scooter/scooter/v1"
	"context"
)

func (i *Implementation) GetScooter(ctx context.Context, request *pb.GetScooterRequest) (*pb.GetScooterResponse, error) {
	response, err := i.external.GetScooter(ctx, &external.GetScooterRequest{
		ScooterId: request.GetScooterId(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.GetScooterResponse{
		Scooter: &pb.Scooter{
			Id:            response.GetScooter().GetId(),
			Number:        response.GetScooter().GetNumber(),
			Brand:         response.GetScooter().GetBrand(),
			ChargePercent: response.GetScooter().GetChargePercent(),
			RiddenSeconds: response.GetScooter().GetRiddenSeconds(),
			Version:       response.GetScooter().GetVersion(),
			UpdatedAt:     response.GetScooter().GetUpdatedAt(),
		},
	}, nil
}
