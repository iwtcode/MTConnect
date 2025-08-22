package app

import (
	"MTConnect/internal/adapters/handlers"
	"MTConnect/internal/adapters/producers"
	"MTConnect/internal/adapters/repositories/datastore"
	"MTConnect/internal/config"
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
		// Регистрируем все модули приложения
		ConfigModule,
		RepositoryModule,
		ProducerModule, // Новый модуль для Kafka
		ServiceModule,
		UsecaseModule,
		HttpServerModule,
	)
}

// --- Модули FX ---

var ConfigModule = fx.Module("config_module",
	fx.Provide(
		// Загрузчик конфигурации
		config.LoadConfiguration,
	),
)

var RepositoryModule = fx.Module("repository_module",
	fx.Provide(
		// Предоставляем DataStore как реализацию интерфейса Repository
		func(ds interfaces.DataStoreRepository) interfaces.Repository {
			return struct{ interfaces.DataStoreRepository }{ds}
		},
		// Конструктор для нашего in-memory хранилища
		datastore.NewDataStore,
	),
)

var ProducerModule = fx.Module("producer_module",
	fx.Provide(
		// Продюсер для отправки данных (например, в Kafka)
		producers.NewKafkaProducer,
	),
)

var ServiceModule = fx.Module("service_module",
	fx.Provide(
		// Сервис для опроса MTConnect эндпоинтов
		services.NewPollingService,
	),
)

var UsecaseModule = fx.Module("usecases_module",
	fx.Provide(
		// Конструктор для бизнес-логики (use cases)
		usecases.NewUsecases,
	),
)

var HttpServerModule = fx.Module("http_server_module",
	fx.Provide(
		// Обработчики HTTP-запросов
		handlers.NewHandler,
		// Роутер
		handlers.ProvideRouter,
	),
	// Запускаем сервер при старте приложения
	fx.Invoke(InvokeHttpServer, InvokeGracefulShutdown),
)

// InvokeHttpServer запускает HTTP-сервер
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

// InvokeGracefulShutdown обеспечивает корректное завершение работы сервисов
func InvokeGracefulShutdown(lc fx.Lifecycle, poller interfaces.PollingService, producer interfaces.DataProducer) {
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
