package echo

import (
	"context"

	echoV1 "rental/internal/pkg/pb/rental/echo/v1"
)

func (i *Implementation) Echo(_ context.Context, req *echoV1.EchoRequest) (*echoV1.EchoResponse, error) {
	return &echoV1.EchoResponse{
		Message: req.GetMessage(),
		Service: "rental",
	}, nil
}
