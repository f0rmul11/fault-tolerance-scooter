//go:build integration

package reserve

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	scooterdto "rental/internal/dto/scooter"
	rentalstorage "rental/internal/infrastructure/storage/rental"
	"rental/internal/pkg/event"
	"rental/internal/pkg/outbox"
	app "rental/internal/pkg/rentalerror"
	"rental/internal/pkg/transaction/wrapper"
)

type abortFlusher struct {
	outbox event.Flusher
	err    error
}

func (f abortFlusher) Flush(ctx context.Context, events event.Events) error {
	if err := f.outbox.Flush(ctx, events); err != nil {
		return err
	}
	return f.err
}

// Запускать только с отдельно мигрированной тестовой БД rental.
func TestPostgresAtomicOutboxAndCAS(t *testing.T) {
	dsn := os.Getenv("SCOOTER_RENTAL_TEST_DSN")
	if dsn == "" {
		t.Skip("set SCOOTER_RENTAL_TEST_DSN to an isolated migrated rental database")
	}
	pool, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	storage := rentalstorage.New(wrapper.NewDatabase(pool))
	flusher := outbox.NewOutbox(pool, outbox.Config{
		LockedKeysLimit:       10,
		MessagesLimit:         10,
		MaxErrCountForMessage: 5,
	})
	input := validInput()
	gateway := &scooterFake{value: scooterdto.Scooter{
		ID:            input.ScooterID,
		Number:        "TEST-ATOMIC",
		Brand:         "Ninebot",
		ChargePercent: 90,
	}}
	failure := errors.New("abort after both inserts")
	result, err := New(storage, gateway, abortFlusher{
		outbox: flusher,
		err:    failure,
	}).Execute(t.Context(), input)
	if !errors.Is(err, failure) || result != nil {
		t.Fatalf("unexpected result: %v", err)
	}
	var count int
	if err := pool.QueryRow(t.Context(), "select count(*) from public.rentals where user_id = $1", input.UserID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rental survived rollback: count=%d, err=%v", count, err)
	}
	if err := pool.QueryRow(t.Context(), "select count(*) from public.outbox where convert_from(body, 'UTF8')::jsonb -> 'payload' ->> 'user_id' = $1", input.UserID.String()).Scan(&count); err != nil || count != 0 {
		t.Fatalf("outbox survived rollback: count=%d, err=%v", count, err)
	}
	result, err = New(storage, gateway, flusher).Execute(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(context.Background(), "delete from public.outbox where entity_id = $1", result.ID().String())
		_, _ = pool.Exec(context.Background(), "delete from public.rentals where id = $1", result.ID())
	}()
	if err := pool.QueryRow(t.Context(), "select count(*) from public.outbox where entity_id = $1", result.ID().String()).Scan(&count); err != nil || count != 1 {
		t.Fatalf("committed state missing event: count=%d, err=%v", count, err)
	}
	first, err := storage.Get(t.Context(), result.ID())
	if err != nil {
		t.Fatal(err)
	}
	stale, err := storage.Get(t.Context(), result.ID())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.Start(input.RequestID, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := stale.Cancel(input.RequestID, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := storage.Save(t.Context(), first); err != nil {
		t.Fatal(err)
	}
	if err := storage.Save(t.Context(), stale); !app.IsCode(err, app.Conflict) {
		t.Fatalf("stale CAS succeeded: %v", err)
	}
}
