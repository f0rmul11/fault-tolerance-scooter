package tariff

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"pricing/internal/dto/tariff"
	"pricing/internal/infrastructure/storage/tariff/dao"
)

type Storage struct {
	mtx     sync.RWMutex
	tariffs map[string]dao.Tariff
}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) IsLoaded() bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return len(s.tariffs) != 0
}

func (s *Storage) LoadTariffs(_ context.Context, filePath string) error {
	// Имитация тяжёлой загрузки.
	time.Sleep(30 * time.Second)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var result struct {
		Tariffs []dao.Tariff `json:"tariffs"`
	}
	if err = json.Unmarshal(data, &result); err != nil {
		return err
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.tariffs = make(map[string]dao.Tariff, len(result.Tariffs))
	for _, item := range result.Tariffs {
		s.tariffs[item.Plan] = item
	}
	return nil
}

func (s *Storage) Get(_ context.Context, plan string) (tariff.Tariff, error) {
	s.mtx.RLock()
	item := s.tariffs[plan]
	s.mtx.RUnlock()

	return tariff.Tariff{
		Plan:         item.Plan,
		RateMinor:    item.RateMinor,
		AdvanceMinor: item.AdvanceMinor,
	}, nil
}

func (s *Storage) List(_ context.Context) ([]tariff.Tariff, error) {
	var tariffs []tariff.Tariff

	s.mtx.RLock()
	for _, item := range s.tariffs {
		tariffs = append(tariffs, tariff.Tariff{
			Plan:         item.Plan,
			RateMinor:    item.RateMinor,
			AdvanceMinor: item.AdvanceMinor,
		})
	}
	s.mtx.RUnlock()

	return tariffs, nil
}
