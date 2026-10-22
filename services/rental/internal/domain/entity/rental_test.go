package entity

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"rental/internal/pkg/event"
)

func TestTransitions(t *testing.T) {
	at := time.Date(
		2026,
		10,
		22,
		10,
		0,
		0,
		0,
		time.UTC,
	)
	for _, test := range []struct {
		name     string
		initial  Status
		action   string
		expected Status
		changed  bool
		rejected bool
	}{
		{
			"start",
			Reserved,
			"start",
			Active,
			true,
			false,
		},
		{
			"repeat start",
			Active,
			"start",
			Active,
			false,
			false,
		},
		{
			"late start",
			Completed,
			"start",
			Completed,
			false,
			true,
		},
		{
			"start canceled",
			Canceled,
			"start",
			Canceled,
			false,
			true,
		},
		{
			"finish",
			Active,
			"finish",
			Completed,
			true,
			false,
		},
		{
			"repeat finish",
			Completed,
			"finish",
			Completed,
			false,
			false,
		},
		{
			"finish before start",
			Reserved,
			"finish",
			Reserved,
			false,
			true,
		},
		{
			"cancel",
			Reserved,
			"cancel",
			Canceled,
			true,
			false,
		},
		{
			"repeat cancel",
			Canceled,
			"cancel",
			Canceled,
			false,
			false,
		},
		{
			"cancel active",
			Active,
			"cancel",
			Active,
			false,
			true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			rental := With(
				uuid.Must(uuid.NewV4()),
				WithStatus(test.initial),
				WithVersion(3),
				WithReservedAt(at),
				WithUpdatedAt(at),
			)
			action := rental.Start
			if test.action == "finish" {
				action = rental.Finish
			}
			if test.action == "cancel" {
				action = rental.Cancel
			}
			flusher := &captureFlusher{}
			buffer, ctx := event.WithContext(t.Context(), flusher)
			changed, err := action(ctx, uuid.Must(uuid.NewV4()), at.Add(time.Minute))
			if (err != nil) != test.rejected || changed != test.changed || rental.Status() != test.expected {
				t.Fatalf("changed=%t error=%v status=%s", changed, err, rental.Status())
			}
			if err := buffer.Flush(ctx); err != nil {
				t.Fatal(err)
			}
			if (len(flusher.batch) == 1) != test.changed {
				t.Fatal("event must be emitted only for a successful transition")
			}
			expectedVersion := 3
			if test.changed {
				expectedVersion++
			}
			if rental.Version() != expectedVersion || rental.OldVersion() != 3 {
				t.Fatal("unexpected version transition")
			}
		})
	}
}

func TestReservationAvailabilityAndIdentity(t *testing.T) {
	for _, isActive := range []bool{false, true} {
		_, ctx := event.WithContext(t.Context(), &captureFlusher{})
		r, err := NewRental(
			ctx,
			uuid.Must(uuid.NewV4()),
			uuid.Must(uuid.NewV4()),
			uuid.Must(uuid.NewV4()),
			uuid.Must(uuid.NewV4()),
			uuid.Must(uuid.NewV4()),
			"SC-1",
			"Ninebot",
			isActive,
			time.Now(),
		)
		valid := isActive
		if (err == nil) != valid {
			t.Fatalf("isActive=%t: %v", isActive, err)
		}
		if valid && (r.Version() != 1 || r.Status() != Reserved || r.CheckOwner(uuid.Must(uuid.NewV4())) == nil) {
			t.Fatal("reservation identity or state is invalid")
		}
	}
}

func TestClockRegressionDoesNotChangeRental(t *testing.T) {
	at := time.Now()
	r := With(uuid.Must(uuid.NewV4()), WithStatus(Active), WithUpdatedAt(at), WithVersion(2))
	_, ctx := event.WithContext(t.Context(), &captureFlusher{})
	changed, err := r.Finish(ctx, uuid.Must(uuid.NewV4()), at.Add(-time.Second))
	if err == nil || changed || r.Status() != Active || r.Version() != 2 {
		t.Fatal("regressed clock changed rental")
	}
}

type captureFlusher struct {
	batch event.Events
	err   error
}

func (f *captureFlusher) Flush(_ context.Context, batch event.Events) error {
	f.batch = batch
	return f.err
}

func TestDomainEventsKeepTransitionSnapshotAndIdentity(t *testing.T) {
	flusher := &captureFlusher{err: errors.New("outbox unavailable")}
	buffer, ctx := event.WithContext(t.Context(), flusher)
	at := time.Date(
		2026,
		10,
		22,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	rental, err := NewRental(
		ctx,
		uuid.Must(uuid.NewV4()),
		uuid.Must(uuid.NewV4()),
		uuid.Must(uuid.NewV4()),
		uuid.Must(uuid.NewV4()),
		uuid.Must(uuid.NewV4()),
		"SC-1",
		"Ninebot",
		true,
		at,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rental.Start(ctx, uuid.Must(uuid.NewV4()), at.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err := rental.Finish(ctx, uuid.Must(uuid.NewV4()), at.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}

	// Все переходы выполнены до flush: каждое событие обязано сохранить свой снимок.
	if err := buffer.Flush(ctx); !errors.Is(err, flusher.err) {
		t.Fatal("failed flush must retain pending events")
	}
	first := append(event.Events(nil), flusher.batch...)
	if len(first) != 3 {
		t.Fatalf("expected three domain events, got %d", len(first))
	}
	flusher.err = nil
	if err := buffer.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	states := []Status{
		Reserved,
		Active,
		Completed,
	}
	for i, msg := range flusher.batch {
		var envelope struct {
			CorrelationID string `json:"correlation_id"`
			RentalVersion int    `json:"rental_version"`
			Payload       struct {
				Status Status `json:"status"`
			} `json:"payload"`
		}
		if err := json.Unmarshal(msg.Body, &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Payload.Status != states[i] || envelope.RentalVersion != i+1 || envelope.CorrelationID == "" {
			t.Fatalf("event snapshot changed: %+v", envelope)
		}
		if string(first[i].Body) != string(msg.Body) || string(first[i].Headers) != string(msg.Headers) {
			t.Fatal("flush retry changed event identity")
		}
	}
}
