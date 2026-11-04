package discount

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"

	"profile/internal/domain/entity"
)

type profileStore struct{ err error }

func (s profileStore) Get(context.Context, uuid.UUID) (*entity.User, error) {
	return nil, s.err
}

type pricingClient struct {
	calls int
	ctx   context.Context
	err   error
}

func (p *pricingClient) CompletedTripsCount(ctx context.Context, _ uuid.UUID) (int64, error) {
	p.calls++
	p.ctx = ctx
	return 12, p.err
}

func TestLocalFailureDoesNotCallPricing(t *testing.T) {
	dbErr := errors.New("database unavailable")
	client := &pricingClient{}
	uc := New(profileStore{err: dbErr}, client)
	userID := uuid.FromStringOrNil("704137fe-93b3-4113-8e09-ed1df4d0e0ea")
	if _, err := uc.Execute(context.Background(), userID); !errors.Is(err, dbErr) {
		t.Fatalf("lost database error: %v", err)
	}
	if _, err := uc.Execute(context.Background(), uuid.Nil); err == nil {
		t.Fatal("invalid user accepted")
	}
	if client.calls != 0 {
		t.Fatal("called dependency after local failure")
	}
}

func TestPricingResultAndContext(t *testing.T) {
	client := &pricingClient{}
	uc := New(profileStore{}, client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := uc.Execute(ctx, uuid.FromStringOrNil("704137fe-93b3-4113-8e09-ed1df4d0e0ea"))
	if err != nil || result.Percent != 10 || result.CompletedTripsCount == nil || *result.CompletedTripsCount != 12 || result.Degraded || result.Source != "pricing" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if client.ctx != ctx {
		t.Fatal("lost incoming context")
	}
	client.err = errors.New("dependency unavailable")
	if result, err := uc.Execute(ctx, uuid.FromStringOrNil("704137fe-93b3-4113-8e09-ed1df4d0e0ea")); result != nil || !errors.Is(err, client.err) {
		t.Fatalf("baseline masked dependency failure: result=%v err=%v", result, err)
	}
}
