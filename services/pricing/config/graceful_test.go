package config

import (
	"testing"
	"time"
)

func TestGracefulDelay(t *testing.T) {
	t.Setenv("CONFIG_PATH", "config.yaml")
	t.Setenv("GRACEFUL_TIMEOUT", "10s")
	tests := []struct {
		name  string
		delay string
		valid bool
	}{
		{name: "disabled", delay: "0s", valid: true},
		{name: "draining", delay: "3s", valid: true},
		{name: "negative", delay: "-1s"},
		{name: "equal to timeout", delay: "10s"},
		{name: "longer than timeout", delay: "11s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GRACEFUL_DELAY", tt.delay)
			cfg, err := Read()
			if !tt.valid {
				if err == nil {
					t.Fatal("invalid graceful delay accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			expected, err := time.ParseDuration(tt.delay)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Graceful.Delay != expected {
				t.Fatalf("expected delay %v, got %v", expected, cfg.Graceful.Delay)
			}
		})
	}
}
