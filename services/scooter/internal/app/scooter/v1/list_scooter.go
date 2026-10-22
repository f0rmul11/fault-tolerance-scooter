package v1

import (
	"context"
	"scooter/internal/application/usecase/scooter/list"
	pb "scooter/internal/pkg/pb/scooter/scooter/v1"
	"time"
)

func (i *Implementation) ListScooters(ctx context.Context, request *pb.ListScootersRequest) (*pb.ListScootersResponse, error) {
	result, err := i.usecases.ListScooters.Execute(ctx, list.Input{
		Limit:  int(request.GetLimit()),
		Offset: int(request.GetOffset()),
	})
	if err != nil {
		return nil, err
	}
	items := make([]*pb.Scooter, 0, len(result))
	for _, object := range result {
		items = append(items, &pb.Scooter{
			Id:            object.ID().String(),
			Number:        object.Number(),
			Brand:         object.Brand(),
			IsActive:      object.IsActive(),
			Owner:         object.Owner(),
			RiddenSeconds: object.RiddenSeconds(),
			Version:       int64(object.Version()),
			UpdatedAt:     object.UpdatedAt().Format(time.RFC3339Nano),
		})
	}
	return &pb.ListScootersResponse{
		Scooters: items,
	}, nil
}
