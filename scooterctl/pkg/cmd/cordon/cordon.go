package cordon

import (
	"fmt"
	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/iostreams"
	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/kube"
	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/pkg/cmd/util"
	"github.com/spf13/cobra"
)

func New(f util.Factory, streams iostreams.IOStreams, cordoned bool) *cobra.Command {
	action := "uncordon"
	if cordoned {
		action = "cordon"
	}
	o := util.NewOptions(streams)
	command := &cobra.Command{
		Use:   action + " SERVICE POD",
		Short: "Вызвать POST /" + action + " у конкретного pod через admin-сервер",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(f); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			services, err := o.Config.Select([]string{args[0]})
			if err != nil {
				return err
			}
			client, err := kube.New(o.Config.Kubeconfig())
			if err != nil {
				return err
			}
			ctx, cancel := o.Context(cmd.Context())
			defer cancel()
			if err := client.SetCordon(ctx, services[0], args[1], cordoned); err != nil {
				return err
			}
			fmt.Fprintf(streams.Out, "%s: POST /%s выполнен. Проверьте состояние: scooterctl status %s\n", args[1], action, args[0])
			return nil
		},
	}
	o.AddFlags(command.Flags())
	return command
}
