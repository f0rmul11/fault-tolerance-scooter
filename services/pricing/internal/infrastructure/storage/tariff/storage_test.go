package tariff

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTariffs(t *testing.T) {
	tests := []struct {
		name    string
		content string
		missing bool
		wantErr bool
		loaded  bool
	}{
		{
			name:    "loaded",
			content: `{"tariffs":[{"plan":"BASE","rate_minor":1000,"advance_minor":5000}]}`,
			loaded:  true,
		},
		{
			name:    "empty catalogue",
			content: `{"tariffs":[]}`,
		},
		{
			name:    "invalid json",
			content: `broken`,
			wantErr: true,
		},
		{
			name:    "missing file",
			missing: true,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			filePath := filepath.Join(t.TempDir(), "tariffs.json")
			if !tt.missing {
				if err := os.WriteFile(filePath, []byte(tt.content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			s := NewStorage()
			if s.IsLoaded() {
				t.Fatal("new storage must not be loaded")
			}
			items, err := s.List(context.Background())
			if err != nil || len(items) != 0 {
				t.Fatal("before loading the catalogue must be empty")
			}
			err = s.LoadTariffs(context.Background(), filePath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected loading result: %v", err)
			}
			if s.IsLoaded() != tt.loaded {
				t.Fatal("loaded state must reflect actual data")
			}
			if tt.loaded {
				item, err := s.Get(context.Background(), "BASE")
				if err != nil || item.RateMinor != 1000 || item.AdvanceMinor != 5000 {
					t.Fatalf("incorrect tariff: %+v, %v", item, err)
				}
				items, _ := s.List(context.Background())
				items[0].RateMinor = 1
				item, _ = s.Get(context.Background(), "BASE")
				if item.RateMinor != 1000 {
					t.Fatal("caller changed stored data")
				}
			}
		})
	}
}
