package pricing

import (
	"context"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"profile/internal/pkg/grpc/intercept"
	pb "profile/internal/pkg/pb/external/pricing/pricing/v1"
)

type ExternalClient pb.PricingServiceClient

func NewExternalClient(conn grpc.ClientConnInterface) ExternalClient {
	return pb.NewPricingServiceClient(conn)
}

type Client struct {
	external ExternalClient
}

func NewClient(external ExternalClient) *Client {
	return &Client{external: external}
}

func (c *Client) CompletedTripsCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var trailers metadata.MD
	response, err := c.external.GetCompletedTripsCount(ctx, &pb.GetCompletedTripsCountRequest{
		UserId: userID.String(),
	}, grpc.Trailer(&trailers))
	if err != nil {
		return 0, intercept.FromRPC(err, trailers)
	}
	return response.GetCount(), nil
}
