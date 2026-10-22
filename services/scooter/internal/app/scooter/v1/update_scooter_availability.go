package v1

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"scooter/internal/application/usecase/scooter/availability"
	pb "scooter/internal/pkg/pb/scooter/scooter/v1"
	app "scooter/internal/pkg/scootererror"
)

func (i *Implementation) UpdateScooterAvailability(ctx context.Context, request *pb.UpdateScooterAvailabilityRequest) (*pb.UpdateScooterAvailabilityResponse, error) {
	if request.IsActive == nil {
		return nil, app.NewInvalidArgumentErr("is_active должен быть указан")
	}
	result, err := i.usecases.UpdateScooterAvailability.Execute(ctx, availability.Input{
		ID:              uuid.FromStringOrNil(request.GetScooterId()),
		ExpectedVersion: int(request.GetExpectedVersion()),
		IsActive:        request.GetIsActive(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.UpdateScooterAvailabilityResponse{
		Scooter: &pb.Scooter{
			Id:            result.ID().String(),
			Number:        result.Number(),
			Brand:         result.Brand(),
			IsActive:      result.IsActive(),
			Owner:         result.Owner(),
			RiddenSeconds: result.RiddenSeconds(),
			Version:       int64(result.Version()),
			UpdatedAt:     result.UpdatedAt().Format(time.RFC3339Nano),
		},
	}, nil
}
