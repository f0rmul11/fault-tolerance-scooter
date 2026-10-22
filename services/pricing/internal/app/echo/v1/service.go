package echo

import echoV1 "pricing/internal/pkg/pb/pricing/echo/v1"

type Implementation struct {
	echoV1.UnimplementedEchoServiceServer
}

func NewEchoService() *Implementation {
	return &Implementation{}
}
