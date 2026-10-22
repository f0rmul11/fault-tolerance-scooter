package entity

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
)

func TestTransitions(t *testing.T) {
	at := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
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
			rental := With(uuid.Must(uuid.NewV4()), WithStatus(test.initial), WithVersion(3), WithReservedAt(at), WithUpdatedAt(at))
			action := rental.Start
			if test.action == "finish" {
				action = rental.Finish
			}
			if test.action == "cancel" {
				action = rental.Cancel
			}
			changed, err := action(uuid.Must(uuid.NewV4()), at.Add(time.Minute))
			if (err != nil) != test.rejected || changed != test.changed || rental.Status() != test.expected {
				t.Fatalf("changed=%t error=%v status=%s", changed, err, rental.Status())
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
func TestReservationChargeBoundaryAndIdentity(t *testing.T) {
	for _, charge := range []int{
		19,
		20,
		100,
		101,
	} {
		r, err := NewRental(uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4()), "SC-1", "Ninebot", charge, time.Now())
		valid := charge >= 20 && charge <= 100
		if (err == nil) != valid {
			t.Fatalf("charge=%d: %v", charge, err)
		}
		if valid && (r.Version() != 1 || r.Status() != Reserved || r.CheckOwner(uuid.Must(uuid.NewV4())) == nil) {
			t.Fatal("reservation identity or state is invalid")
		}
	}
}
func TestClockRegressionDoesNotChangeRental(t *testing.T) {
	at := time.Now()
	r := With(uuid.Must(uuid.NewV4()), WithStatus(Active), WithUpdatedAt(at), WithVersion(2))
	changed, err := r.Finish(uuid.Must(uuid.NewV4()), at.Add(-time.Second))
	if err == nil || changed || r.Status() != Active || r.Version() != 2 {
		t.Fatal("regressed clock changed rental")
	}
}
