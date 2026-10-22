package entity

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"

	app "profile/internal/pkg/profileerror"
)

func TestFieldChangesValidateBeforeMutation(t *testing.T) {
	tests := []struct {
		name   string
		change func(*User) error
	}{
		{
			name: "empty name",
			change: func(user *User) error {
				return user.Rename(" ", time.Now())
			},
		},
		{
			name: "invalid email",
			change: func(user *User) error {
				return user.ChangeEmail("bad", time.Now())
			},
		},
		{
			name: "invalid subscription",
			change: func(user *User) error {
				return user.ChangeSubscription("UNKNOWN", time.Now())
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user := With(
				uuid.Must(uuid.NewV4()),
				WithName("Имя"),
				WithEmail("old@example.test"),
				WithPlan(Base),
				WithVersion(2),
			)
			if err := test.change(user); !app.IsCode(err, app.InvalidArgument) {
				t.Fatalf("unexpected error: %v", err)
			}
			if user.Name() != "Имя" || user.Email() != "old@example.test" || user.Plan() != Base || user.Version() != 2 {
				t.Fatal("rejected field change mutated profile")
			}
		})
	}
}

func TestSubscriptionChangePreservesOtherFields(t *testing.T) {
	user := With(
		uuid.Must(uuid.NewV4()),
		WithName("Имя"),
		WithEmail("a@example.test"),
		WithPlan(Base),
		WithVersion(2),
	)
	if err := user.ChangeSubscription(Ultra, time.Now()); err != nil {
		t.Fatal(err)
	}
	if user.Plan() != Ultra || user.Version() != 3 || user.OldVersion() != 2 || user.Email() != "a@example.test" {
		t.Fatal("subscription change affected unrelated fields")
	}
}
