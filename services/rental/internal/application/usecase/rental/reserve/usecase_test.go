package reserve

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"

	reserveservice "rental/internal/application/service/rental/reserve"
	"rental/internal/domain/entity"
	scooterdto "rental/internal/dto/scooter"
	"rental/internal/pkg/event"
	app "rental/internal/pkg/rentalerror"
	"rental/internal/pkg/transaction/abstract"
)

type storeFake struct {
	saved     *entity.Rental
	created   *entity.Rental
	err       error
	txContext context.Context
}

func (s *storeFake) FindKey(context.Context, uuid.UUID, uuid.UUID) (*entity.Rental, error) {
	return s.saved, nil
}

func (s *storeFake) Create(ctx context.Context, rental *entity.Rental) error {
	s.txContext = ctx
	s.created = rental
	return s.err
}

type scooterFake struct {
	value scooterdto.Scooter
	calls int
	err   error
}

func (s *scooterFake) Get(context.Context, uuid.UUID) (scooterdto.Scooter, error) {
	s.calls++
	return s.value, s.err
}

type flusherFake struct {
	calls     int
	txContext context.Context
	batch     event.Events
	err       error
}

func (f *flusherFake) Flush(ctx context.Context, events event.Events) error {
	f.calls++
	f.txContext = ctx
	f.batch = events
	return f.err
}

func validInput() Input {
	return Input{
		UserID:         uuid.Must(uuid.NewV4()),
		ScooterID:      uuid.Must(uuid.NewV4()),
		IdempotencyKey: uuid.Must(uuid.NewV4()),
		RequestID:      uuid.Must(uuid.NewV4()),
	}
}

func TestReplayDoesNotDependOnCatalog(t *testing.T) {
	input := validInput()
	saved := entity.With(uuid.Must(uuid.NewV4()), entity.WithScooterID(input.ScooterID))
	store := &storeFake{saved: saved}
	gateway := &scooterFake{err: errors.New("catalog unavailable")}
	flusher := &flusherFake{}
	result, err := New(store, reserveservice.New(store), gateway, flusher).Execute(t.Context(), input)
	if err != nil || result != saved || gateway.calls != 0 || flusher.calls != 0 {
		t.Fatalf("replay queried catalog or wrote event: %v", err)
	}
	input.ScooterID = uuid.Must(uuid.NewV4())
	if _, err := New(store, reserveservice.New(store), gateway, flusher).Execute(t.Context(), input); !app.IsCode(err, app.Conflict) {
		t.Fatalf("different request reused key: %v", err)
	}
}

func TestReservationPersistsAndFlushesInSameTransactionContext(t *testing.T) {
	input := validInput()
	store := &storeFake{}
	gateway := &scooterFake{value: scooterdto.Scooter{
		ID:       input.ScooterID,
		Number:   "SC-1",
		Brand:    "Ninebot",
		IsActive: true,
	}}
	flusher := &flusherFake{}
	result, err := New(store, reserveservice.New(store), gateway, flusher).Execute(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}
	if result.ID() == uuid.Nil || store.created != result || len(flusher.batch) != 1 {
		t.Fatal("rental/event identity lost")
	}
	if abstract.TxFromContext(store.txContext) == nil || store.txContext != flusher.txContext {
		t.Fatal("state and outbox do not share transaction context")
	}
	if flusher.batch[0].EntityID != result.ID().String() || string(flusher.batch[0].Key) != result.ID().String() {
		t.Fatal("event does not identify saved rental")
	}
}

func TestFailedPersistenceDoesNotReportSuccess(t *testing.T) {
	for _, step := range []string{
		"catalog",
		"save",
		"outbox",
	} {
		t.Run(step, func(t *testing.T) {
			input := validInput()
			failure := errors.New("failure")
			store := &storeFake{}
			gateway := &scooterFake{value: scooterdto.Scooter{
				ID:       input.ScooterID,
				Number:   "SC-1",
				Brand:    "Ninebot",
				IsActive: true,
			}}
			flusher := &flusherFake{}
			switch step {
			case "catalog":
				gateway.err = failure
			case "save":
				store.err = failure
			case "outbox":
				flusher.err = failure
			}
			result, err := New(store, reserveservice.New(store), gateway, flusher).Execute(t.Context(), input)
			if !errors.Is(err, failure) || result != nil {
				t.Fatalf("false success: %v", err)
			}
			if step != "outbox" && flusher.calls != 0 {
				t.Fatal("outbox called after earlier failure")
			}
		})
	}
}
