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

	"api-gateway/config"
	"api-gateway/internal/app/echo/v1"
	pricing "api-gateway/internal/app/pricing/v1"
	profile "api-gateway/internal/app/profile/v1"
	rental "api-gateway/internal/app/rental/v1"
	scooter "api-gateway/internal/app/scooter/v1"
	runtimeconfig "api-gateway/internal/config"
	"api-gateway/internal/pkg/closer"
	"api-gateway/internal/pkg/grpc/intercept"
	"api-gateway/internal/pkg/healthcheck"
	"api-gateway/internal/pkg/httptransport"
	echoV1 "api-gateway/internal/pkg/pb/api-gateway/echo/v1"
	pricingV1 "api-gateway/internal/pkg/pb/api-gateway/pricing/v1"
	profileV1 "api-gateway/internal/pkg/pb/api-gateway/profile/v1"
	rentalV1 "api-gateway/internal/pkg/pb/api-gateway/rental/v1"
	scooterV1 "api-gateway/internal/pkg/pb/api-gateway/scooter/v1"
	externalPricingV1 "api-gateway/internal/pkg/pb/external/pricing/pricing/v1"
	externalProfileV1 "api-gateway/internal/pkg/pb/external/profile/profile/v1"
	externalRentalV1 "api-gateway/internal/pkg/pb/external/rental/rental/v1"
	externalScooterV1 "api-gateway/internal/pkg/pb/external/scooter/scooter/v1"
	"api-gateway/internal/pkg/requestid"
	"api-gateway/internal/pkg/rtconfig/global"
)

func (a *App) initControllers(_ context.Context) error {
	a.controllers = []transport.ServiceDesc{
		echoV1.NewEchoServiceServiceDesc(echo.NewEchoService()),
		profileV1.NewProfileServiceServiceDesc(profile.NewProfileService(
			externalProfileV1.NewProfileServiceClient(a.grpcConn[config.ProfileService]),
		)),
		scooterV1.NewScooterServiceServiceDesc(scooter.NewScooterService(
			externalScooterV1.NewScooterServiceClient(a.grpcConn[config.ScooterService]),
		)),
		rentalV1.NewRentalServiceServiceDesc(rental.NewRentalService(
			externalRentalV1.NewRentalServiceClient(a.grpcConn[config.RentalService]),
		)),
		pricingV1.NewPricingServiceServiceDesc(pricing.NewPricingService(
			externalPricingV1.NewPricingServiceClient(a.grpcConn[config.PricingService]),
		)),
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
		server.WithRuntimeServeMuxOpts(httptransport.Options()...),
		server.WithGRPCUnaryMiddlewares(
			requestid.Interceptor(),
		),
		server.WithGRPCOpts(
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
				return fmt.Errorf("api-gateway: server did not finish initialization: %w", gracefulCtx.Err())
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
				slog.Error(
					"Ошибка остановки основного сервера",
					"service",
					"api-gateway",
					"error",
					err,
				)
			}
			close(done)
		}()

		select {
		case <-done:
			slog.Warn("api-gateway: main server gracefully stopped")
		case <-gracefulCtx.Done():
			err := fmt.Errorf("api-gateway: error while graceful shutdown server: %w", gracefulCtx.Err())
			// Do not call Stop again: Clay GracefulStop does not honor the context.
			// Returning lets main exit after closers complete.
			return fmt.Errorf("api-gateway: stopped: %w", err)
		}
		return nil
	})

	return nil
}

func (a *App) initRuntimeConfig(ctx context.Context) error {
	client, err := global.Init(ctx, config.Instance().RuntimeConfig, runtimeconfig.SchemaHash)
	if err != nil {
		return err
	}
	a.runtimeConfig.Store(client)
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
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(intercept.ProxyMetadata()),
		)

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
	a.healthCheck.AddReadinessCheck("runtime_config", func() error {
		return a.runtimeConfig.Load().CheckReady()
	})

	// поверяю, что нет утечки горутин на старте (как пример)
	a.healthCheck.AddLivenessCheck("goroutines", func() error {
		if runtime.NumGoroutine() < 1000 {
			return nil
		}
		return fmt.Errorf("application has too much running goroutines")
	})

	a.healthCheck.AddStartupCheck("started", func() error {
		if atomic.LoadInt32(&a.started) != 0 {
			return nil
		}
		return fmt.Errorf("application is not started yet")
	})

	// readiness - т.к. я уже проинициализировал все компоненты
	a.healthCheck.AddReadinessCheck("started", func() error {
		if atomic.LoadInt32(&a.started) != 0 {
			return nil
		}
		return fmt.Errorf("application is not statred yet")
	})

	a.healthCheck.AddReadinessCheck("termination", func() error {
		if atomic.LoadInt32(&a.terminated) == 0 {
			return nil
		}
		return fmt.Errorf("application is terminating now")
	})

	a.publicCloser.Add(func() error {
		slog.Info("Получен сигнал завершения", "service", "api-gateway",
			"graceful_timeout", config.Instance().Graceful.Timeout)

		atomic.StoreInt32(&a.terminated, 1)
		return nil
	})
	return nil
}
