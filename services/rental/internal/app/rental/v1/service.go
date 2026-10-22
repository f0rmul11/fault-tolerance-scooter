package v1

import (
	"rental/internal/application/usecase"
	pb "rental/internal/pkg/pb/rental/rental/v1"
)

type Implementation struct {
	pb.UnimplementedRentalServiceServer
	usecases *usecase.Registry
}

func New(usecases *usecase.Registry) *Implementation {
	return &Implementation{usecases: usecases}
}
