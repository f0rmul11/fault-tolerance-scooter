package status

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/kube"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"text/tabwriter"
	"time"

	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/iostreams"
	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/pkg/cmd/util"
	"github.com/spf13/cobra"
)

type StatusOptions struct {
	util.Options
	Output string
	Watch  bool
}

func NewStatusOptions(streams iostreams.IOStreams) *StatusOptions {
	return &StatusOptions{
		Options: util.NewOptions(streams),
		Output:  "table",
	}
}

func NewCmdStatus(f util.Factory, streams iostreams.IOStreams) *cobra.Command {
	o := NewStatusOptions(streams)
	command := &cobra.Command{
		Use:   "status [service]",
		Short: "Show the node and service status",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(f); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if len(args) == 1 {
				return o.pods(cmd.Context(), args[0])
			}
			if o.Watch {
				return fmt.Errorf("для --watch укажите сервис")
			}
			return o.Run(cmd.Context())
		},
	}
	o.AddFlags(command.Flags())
	command.Flags().BoolVar(&o.Watch, "watch", false, "Обновлять состояние отдельных pod")
	command.Flags().StringVarP(&o.Output, "output", "o", o.Output, "Output format: table or json")
	return command
}

func (o *StatusOptions) Complete(f util.Factory) error {
	return o.Options.Complete(f)
}

func (o *StatusOptions) Validate() error {
	if o.Output != "table" && o.Output != "json" {
		return fmt.Errorf("--output must be table or json")
	}
	return o.Options.Validate()
}

func (o *StatusOptions) Run(ctx context.Context) error {
	runner, err := o.NewStand()
	if err != nil {
		return err
	}
	defer runner.Close()
	ctx, cancel := o.Context(ctx)
	defer cancel()
	return runner.Status(ctx, o.Output)
}

func (o *StatusOptions) pods(ctx context.Context, name string) error {
	var namespace string
	for _, service := range o.Config.Services {
		if service.Name == name {
			namespace = service.Namespace
		}
	}
	if namespace == "" {
		return fmt.Errorf("неизвестный сервис %q", name)
	}
	client, err := kube.New(o.Config.Kubeconfig())
	if err != nil {
		return err
	}
	for {
		pods, err := client.Typed.CoreV1().Pods(namespace).List(ctx, meta.ListOptions{LabelSelector: "app.kubernetes.io/name=" + name})
		if err != nil {
			return err
		}
		fmt.Fprintf(o.Out, "\n%s — %s\n", name, time.Now().Format("15:04:05"))
		w := tabwriter.NewWriter(o.Out, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "POD\tSTATUS\tREADY\tRESTARTS\tREADINESS")
		for _, pod := range pods.Items {
			ready := false
			var restarts int32
			for _, c := range pod.Status.Conditions {
				if c.Type == core.PodReady {
					ready = c.Status == core.ConditionTrue
				}
			}
			for _, c := range pod.Status.ContainerStatuses {
				restarts += c.RestartCount
			}
			phase := string(pod.Status.Phase)
			if pod.DeletionTimestamp != nil {
				phase = "Terminating"
			}
			probeCtx, cancel := context.WithTimeout(ctx, time.Second)
			body, probeErr := client.Typed.CoreV1().RESTClient().Get().Namespace(namespace).Resource("pods").Name(pod.Name + ":8081").SubResource("proxy").Suffix("healthcheck/ready").DoRaw(probeCtx)
			cancel()
			details := string(body)
			var checks map[string]string
			if json.Unmarshal(body, &checks) == nil {
				encoded, _ := json.Marshal(checks)
				details = string(encoded)
			} else if probeErr != nil {
				details = "admin недоступен"
			}
			fmt.Fprintf(w, "%s\t%s\t%t\t%d\t%s\n", pod.Name, phase, ready, restarts, details)
		}
		if err := w.Flush(); err != nil {
			return err
		}
		if !o.Watch {
			return nil
		}
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}
