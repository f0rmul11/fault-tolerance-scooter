package echo

import echoV1 "scooter/internal/pkg/pb/scooter/echo/v1"

type Implementation struct {
	echoV1.UnimplementedEchoServiceServer
}

func NewEchoService() *Implementation {
	return &Implementation{}
}
