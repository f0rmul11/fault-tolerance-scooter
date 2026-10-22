package echo

import echoV1 "rental/internal/pkg/pb/rental/echo/v1"

type Implementation struct {
	echoV1.UnimplementedEchoServiceServer
}

func NewEchoService() *Implementation {
	return &Implementation{}
}
