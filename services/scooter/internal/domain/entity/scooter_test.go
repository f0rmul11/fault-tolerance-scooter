package entity

import (
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	app "scooter/internal/pkg/scootererror"
)

func TestAvailabilityDoesNotChangeCatalogueOrOwner(t *testing.T) {
	at := time.Now().UTC().Truncate(time.Microsecond)
	scooter := With(uuid.Must(uuid.NewV4()), WithNumber("SC-1"), WithBrand("Ninebot"), WithOwner("ООО Прокат"), WithIsActive(true), WithRiddenSeconds(7200), WithVersion(1))
	scooter.SetActive(false, at)
	if scooter.IsActive() || scooter.Version() != 2 || scooter.OldVersion() != 1 || scooter.Owner() != "ООО Прокат" || scooter.Number() != "SC-1" || scooter.Brand() != "Ninebot" || scooter.RiddenSeconds() != 7200 {
		t.Fatal("availability change affected other fields")
	}
	scooter.SetActive(false, at.Add(time.Hour))
	if !scooter.UpdatedAt().Equal(at) || scooter.Version() != 2 {
		t.Fatal("same availability changed timestamp or version")
	}
}

func TestOwnerValidationAndNormalization(t *testing.T) {
	for _, owner := range []string{"", "  ", strings.Repeat("я", 201)} {
		scooter := With(uuid.Must(uuid.NewV4()), WithOwner("ООО Прокат"), WithIsActive(false), WithVersion(1))
		if err := scooter.ChangeOwner(owner, time.Now()); !app.IsCode(err, app.InvalidArgument) {
			t.Fatalf("invalid owner accepted: %v", err)
		}
		if scooter.Owner() != "ООО Прокат" || scooter.Version() != 1 {
			t.Fatal("rejected owner mutated entity")
		}
	}
	scooter := With(uuid.Must(uuid.NewV4()), WithOwner("ООО Прокат"), WithIsActive(false), WithVersion(1))
	if err := scooter.ChangeOwner("  ООО Новый владелец  ", time.Now()); err != nil {
		t.Fatal(err)
	}
	if scooter.Owner() != "ООО Новый владелец" || scooter.IsActive() || scooter.Version() != 2 {
		t.Fatal("owner was not normalized or changed availability")
	}
	if err := scooter.ChangeOwner(strings.Repeat("я", 200), time.Now()); err != nil {
		t.Fatalf("200 Unicode characters rejected: %v", err)
	}
}
