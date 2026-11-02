package stand

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/fault"
	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/kube"
	"github.com/gofrs/flock"
)

func (m *Manager) Fault(ctx context.Context, options fault.Options) error {
	if err := options.Validate(); err != nil {
		return err
	}
	state, err := m.running(ctx)
	if err != nil {
		return err
	}
	if state.Reset != nil {
		return fmt.Errorf("сначала завершите scooterctl reset --yes")
	}
	lock := flock.New(filepath.Join(m.Config.StateDir(), "fault-"+options.From+"-"+options.To+".lock"))
	locked, err := lock.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		return fmt.Errorf("на этой связи уже работает команда fault; остановите её через Ctrl+C")
	}
	defer lock.Unlock()
	edge, err := fault.FindEdge(options.From, options.To)
	if err != nil {
		return err
	}
	services, err := m.Config.Select([]string{edge.From, edge.To})
	if err != nil {
		return err
	}
	client, err := kube.New(m.Config.Kubeconfig())
	if err != nil {
		return err
	}
	setup, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if !options.Clear {
		if err := client.CheckEnvoyRoute(setup, services[0], edge); err != nil {
			return err
		}
	}
	forward, err := client.ForwardEnvoy(setup, []int{fault.ControlPort}, m.Streams.ErrOut)
	if err != nil {
		return err
	}
	defer forward.Close()
	control := fault.NewRuntimeClient(fmt.Sprintf("http://127.0.0.1:%d", forward.Ports[fault.ControlPort]))
	defer control.Close()
	return fault.RunRuntime(ctx, control, options, m.Streams.Out)
}
