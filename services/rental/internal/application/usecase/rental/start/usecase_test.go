package start

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	transition "rental/internal/application/service/rental/start"
	"rental/internal/domain/entity"
	"rental/internal/pkg/event"
	app "rental/internal/pkg/rentalerror"
	"rental/internal/pkg/transaction/abstract"
)

type storageFake struct {
	loaded *entity.Rental
	saved  *entity.Rental
	ctx    context.Context
	err    error
}

func (s *storageFake) Get(context.Context, uuid.UUID) (*entity.Rental, error) {
	return s.loaded, nil
}

func (s *storageFake) Save(ctx context.Context, objects ...*entity.Rental) error {
	s.saved = objects[0]
	s.ctx = ctx
	return s.err
}

type flusherFake struct {
	batch event.Events
	ctx   context.Context
	err   error
}

func (f *flusherFake) Flush(ctx context.Context, events event.Events) error {
	f.batch = events
	f.ctx = ctx
	return f.err
}

func TestTransitionOwnsDomainChangeAndPersistsItsEvent(t *testing.T) {
	for _, scenario := range []string{"success", "repeat", "invalid state", "wrong owner", "save failure", "outbox failure"} {
		t.Run(scenario, func(t *testing.T) {
			owner := uuid.Must(uuid.NewV4())
			status := entity.Reserved
			if scenario == "repeat" {
				status = entity.Active
			}
			if scenario == "invalid state" {
				status = entity.Canceled
			}
			original := entity.With(
				uuid.Must(uuid.NewV4()),
				entity.WithUserID(owner),
				entity.WithScooterID(uuid.Must(uuid.NewV4())),
				entity.WithStatus(status),
				entity.WithVersion(2),
				entity.WithReservedAt(time.Now().Add(-time.Minute)),
				entity.WithStartedAt(time.Now().Add(-30*time.Second)),
				entity.WithUpdatedAt(time.Now().Add(-30*time.Second)),
			)
			store := &storageFake{loaded: original}
			flusher := &flusherFake{}
			failure := errors.New("persistence failed")
			input := Input{
				RentalID:  original.ID(),
				UserID:    owner,
				RequestID: uuid.Must(uuid.NewV4()),
			}
			if scenario == "wrong owner" {
				input.UserID = uuid.Must(uuid.NewV4())
			}
			if scenario == "save failure" {
				store.err = failure
			}
			if scenario == "outbox failure" {
				flusher.err = failure
			}
			result, err := New(store, transition.New(store), flusher).Execute(t.Context(), input)
			if original.Status() != status || original.Version() != 2 {
				t.Fatal("loaded snapshot was mutated")
			}
			switch scenario {
			case "success":
				if err != nil || result != store.saved || result.Status() != entity.Active || result.Version() != 3 || len(flusher.batch) != 1 {
					t.Fatalf("transition or event lost: %v", err)
				}
				if abstract.TxFromContext(store.ctx) == nil || store.ctx != flusher.ctx {
					t.Fatal("state and event do not share transaction context")
				}
				var envelope struct {
					Version   int    `json:"rental_version"`
					RequestID string `json:"request_id"`
				}
				if err := json.Unmarshal(flusher.batch[0].Body, &envelope); err != nil {
					t.Fatal(err)
				}
				if envelope.Version != 3 || envelope.RequestID != input.RequestID.String() || flusher.batch[0].EntityID != original.ID().String() {
					t.Fatal("event differs from the saved domain transition")
				}
			case "repeat":
				if err != nil || result != original || store.saved != nil || len(flusher.batch) != 0 {
					t.Fatal("repeat saved state or emitted another event")
				}
			case "invalid state", "wrong owner":
				if err == nil || result != nil || store.saved != nil || len(flusher.batch) != 0 {
					t.Fatal("rejected command reached persistence")
				}
				if scenario == "wrong owner" && !app.IsCode(err, app.PermissionDenied) {
					t.Fatalf("wrong error for ownership: %v", err)
				}
			case "save failure", "outbox failure":
				if !errors.Is(err, failure) || result != nil {
					t.Fatalf("persistence failure reported success: %v", err)
				}
				if scenario == "save failure" && len(flusher.batch) != 0 {
					t.Fatal("outbox called after failed CAS")
				}
			}
		})
	}
}
