package list

import (
	"context"
	"rental/internal/domain/entity"
	app "rental/internal/pkg/rentalerror"

	"github.com/gofrs/uuid"
)

type Storage interface {
	List(context.Context, uuid.UUID, int, int) ([]*entity.Rental, error)
}
type UseCase struct {
	storage Storage
}

func New(storage Storage) *UseCase {
	return &UseCase{storage: storage}
}
func (s *UseCase) Execute(ctx context.Context, input Input) ([]*entity.Rental, error) {
	if input.UserID == uuid.Nil {
		return nil, app.NewInvalidArgumentErr("user_id должен быть UUID")
	}
	if input.Limit == 0 {
		input.Limit = 20
	}
	if input.Limit < 1 || input.Limit > 100 || input.Offset < 0 {
		return nil, app.NewInvalidArgumentErr("Некорректные параметры страницы")
	}
	return s.storage.List(ctx, input.UserID, input.Limit, input.Offset)
}
