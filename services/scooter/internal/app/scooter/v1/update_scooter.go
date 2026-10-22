package v1

import (
	"context"
	"scooter/internal/application/usecase/scooter/update"
	pb "scooter/internal/pkg/pb/scooter/scooter/v1"
	"time"

	"github.com/gofrs/uuid"
	"github.com/samber/mo"
)

func (i *Implementation) UpdateScooter(ctx context.Context, request *pb.UpdateScooterRequest) (*pb.UpdateScooterResponse, error) {
	var number mo.Option[string]
	if request.Number != nil {
		number = mo.Some(*request.Number)
	}
	var brand mo.Option[string]
	if request.Brand != nil {
		brand = mo.Some(*request.Brand)
	}
	var chargePercent mo.Option[int]
	if request.ChargePercent != nil {
		chargePercent = mo.Some(int(*request.ChargePercent))
	}
	var riddenSeconds mo.Option[int64]
	if request.RiddenSeconds != nil {
		riddenSeconds = mo.Some(*request.RiddenSeconds)
	}
	result, err := i.usecases.UpdateScooter.Execute(ctx, update.Input{
		ID:              uuid.FromStringOrNil(request.GetScooterId()),
		ExpectedVersion: int(request.GetExpectedVersion()),
		Number:          number,
		Brand:           brand,
		ChargePercent:   chargePercent,
		RiddenSeconds:   riddenSeconds,
	})
	if err != nil {
		return nil, err
	}
	return &pb.UpdateScooterResponse{
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
