package fault

import (
	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/fault"
	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/iostreams"
	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/pkg/cmd/util"
	"github.com/spf13/cobra"
)

func NewCmdFault(f util.Factory, streams iostreams.IOStreams) *cobra.Command {
	options := fault.Options{}
	cmd := &cobra.Command{
		Use:     "fault FROM TO",
		Short:   "Добавить ошибки или задержку на выбранной RPC-связи",
		Long:    "Сбой действует до окончания --duration или Ctrl+C. Дождитесь ACTIVE перед запросами.\nПосле аварийного завершения CLI снимите воздействие через --clear.\nПоддерживается api-gateway → scooter, метод GetScooter; остальные вызовы не затрагиваются.",
		Example: "  scooterctl fault api-gateway scooter --method GetScooter --unavailable --percent 5 --duration 15m\n  scooterctl fault api-gateway scooter --method GetScooter --slowdown 500ms --duration 5m\n  scooterctl fault api-gateway scooter --clear",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.From, options.To = args[0], args[1]
			if err := options.Validate(); err != nil {
				return err
			}
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			runner, err := f.NewStand(cfg, streams)
			if err != nil {
				return err
			}
			defer runner.Close()
			return runner.Fault(cmd.Context(), options)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&options.Method, "method", "", "RPC: GetScooter или полное имя метода")
	flags.IntVar(&options.Percent, "percent", 100, "Вероятность сбоя для каждой попытки, от 1 до 100 процентов")
	flags.BoolVar(&options.Unavailable, "unavailable", false, "Вернуть gRPC Unavailable до обработки сервисом")
	flags.DurationVar(&options.Slowdown, "slowdown", 0, "Дополнительная задержка перед обработкой, от 1ms до 1m")
	flags.DurationVar(&options.Duration, "duration", 0, "Как долго действует сбой, от 1ms до 1h")
	flags.BoolVar(&options.Clear, "clear", false, "Снять сбой после аварийного завершения команды")
	cmd.MarkFlagsMutuallyExclusive("slowdown", "unavailable")
	for _, flag := range []string{"method", "percent", "slowdown", "unavailable", "duration"} {
		cmd.MarkFlagsMutuallyExclusive("clear", flag)
	}
	return cmd
}
