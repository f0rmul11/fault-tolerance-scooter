package entity

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/samber/mo"
)

func TestTelemetryUpdate(t *testing.T) {
	for _, test := range []struct {
		name     string
		charge   int
		ridden   int64
		rejected bool
	}{
		{
			name:   "zero charge is explicit",
			charge: 0,
			ridden: 7200,
		},
		{
			name:   "charged",
			charge: 100,
			ridden: 8000,
		},
		{
			name:     "charge above range",
			charge:   101,
			ridden:   7200,
			rejected: true,
		},
		{
			name:     "negative charge",
			charge:   -1,
			ridden:   7200,
			rejected: true,
		},
		{
			name:     "odometer regression",
			charge:   50,
			ridden:   1,
			rejected: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			scooter := With(uuid.Must(uuid.NewV4()), WithNumber("SC-1"), WithBrand("Ninebot"), WithChargePercent(90), WithRiddenSeconds(7200), WithVersion(1))
			err := scooter.Update(mo.None[string](), mo.None[string](), mo.Some(test.charge), mo.Some(test.ridden), 1, time.Now())
			if (err != nil) != test.rejected {
				t.Fatalf("error: %v", err)
			}
			if test.rejected && (scooter.ChargePercent() != 90 || scooter.Version() != 1) {
				t.Fatal("failed update mutated scooter")
			}
			if !test.rejected && (scooter.ChargePercent() != test.charge || scooter.RiddenSeconds() != test.ridden || scooter.Version() != 2) {
				t.Fatal("successful update lost fields")
			}
		})
	}
}
