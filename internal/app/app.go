package app

import (
	"context"
	"net/http"
	"time"

	"github.com/iwtcode/MTConnect/internal/adapters/handlers"
	"github.com/iwtcode/MTConnect/internal/adapters/repositories/postgres"
	"github.com/iwtcode/MTConnect/internal/config"
	"github.com/iwtcode/MTConnect/internal/domain/entities"
	"github.com/iwtcode/MTConnect/internal/interfaces"
	"github.com/iwtcode/MTConnect/internal/middleware/logging"
	"github.com/iwtcode/MTConnect/internal/middleware/swagger"
	"github.com/iwtcode/MTConnect/internal/services/kafka"
	"github.com/iwtcode/MTConnect/internal/services/mtconnect_service"
	"github.com/iwtcode/MTConnect/internal/usecases"

	"go.uber.org/fx"
)

// New создает новый экземпляр fx.App
func New() *fx.App {
	return fx.New(
		ConfigModule,
		LoggingModule,
		RepositoryModule,
		ProducerModule,
		ServiceModule,
		UsecaseModule,
		HttpServerModule,
		fx.Invoke(InvokeRestoreConnections),
		fx.Invoke(InvokeBackgroundHealthChecker),
	)
}

// --- Модули FX ---

var ConfigModule = fx.Module("config_module",
	fx.Provide(config.LoadConfiguration),
)

// Модуль логгера
func ProvideLogger(cfg *config.AppConfig) *logging.Logger {
	loggerCfg := &logging.Config{
		Enabled:    cfg.Logging.Enable,
		Level:      cfg.Logging.Level,
		LogsDir:    cfg.Logging.LogsDir,
		SavingDays: uint(cfg.Logging.SavingDays),
	}
	return logging.NewLogger(loggerCfg, "MTConnectApp")
}

var LoggingModule = fx.Module("logging_module",
	fx.Provide(ProvideLogger),
)

var RepositoryModule = fx.Module("repository_module",
	fx.Provide(
		postgres.NewRepository,
	),
)

var ProducerModule = fx.Module("producer_module",
	fx.Provide(kafka.NewKafkaProducer),
)

var ServiceModule = fx.Module("service_module",
	fx.Provide(mtconnect_service.NewMTConnectService),
)

var UsecaseModule = fx.Module("usecases_module",
	fx.Provide(usecases.NewUsecases),
)

// Конфигуратор Swagger
func NewSwaggerConfig() *swagger.Config {
	return &swagger.Config{
		Enabled: true,
		Path:    "/swagger",
	}
}

var HttpServerModule = fx.Module("http_server_module",
	fx.Provide(
		NewSwaggerConfig,
		handlers.NewHandler,
		handlers.ProvideRouter,
	),
	fx.Invoke(InvokeHttpServer, InvokeGracefulShutdown),
)

// InvokeRestoreConnections восстанавливает подключения и опросы при старте приложения.
func InvokeRestoreConnections(lc fx.Lifecycle, mtconnectSvc interfaces.Usecases, dbRepo interfaces.CncMachineRepository, logger *logging.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Restoring connections from the database...")
			machines, err := dbRepo.GetAll()
			if err != nil {
				logger.Error("Failed to get machine list from DB", "error", err)
				return nil
			}

			if len(machines) == 0 {
				logger.Info("No saved connections found to restore.")
				return nil
			}

			for _, machine := range machines {
				logger.Info("Attempting to restore connection", "sessionID", machine.SessionID, "model", machine.Model, "endpoint", machine.EndpointURL)

				// Эта функция теперь не возвращает ошибку, а всегда добавляет в пул,
				// помечая нездоровые подключения как IsHealthy: false.
				connInfo, _ := mtconnectSvc.RestoreConnection(machine)

				if connInfo.IsHealthy {
					logger.Info("Connection restored successfully in pool", "sessionID", machine.SessionID)
				} else {
					logger.Warn("Connection restored in pool but is unhealthy. Will retry in background.", "sessionID", machine.SessionID)
				}

				if machine.Status == entities.StatusPolled {
					if machine.Interval > 0 {
						interval := time.Duration(machine.Interval) * time.Millisecond
						logger.Info("Starting restored polling", "sessionID", machine.SessionID, "interval", interval)
						// Попытка запуска опроса. Если соединение нездорово, опрос не запустится,
						// но и не вызовет критической ошибки.
						if err := mtconnectSvc.StartPolling(connInfo.SessionID, interval); err != nil {
							logger.Warn("Failed to start polling for restored session (it may be unhealthy)", "sessionID", machine.SessionID, "error", err)
						}
					} else {
						logger.Warn("Status for session is 'polled' but interval is not set. Polling not started.", "sessionID", machine.SessionID)
					}
				}
			}
			return nil
		},
	})
}

// InvokeBackgroundHealthChecker запускает фоновую задачу для периодической проверки состояния всех подключений.
func InvokeBackgroundHealthChecker(lc fx.Lifecycle, mtconnectSvc interfaces.Usecases, logger *logging.Logger) {
	const checkInterval = 1 * time.Second
	var done chan bool

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting background health checker", "interval", checkInterval)
			done = make(chan bool)
			go func() {
				ticker := time.NewTicker(checkInterval)
				defer ticker.Stop()

				for {
					select {
					case <-done:
						logger.Info("Background health checker stopped.")
						return
					case <-ticker.C:
						connections := mtconnectSvc.GetAllConnections()
						if len(connections) > 0 {
							for _, conn := range connections {
								_, _ = mtconnectSvc.CheckConnection(conn.SessionID)
							}
						}
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping background health checker...")
			if done != nil {
				close(done)
			}
			return nil
		},
	})
}

// InvokeHttpServer запускает HTTP-сервер.
func InvokeHttpServer(lc fx.Lifecycle, cfg *config.AppConfig, h http.Handler, logger *logging.Logger) {
	serverAddr := ":" + cfg.ServerPort
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Server is starting", "address", serverAddr)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("Failed to start server", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping HTTP server...")
			return server.Shutdown(ctx)
		},
	})
}

// InvokeGracefulShutdown обеспечивает корректное завершение работы сервисов.
func InvokeGracefulShutdown(lc fx.Lifecycle, mtconnectSvc interfaces.MTConnectService, producer interfaces.KafkaService, logger *logging.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Gracefully shutting down services...")

			if err := producer.Close(); err != nil {
				logger.Error("Error closing Kafka producer", "error", err)
			}
			logger.Info("All services stopped successfully.")
			return nil
		},
	})
}
