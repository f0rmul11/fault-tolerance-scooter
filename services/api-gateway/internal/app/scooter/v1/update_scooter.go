package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/scooter/v1"
	external "api-gateway/internal/pkg/pb/external/scooter/scooter/v1"
	"context"
)

func (i *Implementation) UpdateScooter(ctx context.Context, request *pb.UpdateScooterRequest) (*pb.UpdateScooterResponse, error) {
	response, err := i.external.UpdateScooter(ctx, &external.UpdateScooterRequest{
		ScooterId:       request.GetScooterId(),
		ExpectedVersion: request.GetExpectedVersion(),
		Number:          request.Number,
		Brand:           request.Brand,
		ChargePercent:   request.ChargePercent,
		RiddenSeconds:   request.RiddenSeconds,
	})
	if err != nil {
		return nil, err
	}
	return &pb.UpdateScooterResponse{
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
