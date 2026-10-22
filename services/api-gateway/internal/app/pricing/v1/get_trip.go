package v1

import (
	pb "api-gateway/internal/pkg/pb/api-gateway/pricing/v1"
	external "api-gateway/internal/pkg/pb/external/pricing/pricing/v1"
	"context"
)

func (i *Implementation) GetTrip(ctx context.Context, request *pb.GetTripRequest) (*pb.GetTripResponse, error) {
	response, err := i.external.GetTrip(ctx, &external.GetTripRequest{
		RentalId: request.GetRentalId(),
		UserId:   request.GetUserId(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.GetTripResponse{
		Trip: &pb.Trip{
			RentalId:        response.GetTrip().GetRentalId(),
			UserId:          response.GetTrip().GetUserId(),
			ScooterId:       response.GetTrip().GetScooterId(),
			Plan:            response.GetTrip().GetPlan(),
			RateMinor:       response.GetTrip().GetRateMinor(),
			AdvanceMinor:    response.GetTrip().GetAdvanceMinor(),
			ChargedMinor:    response.GetTrip().GetChargedMinor(),
			TotalMinor:      response.GetTrip().TotalMinor,
			AdjustmentMinor: response.GetTrip().AdjustmentMinor,
			BilledMinutes:   response.GetTrip().GetBilledMinutes(),
			Status:          pb.TripStatus(response.GetTrip().GetStatus()),
			RentalVersion:   response.GetTrip().GetRentalVersion(),
			Currency:        response.GetTrip().GetCurrency(),
		},
	}, nil
}
