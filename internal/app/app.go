package app

import (
	"MTConnect/internal/adapters/handlers"
	"MTConnect/internal/adapters/repositories/datastore"
	"MTConnect/internal/adapters/repositories/postgres"
	"MTConnect/internal/config"
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/interfaces"
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

var HttpServerModule = fx.Module("http_server_module",
	fx.Provide(
		handlers.NewHandler,
		handlers.ProvideRouter,
	),
	fx.Invoke(InvokeHttpServer, InvokeGracefulShutdown),
)

// InvokeRestoreConnections восстанавливает подключения и опросы при старте приложения.
func InvokeRestoreConnections(lc fx.Lifecycle, mtconnectSvc interfaces.Usecases, dbRepo interfaces.CncMachineRepository) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Восстановление подключений из базы данных...")
			machines, err := dbRepo.GetAll()
			if err != nil {
				log.Printf("ОШИБКА: не удалось получить список станков из БД: %v", err)
				return nil
			}

			if len(machines) == 0 {
				log.Println("Не найдено сохраненных подключений для восстановления.")
				return nil
			}

			for _, machine := range machines {
				if machine.Status == entities.StatusDisconnected {
					continue
				}

				log.Printf("Попытка восстановить подключение для сессии '%s' (%s на %s)", machine.SessionID, machine.Model, machine.EndpointURL)

				// ИСПРАВЛЕНО: Вызываем новый метод RestoreConnection вместо CreateConnection
				connInfo, err := mtconnectSvc.RestoreConnection(machine)
				if err != nil {
					log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось восстановить подключение для сессии %s: %v", machine.SessionID, err)
					if err := dbRepo.UpdateStatus(machine.SessionID, entities.StatusDisconnected); err != nil {
						log.Printf("ОШИБКА: не удалось обновить статус для сессии %s в БД: %v", machine.SessionID, err)
					}
					continue
				}

				log.Printf("Подключение для сессии '%s' успешно восстановлено в пуле.", machine.SessionID)

				if machine.Status == entities.StatusPolled {
					if machine.Interval > 0 {
						interval := time.Duration(machine.Interval) * time.Millisecond
						log.Printf("Запуск восстановленного опроса для сессии '%s' с интервалом %v.", machine.SessionID, interval)
						// Передаем восстановленный connInfo, который уже есть в пуле
						if err := mtconnectSvc.StartPolling(connInfo.SessionID, interval); err != nil {
							log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось запустить опрос для сессии %s: %v", machine.SessionID, err)
						}
					} else {
						log.Printf("ПРЕДУПРЕЖДЕНИЕ: статус для сессии %s - 'polled', но интервал не задан. Опрос не запущен.", machine.SessionID)
					}
				}
			}
			return nil
		},
	})
}

// InvokeHttpServer запускает HTTP-сервер.
func InvokeHttpServer(lc fx.Lifecycle, cfg *config.AppConfig, h http.Handler) {
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

// InvokeGracefulShutdown обеспечивает корректное завершение работы сервисов.
func InvokeGracefulShutdown(lc fx.Lifecycle, mtconnectSvc interfaces.MTConnectService, producer interfaces.KafkaService) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Println("Корректное завершение работы сервисов...")

			connections := mtconnectSvc.GetAllConnections()
			for _, conn := range connections {
				if mtconnectSvc.IsPollingActive(conn.SessionID) {
					log.Printf("Остановка опроса для сессии %s при завершении работы...", conn.SessionID)
					if err := mtconnectSvc.StopPolling(conn.SessionID); err != nil {
						log.Printf("Ошибка при остановке опроса для %s: %v", conn.SessionID, err)
					}
				}
			}

			if err := producer.Close(); err != nil {
				log.Printf("Ошибка при закрытии Kafka продюсера: %v", err)
			}
			log.Println("Все сервисы успешно остановлены.")
			return nil
		},
	})
}
