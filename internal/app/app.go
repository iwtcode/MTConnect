package app

import (
	"MTConnect/internal/adapters/handlers"
	"MTConnect/internal/adapters/repositories/datastore"
	"MTConnect/internal/adapters/repositories/postgres"
	"MTConnect/internal/config"
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/interfaces"
	"MTConnect/internal/middleware/logging"
	"MTConnect/internal/middleware/swagger"
	"MTConnect/internal/services/kafka"
	"MTConnect/internal/services/mtconnect_service"
	"MTConnect/internal/usecases"
	"context"
	"log"
	"net/http"
	"time"

	"go.uber.org/fx"
)

// New создает новый экземпляр fx.App
func New() *fx.App {
	return fx.New(
		ConfigModule,
		LoggingModule, // Добавлен модуль логгера
		RepositoryModule,
		ProducerModule,
		ServiceModule,
		UsecaseModule,
		HttpServerModule,
		fx.Invoke(InvokeRestoreConnections),
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
		datastore.NewDataStore,
		postgres.NewRepository,
	),
	fx.Provide(func(ds interfaces.DataStoreRepository, cncRepo interfaces.CncMachineRepository) interfaces.Repository {
		type repositoryImpl struct {
			interfaces.DataStoreRepository
			interfaces.CncMachineRepository
		}
		return repositoryImpl{ds, cncRepo}
	}),
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
		NewSwaggerConfig, // Добавлен провайдер Swagger
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
				if machine.Status == entities.StatusDisconnected {
					continue
				}

				logger.Info("Attempting to restore connection", "sessionID", machine.SessionID, "model", machine.Model, "endpoint", machine.EndpointURL)

				connInfo, err := mtconnectSvc.RestoreConnection(machine)
				if err != nil {
					logger.Warn("Failed to restore connection", "sessionID", machine.SessionID, "error", err)
					if err := dbRepo.UpdateStatus(machine.SessionID, entities.StatusDisconnected); err != nil {
						logger.Error("Failed to update status in DB for session", "sessionID", machine.SessionID, "error", err)
					}
					continue
				}

				logger.Info("Connection restored successfully in pool", "sessionID", machine.SessionID)

				if machine.Status == entities.StatusPolled {
					if machine.Interval > 0 {
						interval := time.Duration(machine.Interval) * time.Millisecond
						logger.Info("Starting restored polling", "sessionID", machine.SessionID, "interval", interval)
						if err := mtconnectSvc.StartPolling(connInfo.SessionID, interval); err != nil {
							logger.Warn("Failed to start polling for session", "sessionID", machine.SessionID, "error", err)
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
					log.Fatalf("Failed to start server: %v", err)
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

			connections := mtconnectSvc.GetAllConnections()
			for _, conn := range connections {
				if mtconnectSvc.IsPollingActive(conn.SessionID) {
					logger.Info("Stopping polling on shutdown", "sessionID", conn.SessionID)
					if err := mtconnectSvc.StopPolling(conn.SessionID); err != nil {
						logger.Error("Error stopping polling", "sessionID", conn.SessionID, "error", err)
					}
				}
			}

			if err := producer.Close(); err != nil {
				logger.Error("Error closing Kafka producer", "error", err)
			}
			logger.Info("All services stopped successfully.")
			return nil
		},
	})
}
