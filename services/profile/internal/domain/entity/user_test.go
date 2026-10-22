package entity

import (
	app "profile/internal/pkg/profileerror"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/samber/mo"
)

func TestUpdateValidatesBeforeChangingAggregate(t *testing.T) {
	for _, test := range []struct {
		name    string
		email   string
		plan    Plan
		version int
		code    app.Code
	}{
		{
			name:    "invalid email",
			email:   "bad",
			plan:    Premium,
			version: 2,
			code:    app.InvalidArgument,
		},
		{
			name:    "invalid plan",
			email:   "a@example.test",
			plan:    "UNKNOWN",
			version: 2,
			code:    app.InvalidArgument,
		},
		{
			name:    "stale version",
			email:   "a@example.test",
			plan:    Ultra,
			version: 1,
			code:    app.Conflict,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			user := With(uuid.Must(uuid.NewV4()), WithName("Имя"), WithEmail("old@example.test"), WithPlan(Base), WithVersion(2))
			err := user.Update(mo.Some("Новое имя"), mo.Some(test.email), mo.Some(test.plan), test.version, time.Now())
			if !app.IsCode(err, test.code) {
				t.Fatalf("unexpected error: %v", err)
			}
			if user.Name() != "Имя" || user.Email() != "old@example.test" || user.Plan() != Base || user.Version() != 2 {
				t.Fatal("rejected update partially changed aggregate")
			}
		})
	}
}
func TestSubscriptionUpdatePreservesOtherFields(t *testing.T) {
	user := With(uuid.Must(uuid.NewV4()), WithName("Имя"), WithEmail("a@example.test"), WithPlan(Base), WithVersion(2))
	err := user.Update(mo.None[string](), mo.None[string](), mo.Some(Ultra), 2, time.Now())
	if err != nil || user.Plan() != Ultra || user.Version() != 3 || user.OldVersion() != 2 || user.Email() != "a@example.test" {
		t.Fatalf("update: %v", err)
	}
}
