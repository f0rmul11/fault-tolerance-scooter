package loyalty

import "testing"

func TestDiscountThresholds(t *testing.T) {
	for _, tc := range []struct {
		trips   int64
		percent int32
	}{{0, 0}, {9, 0}, {10, 10}, {49, 10}, {50, 20}, {55, 20}} {
		if got := DiscountPercent(tc.trips); got != tc.percent {
			t.Errorf("trips=%d: got %d, want %d", tc.trips, got, tc.percent)
		}
	}
}
