package discount

import (
	"context"

	"github.com/gofrs/uuid"

	"profile/internal/domain/entity"
	"profile/internal/domain/loyalty"
	dto "profile/internal/dto/discount"
	app "profile/internal/pkg/profileerror"
)

type Storage interface {
	Get(context.Context, uuid.UUID) (*entity.User, error)
}

type Pricing interface {
	CompletedTripsCount(context.Context, uuid.UUID) (int64, error)
}

type UseCase struct {
	storage Storage
	pricing Pricing
}

func New(storage Storage, pricing Pricing) *UseCase {
	return &UseCase{
		storage: storage,
		pricing: pricing,
	}
}

func (s *UseCase) Execute(ctx context.Context, userID uuid.UUID) (*dto.Discount, error) {
	if userID == uuid.Nil {
		return nil, app.NewInvalidArgumentErr("Идентификатор должен быть UUID")
	}
	if _, err := s.storage.Get(ctx, userID); err != nil {
		return nil, err
	}
	count, err := s.pricing.CompletedTripsCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &dto.Discount{
		Percent:             loyalty.DiscountPercent(count),
		CompletedTripsCount: &count,
		Source:              "pricing",
	}, nil
}
