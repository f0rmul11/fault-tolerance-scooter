package availability

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"

	"scooter/internal/domain/entity"
	app "scooter/internal/pkg/scootererror"
)

type storageFake struct {
	loaded *entity.Scooter
	saved  *entity.Scooter
	err    error
}

func (s *storageFake) Get(context.Context, uuid.UUID) (*entity.Scooter, error) {
	return s.loaded, nil
}

func (s *storageFake) Save(_ context.Context, objects ...*entity.Scooter) error {
	s.saved = objects[0]
	return s.err
}

func TestUpdateAndVersionConflict(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		object := entity.With(uuid.Must(uuid.NewV4()), entity.WithIsActive(true), entity.WithOwner("ООО Прокат"), entity.WithVersion(2))
		storage := &storageFake{loaded: object}
		expected := 2
		if conflict {
			expected = 1
			storage.err = app.NewConflictErr("Данные изменились")
		}
		result, err := New(storage).Execute(t.Context(), Input{
			ID:              object.ID(),
			ExpectedVersion: expected,
			IsActive:        false,
		})
		if storage.saved != object || storage.saved.ExpectedVersion() != expected || storage.saved.OldVersion() != 2 {
			t.Fatal("use case copied entity or did not forward both version conditions to storage")
		}
		if conflict {
			if !errors.Is(err, storage.err) || result != nil {
				t.Fatalf("CAS failure reported success: %v", err)
			}
		} else if err != nil || result != object || result.IsActive() != false || result.Version() != 3 {
			t.Fatalf("update failed: %v", err)
		}
	}
}

func TestSameValueStillPassesSQLVersionCheck(t *testing.T) {
	object := entity.With(uuid.Must(uuid.NewV4()), entity.WithIsActive(true), entity.WithOwner("ООО Прокат"), entity.WithVersion(2))
	storage := &storageFake{
		loaded: object,
		err:    app.NewConflictErr("Данные изменились"),
	}
	result, err := New(storage).Execute(t.Context(), Input{
		ID:              object.ID(),
		ExpectedVersion: 1,
		IsActive:        true,
	})
	if !errors.Is(err, storage.err) || result != nil || storage.saved != object || object.Version() != 2 {
		t.Fatal("same value bypassed CAS or incremented version")
	}
}
