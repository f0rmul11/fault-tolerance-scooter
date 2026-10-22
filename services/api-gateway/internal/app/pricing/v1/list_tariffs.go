package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/pricing/v1"
	external "api-gateway/internal/pkg/pb/external/pricing/pricing/v1"
	"context"

	"github.com/samber/lo"
)

func (i *Implementation) ListTariffs(ctx context.Context, request *pb.ListTariffsRequest) (*pb.ListTariffsResponse, error) {
	response, err := i.external.ListTariffs(ctx, &external.ListTariffsRequest{})
	if err != nil {
		return nil, err
	}
	return &pb.ListTariffsResponse{
		Tariffs: lo.Map(response.GetTariffs(), func(item *external.Tariff, _ int) *pb.Tariff {
			return &pb.Tariff{
				Plan:         item.GetPlan(),
				RateMinor:    item.GetRateMinor(),
				AdvanceMinor: item.GetAdvanceMinor(),
				Currency:     item.GetCurrency(),
			}
		}),
	}, nil
}
