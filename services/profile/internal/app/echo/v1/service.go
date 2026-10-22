package echo

import echoV1 "profile/internal/pkg/pb/profile/echo/v1"

type Implementation struct {
	echoV1.UnimplementedEchoServiceServer
}

func NewEchoService() *Implementation {
	return &Implementation{}
}
