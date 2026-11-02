package kube

import (
	"context"
	"fmt"
	"strings"

	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/config"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) SetCordon(ctx context.Context, service config.Service, podName string, cordoned bool) error {
	pod, err := c.Typed.CoreV1().Pods(service.Namespace).Get(ctx, podName, meta.GetOptions{})
	if err != nil {
		return err
	}
	if pod.Labels["app.kubernetes.io/name"] != service.Name {
		return fmt.Errorf("pod %s не принадлежит сервису %s", podName, service.Name)
	}
	action := "uncordon"
	if cordoned {
		action = "cordon"
	}
	var status int
	result := c.Typed.CoreV1().RESTClient().Post().Namespace(service.Namespace).Resource("pods").Name(podName + ":8081").SubResource("proxy").Suffix(action).Body([]byte("{}")).Do(ctx)
	result.StatusCode(&status)
	body, err := result.Raw()
	if status == 404 {
		return fmt.Errorf("POST /%s недоступен у %s; реализуйте ручку на admin-сервере и выкатите сервис", action, podName)
	}
	if err != nil {
		return fmt.Errorf("POST /%s у %s: %w", action, podName, err)
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("POST /%s: HTTP %d: %s", action, status, strings.TrimSpace(string(body)))
	}
	return nil
}
