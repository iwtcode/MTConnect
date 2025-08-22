package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"MTConnect/internal/adapters/handlers"
	"MTConnect/internal/adapters/repositories/datastore"
	"MTConnect/internal/config"
	"MTConnect/internal/interfaces"
	"MTConnect/internal/services"
	"MTConnect/internal/usecases"

	"go.uber.org/fx"
)

// New создает новый экземпляр fx.App
func New() *fx.App {
	return fx.New(
		// Регистрируем все модули приложения
		ConfigModule,
		RepositoryModule,
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
	// Запускаем сервер и сервис опроса при старте приложения
	fx.Invoke(InvokeHttpServer, InvokePollingService),
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

// InvokePollingService запускает фоновый опрос эндпоинтов
func InvokePollingService(lc fx.Lifecycle, poller interfaces.PollingService) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Запуск сервиса опроса MTConnect эндпоинтов...")
			go poller.StartPolling()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Остановка сервиса опроса.")
			poller.StopPolling()
			return nil
		},
	})
}
