package profile

import (
	"context"
	"pricing/internal/pkg/grpc/intercept"
	pb "pricing/internal/pkg/pb/external/profile/profile/v1"
	"strings"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ExternalClient pb.ProfileServiceClient

func NewExternalClient(conn grpc.ClientConnInterface) ExternalClient {
	return pb.NewProfileServiceClient(conn)
}

type Client struct {
	external ExternalClient
}

func NewClient(external ExternalClient) *Client {
	return &Client{external: external}
}
func (c *Client) Plan(ctx context.Context, id uuid.UUID) (string, error) {
	var trailers metadata.MD
	response, err := c.external.GetProfile(ctx, &pb.GetProfileRequest{
		UserId: id.String(),
	}, grpc.Trailer(&trailers))
	if err != nil {
		return "", intercept.FromRPC(err, trailers)
	}
	return strings.TrimPrefix(response.GetProfile().GetPlan().String(), "SUBSCRIPTION_PLAN_"), nil
}
