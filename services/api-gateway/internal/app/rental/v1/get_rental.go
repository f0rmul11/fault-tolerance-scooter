package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/rental/v1"
	external "api-gateway/internal/pkg/pb/external/rental/rental/v1"
	"context"
)

func (i *Implementation) GetRental(ctx context.Context, request *pb.GetRentalRequest) (*pb.GetRentalResponse, error) {
	response, err := i.external.GetRental(ctx, &external.GetRentalRequest{
		RentalId: request.GetRentalId(),
		UserId:   request.GetUserId(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.GetRentalResponse{
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
