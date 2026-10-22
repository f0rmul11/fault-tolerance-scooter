package v1

import (
	"context"
	"rental/internal/application/usecase/rental/list"
	pb "rental/internal/pkg/pb/rental/rental/v1"
	"time"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
)

func (i *Implementation) ListRentals(ctx context.Context, request *pb.ListRentalsRequest) (*pb.ListRentalsResponse, error) {
	result, err := i.usecases.ListRentals.Execute(ctx, list.Input{
		UserID: uuid.FromStringOrNil(request.GetUserId()),
		Limit:  int(request.GetLimit()),
		Offset: int(request.GetOffset()),
	})
	if err != nil {
		return nil, err
	}
	items := make([]*pb.Rental, 0, len(result))
	for _, object := range result {
		view := &pb.Rental{
			Id:            object.ID().String(),
			UserId:        object.UserID().String(),
			ScooterId:     object.ScooterID().String(),
			ScooterNumber: object.ScooterNumber(),
			ScooterBrand:  object.ScooterBrand(),
			Status:        pb.RentalStatus(pb.RentalStatus_value["RENTAL_STATUS_"+string(object.Status())]),
			ReservedAt:    object.ReservedAt().Format(time.RFC3339Nano),
			Version:       int64(object.Version()),
		}
		if !object.StartedAt().IsZero() {
			view.StartedAt = lo.ToPtr(object.StartedAt().Format(time.RFC3339Nano))
		}
		if !object.FinishedAt().IsZero() {
			view.FinishedAt = lo.ToPtr(object.FinishedAt().Format(time.RFC3339Nano))
		}
		items = append(items, view)
	}
	return &pb.ListRentalsResponse{
		Rentals: items,
	}, nil
}
