package v1

import (
	"context"
	"scooter/internal/application/usecase/scooter/get"
	pb "scooter/internal/pkg/pb/scooter/scooter/v1"
	"time"

	"github.com/gofrs/uuid"
)

func (i *Implementation) GetScooter(ctx context.Context, request *pb.GetScooterRequest) (*pb.GetScooterResponse, error) {
	result, err := i.usecases.GetScooter.Execute(ctx, get.Input{
		ID: uuid.FromStringOrNil(request.GetScooterId()),
	})
	if err != nil {
		return nil, err
	}
	return &pb.GetScooterResponse{
		Scooter: &pb.Scooter{
			Id:            result.ID().String(),
			Number:        result.Number(),
			Brand:         result.Brand(),
			ChargePercent: int32(result.ChargePercent()),
			RiddenSeconds: result.RiddenSeconds(),
			Version:       int64(result.Version()),
			UpdatedAt:     result.UpdatedAt().Format(time.RFC3339Nano),
		},
	}, nil
}
