package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/not-for-prod/clay/server"
	"github.com/not-for-prod/clay/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"pricing/config"
	"pricing/internal/app/echo/v1"
	pricingapp "pricing/internal/app/pricing/v1"
	"pricing/internal/application/usecase"
	"pricing/internal/infrastructure/gateway"
	"pricing/internal/infrastructure/inbox/rental_events"
	"pricing/internal/infrastructure/inbox/rental_events/rental_canceled"
	"pricing/internal/infrastructure/inbox/rental_events/rental_completed"
	"pricing/internal/infrastructure/inbox/rental_events/rental_reserved"
	"pricing/internal/infrastructure/inbox/rental_events/rental_started"
	"pricing/internal/infrastructure/messagebus"
	"pricing/internal/infrastructure/storage"
	"pricing/internal/pkg/closer"
	"pricing/internal/pkg/connector/postgres"
	event_router "pricing/internal/pkg/event-router"
	"pricing/internal/pkg/grpc/intercept"
	"pricing/internal/pkg/grpcdesc"
	"pricing/internal/pkg/healthcheck"
	"pricing/internal/pkg/inbox"
	echoV1 "pricing/internal/pkg/pb/pricing/echo/v1"
	businessV1 "pricing/internal/pkg/pb/pricing/pricing/v1"
	"pricing/internal/pkg/worker"
)

func (a *App) initControllers(_ context.Context) error {
	a.controllers = []transport.ServiceDesc{
		grpcdesc.New(func(server *grpc.Server) {
			echoV1.RegisterEchoServiceServer(server, echo.NewEchoService())
			businessV1.RegisterPricingServiceServer(server, pricingapp.New(a.usecases))
		}),
	}
	return nil
}

func (a *App) initMainServer(ctx context.Context) error {
	serverReady := make(chan struct{})
	var readyOnce sync.Once

	// init server (htt,grpc)
	a.mainServer = server.NewServer(
		config.Instance().GrpcServer.Port,
		server.WithHTTPPort(config.Instance().HttpServer.Port),
		server.WithHTTPMiddlewares(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// HTTP Serve starts after both Clay servers have been initialized.
				readyOnce.Do(func() {
					close(serverReady)
				})
				next.ServeHTTP(w, r)
			})
		}),
		server.WithGRPCOpts(
			grpc.ChainUnaryInterceptor(intercept.ErrorInterceptor()),
			grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle: config.Instance().GrpcServer.MaxConnectionIdle,
				MaxConnectionAge:  config.Instance().GrpcServer.MaxConnectionAge,
				Time:              config.Instance().GrpcServer.Time,
				Timeout:           config.Instance().GrpcServer.Timeout,
			}),
		),
	)

	a.publicCloser.Add(func() error {
		gracefulCtx, cancel := context.WithTimeout(context.Background(), config.Instance().Graceful.Timeout)
		defer cancel()

		// Synchronize Stop with Clay initialization, including a shutdown before
		// the first external request. The admin listener starts independently.
		client := &http.Client{
			Timeout: 100 * time.Millisecond,
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		}
		defer client.CloseIdleConnections()
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
	waiting:
		for {
			select {
			case <-serverReady:
				break waiting
			case <-gracefulCtx.Done():
				return fmt.Errorf("pricing: server did not finish initialization: %w", gracefulCtx.Err())
			case <-ticker.C:
				response, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/swagger.json", config.Instance().HttpServer.Port))
				if err == nil {
					response.Body.Close()
				}
			}
		}

		done := make(chan struct{})
		go func() {
			err := a.mainServer.Stop(gracefulCtx)
			if err != nil {
				slog.Error("Не удалось остановить основной сервер", "error", err)
			}
			close(done)
		}()

		select {
		case <-done:
			slog.Info("Основной сервер остановлен", "service", "pricing")
		case <-gracefulCtx.Done():
			err := fmt.Errorf("pricing: error while graceful shutdown server: %w", gracefulCtx.Err())
			// Do not call Stop again: Clay GracefulStop does not honor the context.
			// Returning lets main exit after closers complete.
			return fmt.Errorf("pricing: stopped: %w", err)
		}
		return nil
	})

	return nil
}

func (a *App) initAdminServer(ctx context.Context) error {
	// init admin listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", config.Instance().HttpServer.AdminPort))
	if err != nil {
		return fmt.Errorf("failed to init admin listener: %w", err)
	}

	a.adminListener = lis

	// init healthcheck for admin server
	if err = a.initHealthCheck(ctx); err != nil {
		return fmt.Errorf("failed to init healthcheck: %w", err)
	}

	a.adminMux = chi.NewMux()

	a.adminMux.Mount("/debug", chimw.Profiler())

	// register healthcheck
	a.adminMux.HandleFunc(healthcheck.LivenessPath, a.healthCheck.LiveEndpoint)
	a.adminMux.HandleFunc(healthcheck.ReadinessPath, a.healthCheck.ReadyEndpoint)
	a.adminMux.HandleFunc(healthcheck.StartupPath, a.healthCheck.StartupEndpoint)

	return nil
}

func (a *App) initGrpcConn(_ context.Context) error {
	for srv, address := range config.Instance().Targets {
		var err error

		conn, err := grpc.NewClient(address,
			grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			return fmt.Errorf(
				"не удалось инициализировать grpc соединение к %s : %s",
				srv,
				err.Error(),
			)
		}

		a.grpcConn[srv] = conn
		closer.Add(conn.Close)
	}
	return nil
}

func (a *App) initHealthCheck(_ context.Context) error {
	a.healthCheck = healthcheck.NewHandler()

	// поверяю, что нет утечки горутин на старте (как пример)
	a.healthCheck.AddLivenessCheck("goroutines", func() error {
		if runtime.NumGoroutine() < 1000 {
			return nil
		}
		return fmt.Errorf("Превышено допустимое число горутин")
	})

	a.healthCheck.AddStartupCheck("started", func() error {
		if atomic.LoadInt32(&a.started) != 0 {
			return nil
		}
		return fmt.Errorf("Приложение ещё не запущено")
	})

	// readiness - т.к. я уже проинициализировал все компоненты
	a.healthCheck.AddReadinessCheck("started", func() error {
		if atomic.LoadInt32(&a.started) != 0 {
			return nil
		}
		return fmt.Errorf("Приложение ещё не запущено")
	})

	a.healthCheck.AddReadinessCheck("termination", func() error {
		if atomic.LoadInt32(&a.terminated) == 0 {
			return nil
		}
		return fmt.Errorf("Приложение останавливается")
	})

	a.publicCloser.Add(func() error {
		slog.Info(
			"Получен сигнал остановки",
			"service",
			"pricing",
			"timeout",
			config.Instance().Graceful.Timeout,
		)

		atomic.StoreInt32(&a.terminated, 1)
		return nil
	})
	return nil
}

func (a *App) initPostgres(ctx context.Context) error {
	pool, err := postgres.Pool(ctx, config.Instance().PostgresDSN)
	if err != nil {
		return fmt.Errorf("[POSTGRES] Не удалось инициализировать pool: %w", err)
	}
	a.pool = pool
	return nil
}

func (a *App) initStorage(context.Context) error {
	a.storage = storage.NewRegistry(a.pool)
	return nil
}

func (a *App) initUsecases(context.Context) error {
	a.usecases = usecase.NewRegistry(a.storage, a.gateway)
	return nil
}

func (a *App) initGateway(context.Context) error {
	a.gateway = gateway.NewRegistry(a.grpcConn)
	return nil
}

func (a *App) initInbox(_ context.Context) error {
	a.inbox = inbox.NewInbox(a.pool, config.Instance().Inbox.Config)
	return nil
}

func (a *App) initInboxWorkers(ctx context.Context) error {
	for _, mode := range []inbox.Mode{
		inbox.ModeNormal,
		inbox.ModeError,
	} {
		interval := config.Instance().Inbox.Config.NormalMessagesPollInterval
		if mode == inbox.ModeError {
			interval = config.Instance().Inbox.Config.ErrorMessagesPollInterval
		}
		background := worker.NewWorker(
			ctx,
			a.inbox.HandlePendingMessages(mode),
			func(context.Context) time.Duration {
				return interval
			},
			func(context.Context) int {
				return 1
			},
		)
		a.inboxWorkers = append(a.inboxWorkers, background)
		a.publicCloser.Add(background.Stop)
	}
	return nil
}

func (a *App) initMessageBus(_ context.Context) error {
	a.messageBus = messagebus.NewRegistry(a.inbox)
	a.publicCloser.Add(a.messageBus.Stop)
	return nil
}

func (a *App) initHandlers(context.Context) error {
	a.eventRouter = event_router.NewEventRouter[string, []byte]()
	a.eventRouter.RegisterAll(
		rental_reserved.NewHandler(a.usecases.ReserveTrip),
		rental_started.NewHandler(a.usecases.StartTrip),
		rental_completed.NewHandler(a.usecases.CompleteTrip),
		rental_canceled.NewHandler(a.usecases.CancelTrip),
	)
	a.inbox.RegisterHandler(rental_events.NewHandler(
		config.RentalEventsTopic,
		config.Instance().InboxConfig(config.RentalEventsTopic).BatchSize,
		a.eventRouter,
	))
	return nil
}
