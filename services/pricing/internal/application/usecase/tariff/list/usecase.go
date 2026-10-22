package list

import (
	"context"
	"pricing/internal/dto/tariff"
)

type Storage interface {
	List(context.Context) ([]tariff.Tariff, error)
}
type UseCase struct {
	storage Storage
}

func New(storage Storage) *UseCase {
	return &UseCase{storage: storage}
}
func (s *UseCase) Execute(ctx context.Context) ([]tariff.Tariff, error) {
	return s.storage.List(ctx)
}
