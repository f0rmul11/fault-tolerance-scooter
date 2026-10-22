package update

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/samber/mo"

	updateservice "profile/internal/application/service/profile/update"
	"profile/internal/domain/entity"
	app "profile/internal/pkg/profileerror"
	"profile/internal/pkg/transaction/abstract"
)

type storageFake struct {
	subscriptionCalls   int
	subscriptionContext context.Context
	subscriptionErr     error
	loaded              *entity.User
	calls               int
	saved               *entity.User
	err                 error
	ctx                 context.Context
}

func (s *storageFake) Get(context.Context, uuid.UUID) (*entity.User, error) {
	return s.loaded, nil
}

func (s *storageFake) Save(ctx context.Context, objects ...*entity.User) error {
	s.calls++
	s.saved = objects[0]
	s.ctx = ctx
	return s.err
}

func TestFailedUpdateDoesNotMutateLoadedSnapshot(t *testing.T) {
	for _, stage := range []string{
		"validation",
		"storage",
	} {
		t.Run(stage, func(t *testing.T) {
			original := entity.With(
				uuid.Must(uuid.NewV4()),
				entity.WithName("Old"),
				entity.WithEmail("old@example.test"),
				entity.WithPlan(entity.Base),
				entity.WithVersion(2),
			)
			storage := &storageFake{loaded: original}
			changes := Input{
				ID:              original.ID(),
				ExpectedVersion: 2,
				Name:            mo.Some("New"),
				Plan:            mo.Some(entity.Plan("INVALID")),
			}
			if stage == "storage" {
				changes = Input{
					ID:              original.ID(),
					ExpectedVersion: 2,
					Name:            mo.Some("New"),
					Plan:            mo.Some(entity.Premium),
				}
				storage.err = errors.New("write failed")
			}
			result, err := New(storage, updateservice.New(storage)).Execute(t.Context(), changes)
			if err == nil || result != nil {
				t.Fatal("failed update reported success")
			}
			if original.Name() != "Old" || original.Plan() != entity.Base || original.Version() != 2 {
				t.Fatal("failed update partially changed loaded aggregate")
			}
			if stage == "validation" && storage.calls != 0 {
				t.Fatal("validation failure reached storage")
			}
			if stage == "storage" && storage.subscriptionCalls != 0 {
				t.Fatal("subscription written after failed profile CAS")
			}
			if stage == "storage" && !errors.Is(err, storage.err) {
				t.Fatalf("storage error lost: %v", err)
			}
		})
	}
}

func TestPatchPreservesOmittedFieldsAndAdvancesVersionOnce(t *testing.T) {
	original := entity.With(
		uuid.Must(uuid.NewV4()),
		entity.WithName("Old"),
		entity.WithEmail("old@example.test"),
		entity.WithPlan(entity.Base),
		entity.WithVersion(2),
	)
	storage := &storageFake{loaded: original}
	result, err := New(storage, updateservice.New(storage)).Execute(t.Context(), Input{
		ID:              original.ID(),
		ExpectedVersion: 2,
		Name:            mo.Some("New"),
		Plan:            mo.Some(entity.Premium),
	})
	if err != nil {
		t.Fatal(err)
	}
	if storage.calls != 1 || result != storage.saved || result.Version() != 3 || result.OldVersion() != 2 || result.Name() != "New" || result.Plan() != entity.Premium || result.Email() != "old@example.test" {
		t.Fatal("patch changed omitted fields or incremented version per field")
	}
	if abstract.TxFromContext(storage.ctx) == nil {
		t.Fatal("profile and subscription must share a transaction")
	}
}

func (s *storageFake) SaveSubscriptions(ctx context.Context, objects ...*entity.User) error {
	s.subscriptionCalls++
	s.subscriptionContext = ctx
	return s.subscriptionErr
}

func TestSubscriptionFailureIsReturnedFromSharedTransaction(t *testing.T) {
	original := entity.With(uuid.Must(uuid.NewV4()), entity.WithVersion(2))
	failure := errors.New("subscription write failed")
	storage := &storageFake{
		loaded:          original,
		subscriptionErr: failure,
	}
	result, err := New(storage, updateservice.New(storage)).Execute(t.Context(), Input{
		ID:              original.ID(),
		ExpectedVersion: 2,
		Plan:            mo.Some(entity.Ultra),
	})
	if !errors.Is(err, failure) || result != nil {
		t.Fatalf("subscription failure reported success: %v", err)
	}
	if storage.calls != 1 || storage.subscriptionCalls != 1 || storage.ctx != storage.subscriptionContext || abstract.TxFromContext(storage.ctx) == nil {
		t.Fatal("profile and subscription writes must share the service transaction")
	}
}

func TestExpectedVersionIsCheckedByStorageForEveryPatch(t *testing.T) {
	for _, expected := range []int{1, 2, 3} {
		for _, value := range []mo.Option[string]{mo.None[string](), mo.Some("Original"), mo.Some("Changed")} {
			original := entity.With(
				uuid.Must(uuid.NewV4()),
				entity.WithName("Original"),
				entity.WithVersion(2),
			)
			conflict := app.NewConflictErr("Данные изменились")
			storage := &storageFake{
				loaded: original,
				err:    conflict,
			}
			result, err := New(storage, updateservice.New(storage)).Execute(t.Context(), Input{
				ID:              original.ID(),
				ExpectedVersion: expected,
				Name:            value,
			})
			if !errors.Is(err, conflict) || result != nil || storage.calls != 1 {
				t.Fatalf("version %d was checked outside storage or skipped: %v", expected, err)
			}
			if storage.saved.OldVersion() != 2 || storage.saved.ExpectedVersion() != expected {
				t.Fatal("client condition overwrote the loaded version")
			}
			wantVersion := 2
			if value.OrEmpty() == "Changed" {
				wantVersion = 3
			}
			if storage.saved.Version() != wantVersion || original.Version() != 2 {
				t.Fatal("patch/no-op changed version incorrectly")
			}
		}
	}
}
