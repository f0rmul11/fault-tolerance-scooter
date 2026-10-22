package v1

import (
	"pricing/internal/application/usecase"
	pb "pricing/internal/pkg/pb/pricing/pricing/v1"
)

type Implementation struct {
	pb.UnimplementedPricingServiceServer
	usecases *usecase.Registry
}

func New(usecases *usecase.Registry) *Implementation {
	return &Implementation{usecases: usecases}
}
