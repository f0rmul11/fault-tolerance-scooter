package v1

import (
	"scooter/internal/application/usecase"
	pb "scooter/internal/pkg/pb/scooter/scooter/v1"
)

type Implementation struct {
	pb.UnimplementedScooterServiceServer
	usecases *usecase.Registry
}

func New(usecases *usecase.Registry) *Implementation {
	return &Implementation{usecases: usecases}
}
