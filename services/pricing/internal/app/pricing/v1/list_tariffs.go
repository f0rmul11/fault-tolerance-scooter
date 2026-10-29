package v1

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"os"
	pb "pricing/internal/pkg/pb/pricing/pricing/v1"
)

func (i *Implementation) ListTariffs(ctx context.Context, _ *pb.ListTariffsRequest) (*pb.ListTariffsResponse, error) {
	_ = grpc.SetHeader(ctx, metadata.Pairs("x-pricing-pod", os.Getenv("POD_NAME")))
	result, err := i.usecases.ListTariffs.Execute(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*pb.Tariff, 0, len(result))
	for _, item := range result {
		items = append(items, &pb.Tariff{
			Plan:         item.Plan,
			RateMinor:    item.RateMinor,
			AdvanceMinor: item.AdvanceMinor,
			Currency:     "RUB",
		})
	}
	return &pb.ListTariffsResponse{
		Tariffs: items,
	}, nil
}
