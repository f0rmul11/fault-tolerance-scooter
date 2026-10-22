package v1

import (
	"context"
	"rental/internal/application/usecase/rental/start"
	pb "rental/internal/pkg/pb/rental/rental/v1"
	"time"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
)

func (i *Implementation) StartRental(ctx context.Context, request *pb.StartRentalRequest) (*pb.StartRentalResponse, error) {
	result, err := i.usecases.StartRental.Execute(ctx, start.Input{
		UserID:    uuid.FromStringOrNil(request.GetUserId()),
		RentalID:  uuid.FromStringOrNil(request.GetRentalId()),
		RequestID: uuid.FromStringOrNil(request.GetRequestId()),
	})
	if err != nil {
		return nil, err
	}
	object := result
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
	return &pb.StartRentalResponse{
		Rental: view,
	}, nil
}
