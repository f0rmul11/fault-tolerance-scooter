package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/rental/v1"
	external "api-gateway/internal/pkg/pb/external/rental/rental/v1"
	"api-gateway/internal/pkg/requestid"
	"context"
)

func (i *Implementation) FinishRental(ctx context.Context, request *pb.FinishRentalRequest) (*pb.FinishRentalResponse, error) {
	response, err := i.external.FinishRental(ctx, &external.FinishRentalRequest{
		RentalId:  request.GetRentalId(),
		UserId:    request.GetUserId(),
		RequestId: requestid.FromContext(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &pb.FinishRentalResponse{
		Rental: &pb.Rental{
			Id:            response.GetRental().GetId(),
			UserId:        response.GetRental().GetUserId(),
			ScooterId:     response.GetRental().GetScooterId(),
			ScooterNumber: response.GetRental().GetScooterNumber(),
			ScooterBrand:  response.GetRental().GetScooterBrand(),
			Status:        pb.RentalStatus(response.GetRental().GetStatus()),
			ReservedAt:    response.GetRental().GetReservedAt(),
			StartedAt:     response.GetRental().StartedAt,
			FinishedAt:    response.GetRental().FinishedAt,
			Version:       response.GetRental().GetVersion(),
		},
	}, nil
}
