package kube

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/config"
	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/fault"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestFaultOnlyMatchesGetScooter(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(root, "")
	if err != nil {
		t.Fatal(err)
	}
	objects, err := RenderEnvoy(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var bootstrap map[string]any
	for _, object := range objects {
		if object.GetKind() != "ConfigMap" {
			continue
		}
		value, _, _ := unstructured.NestedString(object.Object, "data", "envoy.json")
		if err := json.Unmarshal([]byte(value), &bootstrap); err != nil {
			t.Fatal(err)
		}
	}
	listeners := bootstrap["static_resources"].(map[string]any)["listeners"].([]any)
	if len(listeners) != len(fault.Edges()) {
		t.Fatal("missing fault listener")
	}
	for index, listener := range listeners {
		hcm := listener.(map[string]any)["filter_chains"].([]any)[0].(map[string]any)["filters"].([]any)[0].(map[string]any)["typed_config"].(map[string]any)
		routes := hcm["route_config"].(map[string]any)["virtual_hosts"].([]any)[0].(map[string]any)["routes"].([]any)
		first, rest := routes[0].(map[string]any), routes[1].(map[string]any)
		if len(routes) != 2 || first["match"].(map[string]any)["path"] != fault.Edges()[index].Method || first["typed_per_filter_config"] == nil || rest["typed_per_filter_config"] != nil {
			t.Fatalf("fault route scope incorrect: %v", routes)
		}
	}
	for _, name := range []string{"api-gateway", "profile", "rental"} {
		services, err := cfg.Select([]string{name})
		if err != nil {
			t.Fatal(err)
		}
		rendered, err := Render(root, services[0], "test:1")
		if err != nil {
			t.Fatal(err)
		}
		hasOverride := false
		for _, object := range rendered {
			containers, _, _ := unstructured.NestedSlice(object.Object, "spec", "template", "spec", "containers")
			for _, container := range containers {
				env, _, _ := unstructured.NestedSlice(container.(map[string]any), "env")
				for _, raw := range env {
					entry := raw.(map[string]any)
					if entry["name"] == "SERVICE_TARGETS" {
						hasOverride = true
						expected := map[string]string{"api-gateway": "scooter:envoy.scooterflow.svc.cluster.local:10001", "profile": "pricing:envoy.scooterflow.svc.cluster.local:10002"}[name]
						if entry["value"] != expected {
							t.Fatalf("unexpected targets: %v", entry)
						}
					}
				}
			}
		}
		if hasOverride != (name == "api-gateway" || name == "profile") {
			t.Fatalf("unexpected override for %s", name)
		}
	}
}
