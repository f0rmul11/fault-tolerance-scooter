package entity

import (
	"math"
	"testing"
	"time"

	"github.com/gofrs/uuid"
)

func TestBillingRoundsStartedMinutesAndSettlesAdvance(t *testing.T) {
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
		duration time.Duration
		minutes  int64
	}{
		{
			"immediate finish",
			0,
			1,
		},
		{
			"fraction",
			time.Microsecond,
			1,
		},
		{
			"exact minute",
			time.Minute,
			1,
		},
		{
			"next minute",
			time.Minute + time.Microsecond,
			2,
		},
		{
			"additional charge",
			10*time.Minute + time.Second,
			11,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			trip := With(
				uuid.Must(uuid.NewV4()),
				WithStatus(Active),
				WithStartedAt(at),
				WithRateMinor(700),
				WithAdvanceMinor(3000),
				WithChargedMinor(3000),
				WithRentalVersion(2),
				WithVersion(2),
			)
			if err := trip.Complete(3, at, at.Add(test.duration)); err != nil {
				t.Fatal(err)
			}
			total := test.minutes * 700
			if trip.TotalMinor() != total || trip.ChargedMinor() != total || trip.AdjustmentMinor() != total-3000 || trip.BilledMinutes() != test.minutes || trip.Version() != 3 {
				t.Fatalf("incorrect settlement: %+v", trip)
			}
		})
	}
}

func TestEventOrderAndDuplicates(t *testing.T) {
	userID, scooterID := uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4())
	trip := With(
		uuid.Must(uuid.NewV4()),
		WithUserID(userID),
		WithScooterID(scooterID),
		WithRentalVersion(1),
		WithVersion(1),
		WithStatus(Reserved),
		WithAdvanceMinor(5000),
		WithReservedAt(time.Now()),
	)
	apply, err := trip.CheckEvent(userID, scooterID, 3)
	if err == nil || apply {
		t.Fatal("out-of-order completion was accepted")
	}
	apply, err = trip.CheckEvent(userID, scooterID, 1)
	if err != nil || apply {
		t.Fatal("duplicate reservation must be a no-op")
	}
	if err := trip.Start(2, time.Now()); err != nil || trip.ChargedMinor() != 5000 {
		t.Fatalf("advance not charged: %v", err)
	}
	apply, err = trip.CheckEvent(userID, scooterID, 2)
	if err != nil || apply || trip.ChargedMinor() != 5000 {
		t.Fatal("duplicate start charged advance again")
	}
}

func TestInvalidBillingDoesNotMutateTrip(t *testing.T) {
	at := time.Now()
	for _, test := range []struct {
		name string
		rate int64
		end  time.Time
	}{
		{
			"negative duration",
			1000,
			at.Add(-time.Second),
		},
		{
			"overflow",
			math.MaxInt64,
			at.Add(2 * time.Minute),
		},
		{
			"duration overflow",
			1000,
			at.AddDate(500, 0, 0),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			trip := With(
				uuid.Must(uuid.NewV4()),
				WithStatus(Active),
				WithStartedAt(at),
				WithRateMinor(test.rate),
				WithAdvanceMinor(5000),
				WithChargedMinor(5000),
				WithRentalVersion(2),
				WithVersion(2),
			)
			if err := trip.Complete(3, at, test.end); err == nil {
				t.Fatal("invalid settlement accepted")
			}
			if trip.Status() != Active || trip.ChargedMinor() != 5000 || trip.Version() != 2 {
				t.Fatal("rejected settlement mutated trip")
			}
		})
	}
}

func TestCancelReservationDoesNotCharge(t *testing.T) {
	trip := With(
		uuid.Must(uuid.NewV4()),
		WithStatus(Reserved),
		WithAdvanceMinor(5000),
		WithRentalVersion(1),
		WithVersion(1),
	)
	if err := trip.Cancel(2); err != nil || trip.ChargedMinor() != 0 || trip.Status() != Canceled {
		t.Fatalf("cancel: %v", err)
	}
}
