package app

import (
	"MTConnect/internal/adapters/handlers"
	"MTConnect/internal/adapters/producers"
	"MTConnect/internal/adapters/repositories/datastore"
	"MTConnect/internal/adapters/repositories/postgres" // Изменено
	"MTConnect/internal/config"
	"MTConnect/internal/domain/entities" // Добавлено
	"MTConnect/internal/interfaces"
	"MTConnect/internal/services"
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
		RepositoryModule,
		ProducerModule,
		ServiceModule,
		UsecaseModule,
		HttpServerModule,
		// Добавляем вызов новой функции
		fx.Invoke(InvokeRestoreConnections),
	)
}

// --- Модули FX ---

var ConfigModule = fx.Module("config_module",
	fx.Provide(config.LoadConfiguration),
)

// --- ИЗМЕНЕНИЕ: RepositoryModule теперь работает с PostgreSQL ---
var RepositoryModule = fx.Module("repository_module",
	fx.Provide(
		// Провайдер для in-memory хранилища данных (остается)
		datastore.NewDataStore,
		// Провайдер для репозитория БД
		postgres.NewRepository,
	),
	// Приводим конкретные реализации к общему интерфейсу Repository
	fx.Provide(func(ds interfaces.DataStoreRepository, cncRepo interfaces.CncMachineRepository) interfaces.Repository {
		return struct {
			interfaces.DataStoreRepository
			interfaces.CncMachineRepository
		}{ds, cncRepo}
	}),
)

var ProducerModule = fx.Module("producer_module",
	fx.Provide(producers.NewKafkaProducer),
)

var ServiceModule = fx.Module("service_module",
	fx.Provide(
		services.NewPollingService,
		// Передаем новый репозиторий в сервис
		func(pollSvc interfaces.PollingService, dbRepo interfaces.CncMachineRepository) interfaces.ConnectionService {
			return services.NewConnectionService(pollSvc, dbRepo)
		},
	),
)

var UsecaseModule = fx.Module("usecases_module",
	fx.Provide(usecases.NewUsecases),
)

var HttpServerModule = fx.Module("http_server_module",
	fx.Provide(
		handlers.NewHandler,
		handlers.ProvideRouter,
	),
	fx.Invoke(InvokeHttpServer, InvokeGracefulShutdown),
)

// --- НОВАЯ ФУНКЦИЯ: Восстановление подключений при старте ---
func InvokeRestoreConnections(lc fx.Lifecycle, connSvc interfaces.ConnectionService, dbRepo interfaces.CncMachineRepository) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Восстановление активных подключений из базы данных...")
			machines, err := dbRepo.GetAllByStatus(entities.StatusConnected)
			if err != nil {
				log.Printf("ОШИБКА: не удалось получить список активных станков из БД: %v", err)
				return nil // Не блокируем запуск приложения из-за этой ошибки
			}

			if len(machines) == 0 {
				log.Println("Не найдено активных подключений для восстановления.")
				return nil
			}

			for _, machine := range machines {
				log.Printf("Попытка восстановить подключение для модели '%s' на '%s'", machine.Model, machine.EndpointURL)
				req := entities.ConnectionRequest{
					EndpointURL:  machine.EndpointURL,
					Model:        machine.Model,
					Manufacturer: machine.Manufacturer,
				}
				if _, err := connSvc.CreateConnection(req); err != nil {
					log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось восстановить подключение для сессии %s: %v", machine.SessionID, err)
					// Обновляем статус в БД, чтобы не пытаться восстановить в следующий раз
					if err := dbRepo.UpdateStatus(machine.SessionID, entities.StatusDisconnected); err != nil {
						log.Printf("ОШИБКА: не удалось обновить статус для сессии %s в БД: %v", machine.SessionID, err)
					}
				} else {
					log.Printf("Подключение для модели '%s' успешно восстановлено.", machine.Model)
				}
			}
			return nil
		},
	})
}

// InvokeHttpServer запускает HTTP-сервер
func InvokeHttpServer(lc fx.Lifecycle, cfg *config.AppConfig, h http.Handler) {
	// ... (код функции остается без изменений)
	serverAddr := ":" + cfg.ServerPort
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("Сервер запущен на http://localhost%s", serverAddr)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalf("Не удалось запустить сервер: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Остановка HTTP-сервера...")
			return server.Shutdown(ctx)
		},
	})
}

// InvokeGracefulShutdown обеспечивает корректное завершение работы сервисов
func InvokeGracefulShutdown(lc fx.Lifecycle, poller interfaces.PollingService, producer interfaces.DataProducer) {
	// ... (код функции остается без изменений)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Println("Корректное завершение работы сервисов...")
			poller.StopAllPolling()
			if err := producer.Close(); err != nil {
				log.Printf("Ошибка при закрытии Kafka продюсера: %v", err)
				return err
			}
			log.Println("Все сервисы успешно остановлены.")
			return nil
		},
	})
}
