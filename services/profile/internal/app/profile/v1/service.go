package v1

import (
	"profile/internal/application/usecase"
	pb "profile/internal/pkg/pb/profile/profile/v1"
)

type Implementation struct {
	pb.UnimplementedProfileServiceServer
	usecases *usecase.Registry
}

func New(usecases *usecase.Registry) *Implementation {
	return &Implementation{usecases: usecases}
}
