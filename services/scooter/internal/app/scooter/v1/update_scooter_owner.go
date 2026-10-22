package v1

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"scooter/internal/application/usecase/scooter/owner"
	pb "scooter/internal/pkg/pb/scooter/scooter/v1"
)

func (i *Implementation) UpdateScooterOwner(ctx context.Context, request *pb.UpdateScooterOwnerRequest) (*pb.UpdateScooterOwnerResponse, error) {
	result, err := i.usecases.UpdateScooterOwner.Execute(ctx, owner.Input{
		ID:              uuid.FromStringOrNil(request.GetScooterId()),
		ExpectedVersion: int(request.GetExpectedVersion()),
		Owner:           request.GetOwner(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.UpdateScooterOwnerResponse{
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
