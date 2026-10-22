package scooter

import (
	"context"
	scooterdto "rental/internal/dto/scooter"
	"rental/internal/pkg/grpc/intercept"
	pb "rental/internal/pkg/pb/external/scooter/scooter/v1"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ExternalClient pb.ScooterServiceClient

func NewExternalClient(conn grpc.ClientConnInterface) ExternalClient {
	return pb.NewScooterServiceClient(conn)
}

type Client struct {
	external ExternalClient
}

func NewClient(external ExternalClient) *Client {
	return &Client{external: external}
}
func (c *Client) Get(ctx context.Context, id uuid.UUID) (scooterdto.Scooter, error) {
	var trailers metadata.MD
	response, err := c.external.GetScooter(ctx, &pb.GetScooterRequest{
		ScooterId: id.String(),
	}, grpc.Trailer(&trailers))
	if err != nil {
		return scooterdto.Scooter{}, intercept.FromRPC(err, trailers)
	}
	return scooterdto.Scooter{
		ID:            uuid.FromStringOrNil(response.GetScooter().GetId()),
		Number:        response.GetScooter().GetNumber(),
		Brand:         response.GetScooter().GetBrand(),
		ChargePercent: int(response.GetScooter().GetChargePercent()),
	}, nil
}
