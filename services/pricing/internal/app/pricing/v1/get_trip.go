package v1

import (
	"context"
	"pricing/internal/application/usecase/trip/get"
	"pricing/internal/domain/entity"
	pb "pricing/internal/pkg/pb/pricing/pricing/v1"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
)

func (i *Implementation) GetTrip(ctx context.Context, request *pb.GetTripRequest) (*pb.GetTripResponse, error) {
	result, err := i.usecases.GetTrip.Execute(ctx, get.Input{
		RentalID: uuid.FromStringOrNil(request.GetRentalId()),
		UserID:   uuid.FromStringOrNil(request.GetUserId()),
	})
	if err != nil {
		return nil, err
	}
	view := &pb.Trip{
		RentalId:      result.RentalID().String(),
		UserId:        result.UserID().String(),
		ScooterId:     result.ScooterID().String(),
		Plan:          result.Plan(),
		RateMinor:     result.RateMinor(),
		AdvanceMinor:  result.AdvanceMinor(),
		ChargedMinor:  result.ChargedMinor(),
		BilledMinutes: result.BilledMinutes(),
		Status:        pb.TripStatus(pb.TripStatus_value["TRIP_STATUS_"+string(result.Status())]),
		RentalVersion: int64(result.RentalVersion()),
		Currency:      "RUB",
	}
	if result.Status() == entity.Completed || result.Status() == entity.Canceled {
		view.TotalMinor = lo.ToPtr(result.TotalMinor())
		view.AdjustmentMinor = lo.ToPtr(result.AdjustmentMinor())
	}
	return &pb.GetTripResponse{
		Trip: view,
	}, nil
}
