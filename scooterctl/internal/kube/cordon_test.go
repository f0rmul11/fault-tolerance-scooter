package kube

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestSetCordonTargetsOnlySelectedServicePod(t *testing.T) {
	for _, test := range []struct {
		name, label string
		cordoned    bool
		status      int
		wantError   bool
	}{
		{name: "cordon", label: "pricing", cordoned: true, status: 200},
		{name: "uncordon", label: "pricing", status: 200},
		{name: "wrong service", label: "rental", cordoned: true, status: 200, wantError: true},
		{name: "not implemented", label: "pricing", cordoned: true, status: 404, wantError: true},
		{name: "handler error", label: "pricing", cordoned: true, status: 500, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			posts := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet && r.URL.Path == "/api/v1/namespaces/scooterflow/pods/pricing-selected" {
					json.NewEncoder(w).Encode(map[string]any{"apiVersion": "v1", "kind": "Pod", "metadata": map[string]any{"name": "pricing-selected", "labels": map[string]string{"app.kubernetes.io/name": test.label}}})
					return
				}
				action := "uncordon"
				if test.cordoned {
					action = "cordon"
				}
				if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/pods/pricing-selected:8081/proxy/"+action) {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				posts++
				w.WriteHeader(test.status)
				w.Write([]byte(`{}`))
			}))
			defer server.Close()
			typed, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
			if err != nil {
				t.Fatal(err)
			}
			c := &Client{Typed: typed}
			err = c.SetCordon(t.Context(), config.Service{Name: "pricing", Namespace: "scooterflow"}, "pricing-selected", test.cordoned)
			if (err != nil) != test.wantError {
				t.Fatalf("error=%v", err)
			}
			if test.label != "pricing" && posts != 0 {
				t.Fatal("another service was modified")
			}
			if test.label == "pricing" && posts != 1 {
				t.Fatalf("posts=%d", posts)
			}
		})
	}
}
