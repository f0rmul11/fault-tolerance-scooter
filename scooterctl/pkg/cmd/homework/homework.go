package homework

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/iostreams"
	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/kube"
	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/pkg/cmd/util"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const namespace = "scooterflow"

type runner struct {
	client  *kube.Client
	out     io.Writer
	address string
	checks  []check
}
type check struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
}

func New(f util.Factory, streams iostreams.IOStreams) *cobra.Command {
	root := &cobra.Command{Use: "homework", Short: "Проверки домашних заданий"}
	module := &cobra.Command{Use: "module-2"}
	run := &cobra.Command{Use: "run readiness|shutdown|cordon", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		if args[0] != "readiness" && args[0] != "shutdown" && args[0] != "cordon" {
			return fmt.Errorf("неизвестный сценарий %q", args[0])
		}
		cfg, err := f.Config()
		if err != nil {
			return err
		}
		client, err := kube.New(cfg.Kubeconfig())
		if err != nil {
			return err
		}
		address := ""
		for _, s := range cfg.Services {
			if s.Name == "pricing" {
				address = fmt.Sprintf("127.0.0.1:%d", s.GRPC.Host)
			}
		}
		r := &runner{client: client, out: streams.Out, address: address}
		ctx, cancel := context.WithTimeout(command.Context(), 4*time.Minute)
		defer cancel()
		switch args[0] {
		case "readiness":
			err = r.readiness(ctx, cfg.Runtime.Ports.HTTP)
		case "shutdown":
			err = r.shutdown(ctx)
		case "cordon":
			err = r.cordon(ctx)
		}
		if err != nil {
			r.record("Сценарий завершён", false, err.Error())
		}
		for _, c := range r.checks {
			if !c.Passed {
				return fmt.Errorf("проверки не пройдены; исправьте код и повторите ту же команду")
			}
		}
		return nil
	}}
	module.AddCommand(run)
	root.AddCommand(module)
	return root
}
func (r *runner) record(name string, ok bool, details string) {
	r.checks = append(r.checks, check{name, ok, details})
	label := "FAIL"
	if ok {
		label = "OK"
	}
	fmt.Fprintf(r.out, "%s %s: %s\n", label, name, details)
}
func wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
func (r *runner) pods(ctx context.Context) ([]core.Pod, error) {
	p, e := r.client.Typed.CoreV1().Pods(namespace).List(ctx, meta.ListOptions{LabelSelector: "app.kubernetes.io/name=pricing"})
	if e != nil {
		return nil, e
	}
	return p.Items, nil
}
func (r *runner) admin(ctx context.Context, pod, path, method string) (int, []byte) {
	var status int
	req := r.client.Typed.CoreV1().RESTClient().Verb(method).Namespace(namespace).Resource("pods").Name(pod + ":8081").SubResource("proxy").Suffix(path)
	if method == "POST" {
		req.Body([]byte("{}"))
	}
	result := req.Do(ctx)
	result.StatusCode(&status)
	body, err := result.Raw()
	if err != nil && len(body) == 0 {
		body = []byte(err.Error())
	}
	return status, body
}
func ready(p core.Pod) bool {
	for _, c := range p.Status.Conditions {
		if c.Type == core.PodReady {
			return c.Status == core.ConditionTrue
		}
	}
	return false
}
func restarts(p core.Pod) int32 {
	var n int32
	for _, c := range p.Status.ContainerStatuses {
		n += c.RestartCount
	}
	return n
}
func (r *runner) settled(ctx context.Context) error {
	return kube.Poll(ctx, "ожидание прогретых Pricing", func(ctx context.Context) (bool, error) {
		pods, err := r.pods(ctx)
		if err != nil {
			return false, err
		}
		if len(pods) < 2 {
			return false, fmt.Errorf("нужно хотя бы две реплики Pricing; выполните deploy pricing")
		}
		for _, p := range pods {
			if p.DeletionTimestamp != nil || !ready(p) {
				return false, nil
			}
		}
		// A bad readiness can already be green: check actual business responses too.
		for i := 0; i < 6; i++ {
			_, data, err := r.rpc(ctx)
			if err != nil || len(data) == 0 {
				return false, nil
			}
		}
		return true, nil
	})
}

// Raw protobuf codec keeps the CLI independent of generated service packages.
type rawCodec struct{}

func (rawCodec) Name() string {
	return "proto"
}
func (rawCodec) Marshal(v any) ([]byte, error) {
	return *v.(*[]byte), nil
}
func (rawCodec) Unmarshal(b []byte, v any) error {
	*v.(*[]byte) = append([]byte(nil), b...)
	return nil
}
func (r *runner) rpc(ctx context.Context) (string, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(r.address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", nil, err
	}
	defer conn.Close()
	var input, output []byte
	var headers metadata.MD
	err = conn.Invoke(ctx, "/pricing.pricing.v1.PricingService/ListTariffs", &input, &output, grpc.ForceCodec(rawCodec{}), grpc.Header(&headers))
	pod := ""
	if values := headers.Get("x-pricing-pod"); len(values) > 0 {
		pod = values[0]
	}
	return pod, output, err
}
func (r *runner) readiness(ctx context.Context, port int) error {
	fmt.Fprintln(r.out, "Ожидание готовых реплик и загруженных тарифов…")
	if err := r.settled(ctx); err != nil {
		return err
	}
	var mu sync.Mutex
	counts := map[string]int{}
	sampleCtx, stop := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		client := &http.Client{Timeout: 2 * time.Second}
		for sampleCtx.Err() == nil {
			req, _ := http.NewRequestWithContext(sampleCtx, "GET", fmt.Sprintf("http://127.0.0.1:%d/v1/pricing/tariffs", port), nil)
			response, err := client.Do(req)
			kind := "error"
			if err == nil {
				var result struct {
					Tariffs []json.RawMessage `json:"tariffs"`
				}
				err = json.NewDecoder(response.Body).Decode(&result)
				response.Body.Close()
				if err == nil && response.StatusCode == 200 {
					kind = "full"
					if len(result.Tariffs) == 0 {
						kind = "empty"
					}
				}
			}
			if sampleCtx.Err() != nil {
				return
			}
			mu.Lock()
			counts[kind]++
			mu.Unlock()
			_ = wait(sampleCtx, 100*time.Millisecond)
		}
	}()
	defer func() {
		stop()
		wg.Wait()
	}()
	if err := wait(ctx, 3*time.Second); err != nil {
		return err
	}
	fmt.Fprintln(r.out, "Запросы через Gateway запущены. Выполняем rolling update Pricing…")
	patch := []byte(fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"scooterflow/homework-restart":%q}}}}}`, time.Now().Format(time.RFC3339Nano)))
	if _, err := r.client.Typed.AppsV1().Deployments(namespace).Patch(ctx, "pricing", types.MergePatchType, patch, meta.PatchOptions{FieldManager: "scooterctl-homework"}); err != nil {
		return err
	}
	deadline := time.Now().Add(140 * time.Second)
	observedWarming := false
	completed := false
	for time.Now().Before(deadline) {
		pods, err := r.pods(ctx)
		if err != nil {
			return err
		}
		done := true
		for _, p := range pods {
			if p.DeletionTimestamp != nil {
				done = false
				continue
			}
			if !ready(p) {
				observedWarming = true
				done = false
			}
			r.recordRestarts(p)
		}
		d, err := r.client.Typed.AppsV1().Deployments(namespace).Get(ctx, "pricing", meta.GetOptions{})
		if err != nil {
			return err
		}
		if d.Status.UpdatedReplicas != *d.Spec.Replicas || d.Status.ObservedGeneration < d.Generation {
			done = false
		}
		if done && time.Since(d.CreationTimestamp.Time) > 0 {
			// Even broken probes must be sampled through the complete 30s warmup.
			if err := wait(ctx, 35*time.Second); err != nil {
				return err
			}
			completed = true
			break
		}
		if err := wait(ctx, time.Second); err != nil {
			return err
		}
	}
	stop()
	wg.Wait()
	fmt.Fprintf(r.out, "\nТРАФИК: с тарифами=%d · пустых HTTP 200=%d · ошибок=%d\n", counts["full"], counts["empty"], counts["error"])
	r.record("Rollout завершился", completed, "все новые реплики готовы, старые удалены")
	r.record("Тарифы при rollout", counts["full"] > 0 && counts["empty"] == 0 && counts["error"] == 0, fmt.Sprint(counts))
	r.record("Новая реплика ждала готовности", observedWarming, "наблюдение Pod Ready")
	return r.isolated(ctx, "missing-file", false)
}
func (r *runner) recordRestarts(p core.Pod) {
	if restarts(p) > 0 {
		r.record("Без перезапусков", false, p.Name)
	}
}
func (r *runner) isolated(ctx context.Context, mode string, cordoned bool) error {
	fmt.Fprintln(r.out, "Изолированная проверка:", mode)
	d, err := r.client.Typed.AppsV1().Deployments(namespace).Get(ctx, "pricing", meta.GetOptions{})
	if err != nil {
		return err
	}
	pod := &core.Pod{ObjectMeta: meta.ObjectMeta{GenerateName: "homework-pricing-", Namespace: namespace, Labels: map[string]string{"scooterflow/homework": "module-2"}}, Spec: *d.Spec.Template.Spec.DeepCopy()}
	pod.Spec.RestartPolicy = core.RestartPolicyNever
	if mode == "missing-file" {
		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, core.EnvVar{Name: "TARIFFS_FILE", Value: "/missing-tariffs.json"})
	}
	pod, err = r.client.Typed.CoreV1().Pods(namespace).Create(ctx, pod, meta.CreateOptions{})
	if err != nil {
		return err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = r.client.Typed.CoreV1().Pods(namespace).Delete(cleanup, pod.Name, meta.DeleteOptions{})
	}()
	if err := kube.Poll(ctx, "admin временного Pricing", func(ctx context.Context) (bool, error) {
		code, _ := r.admin(ctx, pod.Name, "healthcheck/live", "GET")
		return code == 200, nil
	}); err != nil {
		return err
	}
	if cordoned {
		code, _ := r.admin(ctx, pod.Name, "uncordon", "POST")
		r.record("Uncordon во время прогрева доступен", code == 200, fmt.Sprint(code))
		code, _ = r.admin(ctx, pod.Name, "healthcheck/ready", "GET")
		r.record("Uncordon не отменяет прогрев", code == 503, fmt.Sprint(code))
		code, _ = r.admin(ctx, pod.Name, "cordon", "POST")
		r.record("Cordon во время прогрева", code == 200, fmt.Sprint(code))
	}
	if err := wait(ctx, 35*time.Second); err != nil {
		return err
	}
	code, body := r.admin(ctx, pod.Name, "healthcheck/ready", "GET")
	r.record("Readiness: "+mode, code == 503, string(body))
	if mode == "missing-file" {
		logs, err := r.client.Typed.CoreV1().Pods(namespace).GetLogs(pod.Name, &core.PodLogOptions{}).DoRaw(ctx)
		r.record("Ошибка файла видна в логах", err == nil && bytes.Contains(logs, []byte("missing-tariffs")), "журнал загрузки Pricing")
	}
	live, _ := r.admin(ctx, pod.Name, "healthcheck/live", "GET")
	r.record("Liveness при "+mode, live == 200, fmt.Sprint(live))
	if mode == "missing-file" {
		status, _ := r.admin(ctx, pod.Name, "uncordon", "POST")
		if status == 200 {
			code, _ = r.admin(ctx, pod.Name, "healthcheck/ready", "GET")
			r.record("Uncordon не отменяет ошибку файла", code == 503, fmt.Sprint(code))
		}
	}
	p, err := r.client.Typed.CoreV1().Pods(namespace).Get(ctx, pod.Name, meta.GetOptions{})
	if err != nil {
		return err
	}
	r.record("Прогрев не перезапускает Pod", restarts(*p) == 0 && p.Status.Phase == core.PodRunning, fmt.Sprint(p.Status.Phase))
	return nil
}
func (r *runner) shutdown(ctx context.Context) error {
	fmt.Fprintln(r.out, "Ожидание готовых реплик и загруженных тарифов…")
	if err := r.settled(ctx); err != nil {
		return err
	}
	pods, _ := r.pods(ctx)
	p := pods[0]
	logCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream, err := r.client.Typed.CoreV1().Pods(namespace).GetLogs(p.Name, &core.PodLogOptions{Follow: true, Timestamps: true}).Stream(logCtx)
	if err != nil {
		return err
	}
	defer stream.Close()
	var logs bytes.Buffer
	copied := make(chan struct{})
	go func() {
		_, _ = io.Copy(&logs, stream)
		close(copied)
	}()
	if err := r.client.Typed.CoreV1().Pods(namespace).Delete(ctx, p.Name, meta.DeleteOptions{}); err != nil {
		return err
	}
	observed := false
	liveOK := false
	revived := false
	for i := 0; i < 15; i++ {
		code, body := r.admin(ctx, p.Name, "healthcheck/ready", "GET")
		if code == 503 && bytes.Contains(body, []byte("termination")) {
			observed = true
			status, _ := r.admin(ctx, p.Name, "healthcheck/live", "GET")
			liveOK = liveOK || status == 200
			status, _ = r.admin(ctx, p.Name, "uncordon", "POST")
			if status == 200 {
				after, _ := r.admin(ctx, p.Name, "healthcheck/ready", "GET")
				revived = revived || after == 200
			}
		}
		if err := wait(ctx, 150*time.Millisecond); err != nil {
			return err
		}
	}
	r.record("termination снимает readiness", observed, p.Name)
	r.record("Liveness во время shutdown", liveOK, "live=200 при ready=503")
	r.record("Uncordon не оживляет shutdown", !revived, "termination сохраняется")
	if err := wait(ctx, 3*time.Second); err != nil {
		return err
	}
	cancel()
	stream.Close()
	<-copied
	fmt.Fprintln(r.out, "Журнал остановки:")
	fmt.Fprint(r.out, logs.String())
	return nil
}
func (r *runner) cordon(ctx context.Context) error {
	fmt.Fprintln(r.out, "Ожидание готовых реплик и загруженных тарифов…")
	if err := r.settled(ctx); err != nil {
		return err
	}
	pods, _ := r.pods(ctx)
	p := pods[0]
	defer r.admin(context.Background(), p.Name, "uncordon", "POST")
	for i := 0; i < 2; i++ {
		code, _ := r.admin(ctx, p.Name, "cordon", "POST")
		r.record("POST /cordon", code == 200, fmt.Sprint(code))
	}
	code, _ := r.admin(ctx, p.Name, "healthcheck/ready", "GET")
	r.record("Cordon снимает readiness", code == 503, fmt.Sprint(code))
	code, _ = r.admin(ctx, p.Name, "healthcheck/live", "GET")
	r.record("Cordon сохраняет liveness", code == 200, fmt.Sprint(code))
	if err := wait(ctx, 7*time.Second); err != nil {
		return err
	}
	routed := map[string]int{}
	trafficOK := true
	for i := 0; i < 30; i++ {
		name, data, err := r.rpc(ctx)
		if err != nil || len(data) == 0 || name == "" || name == p.Name {
			trafficOK = false
		}
		routed[name]++
	}
	fmt.Fprintln(r.out, "Распределение новых соединений (pod → запросы):", routed)
	r.record("Новые соединения обходят cordon", trafficOK, fmt.Sprint(routed))
	endpoints, err := r.client.Typed.DiscoveryV1().EndpointSlices(namespace).List(ctx, meta.ListOptions{LabelSelector: "kubernetes.io/service-name=pricing"})
	if err != nil {
		return err
	}
	excluded := true
	for _, slice := range endpoints.Items {
		for _, ep := range slice.Endpoints {
			if ep.TargetRef != nil && ep.TargetRef.Name == p.Name && (ep.Conditions.Ready == nil || *ep.Conditions.Ready) {
				excluded = false
			}
		}
	}
	r.record("Pod исключён из готовых адресов", excluded, p.Name)
	for i := 0; i < 2; i++ {
		code, _ = r.admin(ctx, p.Name, "uncordon", "POST")
		r.record("POST /uncordon", code == 200, fmt.Sprint(code))
	}
	code, _ = r.admin(ctx, p.Name, "healthcheck/ready", "GET")
	r.record("Uncordon возвращает готовность", code == 200, fmt.Sprint(code))
	if err := wait(ctx, 7*time.Second); err != nil {
		return err
	}
	returned := false
	for i := 0; i < 40; i++ {
		name, _, err := r.rpc(ctx)
		if err == nil && name == p.Name {
			returned = true
		}
	}
	r.record("Pod снова получает новые соединения", returned, p.Name)
	after, err := r.client.Typed.CoreV1().Pods(namespace).Get(ctx, p.Name, meta.GetOptions{})
	if err != nil {
		return err
	}
	r.record("Состояние экземпляра сохранено", after.UID == p.UID && restarts(*after) == restarts(p), p.Name)
	for _, other := range pods {
		if other.UID == p.UID {
			continue
		}
		current, err := r.client.Typed.CoreV1().Pods(namespace).Get(ctx, other.Name, meta.GetOptions{})
		r.record("Другой Pod не затронут", err == nil && current.UID == other.UID && ready(*current), other.Name)
	}
	if err := r.isolated(ctx, "cordon-warmup", true); err != nil {
		return err
	}
	return r.isolated(ctx, "missing-file", false)
}
