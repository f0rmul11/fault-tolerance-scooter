package kube

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/config"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestFollowLogsOutlivesRESTRequestTimeout(t *testing.T) {
	release := make(chan struct{})
	canceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/apis/apps/v1/namespaces/test/deployments/gateway":
			_ = json.NewEncoder(w).Encode(&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "gateway"}}},
			})
		case "/api/v1/namespaces/test/pods":
			_ = json.NewEncoder(w).Encode(&corev1.PodList{Items: []corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "gateway-pod"}}}})
		case "/api/v1/namespaces/test/pods/gateway-pod/log":
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			select {
			case <-release:
				_, _ = w.Write([]byte("late log line\n"))
			case <-r.Context().Done():
				close(canceled)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	cfg := &rest.Config{Host: server.URL, Timeout: 50 * time.Millisecond}
	typed, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	client := &Client{Typed: typed, REST: cfg}
	var output bytes.Buffer
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- client.Logs(ctx, config.Service{Namespace: "test", Deployment: "gateway", Name: "gateway"}, true, 0, &output)
	}()
	// Real HTTP streaming test: intentionally exceed the shared REST timeout.
	timer := time.NewTimer(200 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-canceled:
		t.Fatal("follow stream inherited the REST timeout")
	case err := <-done:
		t.Fatalf("follow exited before the late line: %v", err)
	case <-timer.C:
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte("late log line")) {
		t.Fatalf("lost late log: %s", &output)
	}
	if cfg.Timeout != 50*time.Millisecond {
		t.Fatal("follow mutated the shared REST config")
	}
}
