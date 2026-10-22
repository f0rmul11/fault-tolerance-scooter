package reserve

import (
	"context"
	"errors"
	"pricing/internal/domain/entity"
	"pricing/internal/dto/tariff"
	app "pricing/internal/pkg/pricingerror"
	"testing"
	"time"

	"github.com/gofrs/uuid"
)

type storeFake struct {
	saved   *entity.Trip
	created *entity.Trip
}

func (s *storeFake) Get(context.Context, uuid.UUID) (*entity.Trip, error) {
	if s.saved == nil {
		return nil, app.NewNotFoundErr("Нет расчёта", nil)
	}
	return s.saved, nil
}
func (s *storeFake) Create(_ context.Context, trip *entity.Trip) error {
	s.created = trip
	return nil
}

type profileFake struct {
	calls int
	plan  string
	err   error
}

func (p *profileFake) Plan(context.Context, uuid.UUID) (string, error) {
	p.calls++
	return p.plan, p.err
}

type tariffFake struct{}

func (tariffFake) Get(_ context.Context, plan string) (tariff.Tariff, error) {
	return tariff.Tariff{
		Plan:         plan,
		RateMinor:    700,
		AdvanceMinor: 3000,
	}, nil
}
func TestSnapshotAndReplay(t *testing.T) {
	input := Input{
		RentalID:      uuid.Must(uuid.NewV4()),
		UserID:        uuid.Must(uuid.NewV4()),
		ScooterID:     uuid.Must(uuid.NewV4()),
		RentalVersion: 1,
		ReservedAt:    time.Now(),
	}
	store := &storeFake{}
	profile := &profileFake{plan: "PREMIUM"}
	usecase := New(store, profile, tariffFake{})
	if err := usecase.Execute(t.Context(), input); err != nil {
		t.Fatal(err)
	}
	if store.created.Plan() != "PREMIUM" || store.created.RateMinor() != 700 || store.created.ChargedMinor() != 0 {
		t.Fatal("wrong tariff snapshot or premature charge")
	}
	store.saved = store.created
	profile.plan = "ULTRA"
	profile.err = errors.New("profile offline")
	if err := usecase.Execute(t.Context(), input); err != nil || profile.calls != 1 || store.saved.Plan() != "PREMIUM" {
		t.Fatalf("replay replaced snapshot: %v", err)
	}
}
func TestProfileFailureDoesNotFallbackToBase(t *testing.T) {
	input := Input{
		RentalID:      uuid.Must(uuid.NewV4()),
		UserID:        uuid.Must(uuid.NewV4()),
		ScooterID:     uuid.Must(uuid.NewV4()),
		RentalVersion: 1,
		ReservedAt:    time.Now(),
	}
	failure := errors.New("profile offline")
	store := &storeFake{}
	err := New(store, &profileFake{err: failure}, tariffFake{}).Execute(t.Context(), input)
	if !errors.Is(err, failure) || store.created != nil {
		t.Fatal("failed lookup persisted an invented price")
	}
}
