package load

import (
	"bytes"
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/f0rmul11/fault-tolerance-scooter/scooterctl/internal/iostreams"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type delayedHealth struct {
	health.UnimplementedHealthServer
	deadlines chan time.Duration
}

func (s *delayedHealth) Check(ctx context.Context, _ *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return nil, status.Error(codes.Internal, "missing deadline")
	}
	s.deadlines <- time.Until(deadline)
	<-ctx.Done()
	return nil, status.FromContextError(ctx.Err()).Err()
}

// Opt in with the patched binary built by tools/install-framer.sh.
// This catches reserved metadata silently dropped by upstream Framer.
func TestFramerSendsRequestDeadline(t *testing.T) {
	bin := os.Getenv("SCOOTERCTL_FRAMER_BIN")
	if bin == "" {
		t.Skip("set SCOOTERCTL_FRAMER_BIN for the real Framer integration test")
	}
	root := t.TempDir()
	cached := filepath.Join(root, ".local", "scooterctl", "tools", "framer", revision, "framer")
	if err := os.MkdirAll(filepath.Dir(cached), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(bin, cached); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	dependency := &delayedHealth{deadlines: make(chan time.Duration, 10)}
	health.RegisterHealthServer(server, dependency)
	reflection.Register(server)
	t.Cleanup(server.Stop)
	go func() {
		_ = server.Serve(listener)
	}()
	opts := &options{
		addr:           listener.Addr().String(),
		method:         "grpc.health.v1.Health/Check",
		data:           "{}",
		rps:            1,
		clients:        1,
		duration:       time.Second,
		requestTimeout: 100 * time.Millisecond,
	}
	payload, md, err := opts.validate()
	if err != nil {
		t.Fatal(err)
	}
	var output, diagnostic bytes.Buffer
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	err = opts.run(ctx, root, iostreams.IOStreams{Out: &output, ErrOut: &diagnostic}, payload, md)
	if err != nil {
		t.Fatalf("load: %v\n%s", err, &output)
	}
	select {
	case remaining := <-dependency.deadlines:
		if remaining <= 0 || remaining > 100*time.Millisecond {
			t.Fatalf("server deadline: %s", remaining)
		}
	default:
		t.Fatal("server did not receive a deadline")
	}
	reports, err := filepath.Glob(filepath.Join(root, ".local", "scooterctl", "load", "*", "report.phout"))
	if err != nil || len(reports) != 1 {
		t.Fatalf("reports: %v %v", reports, err)
	}
	result, err := readReport(reports[0], &bytes.Buffer{})
	if err != nil || result.Completed != 1 || result.Errors != 1 || result.P99MS > 500 {
		t.Fatalf("deadline report: %+v %v", result, err)
	}
}
