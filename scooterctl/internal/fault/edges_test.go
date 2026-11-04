package fault

import (
	"testing"
	"time"
)

func TestMethodsAreScopedToEdges(t *testing.T) {
	for _, edge := range Edges() {
		options := Options{
			From:        edge.From,
			To:          edge.To,
			Method:      edge.Method,
			Percent:     40,
			Unavailable: true,
			Duration:    time.Minute,
		}
		if err := options.Validate(); err != nil {
			t.Fatal(err)
		}
		options.Method = "ListProfiles"
		if err := options.Validate(); err == nil {
			t.Fatal("unrelated method accepted")
		}
	}
	a := Options{
		From:        "api-gateway",
		To:          "scooter",
		Percent:     5,
		Unavailable: true,
	}
	b := Options{
		From:        "profile",
		To:          "pricing",
		Percent:     100,
		Unavailable: true,
	}
	for key := range a.Values() {
		if _, exists := b.Values()[key]; exists {
			t.Fatalf("faults share runtime key %s", key)
		}
	}
}
