package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/rental/v1"
	external "api-gateway/internal/pkg/pb/external/rental/rental/v1"
	"context"

	"github.com/samber/lo"
)

func (i *Implementation) ListRentals(ctx context.Context, request *pb.ListRentalsRequest) (*pb.ListRentalsResponse, error) {
	response, err := i.external.ListRentals(ctx, &external.ListRentalsRequest{
		UserId: request.GetUserId(),
		Limit:  request.GetLimit(),
		Offset: request.GetOffset(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.ListRentalsResponse{
		Rentals: lo.Map(response.GetRentals(), func(item *external.Rental, _ int) *pb.Rental {
			return &pb.Rental{
				Id:            item.GetId(),
				UserId:        item.GetUserId(),
				ScooterId:     item.GetScooterId(),
				ScooterNumber: item.GetScooterNumber(),
				ScooterBrand:  item.GetScooterBrand(),
				Status:        pb.RentalStatus(item.GetStatus()),
				ReservedAt:    item.GetReservedAt(),
				StartedAt:     item.StartedAt,
				FinishedAt:    item.FinishedAt,
				Version:       item.GetVersion(),
			}
		}),
	}, nil
}
