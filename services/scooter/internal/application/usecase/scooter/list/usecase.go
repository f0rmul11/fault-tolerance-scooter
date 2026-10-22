package list

import (
	"context"

	"scooter/internal/domain/entity"
	app "scooter/internal/pkg/scootererror"
)

type Storage interface {
	List(context.Context, int, int) ([]*entity.Scooter, error)
}

type UseCase struct {
	storage Storage
}

func New(storage Storage) *UseCase {
	return &UseCase{storage: storage}
}

func (s *UseCase) Execute(ctx context.Context, input Input) ([]*entity.Scooter, error) {
	if input.Limit == 0 {
		input.Limit = 20
	}
	if input.Limit < 1 || input.Limit > 100 || input.Offset < 0 {
		return nil, app.NewInvalidArgumentErr("Некорректные параметры страницы")
	}
	return s.storage.List(ctx, input.Limit, input.Offset)
}
