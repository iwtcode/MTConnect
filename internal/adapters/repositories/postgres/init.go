package postgres

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/iwtcode/MTConnect/internal/adapters/repositories/postgres/cnc_machine"
	"github.com/iwtcode/MTConnect/internal/config"
	"github.com/iwtcode/MTConnect/internal/domain/entities"
	"github.com/iwtcode/MTConnect/internal/interfaces"
	"github.com/iwtcode/MTConnect/internal/middleware/logging"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repository struct {
	interfaces.CncMachineRepository
}

func NewRepository(cfg *config.AppConfig, appLogger *logging.Logger) (interfaces.CncMachineRepository, error) {
	// --- НАЧАЛО: Логика автоматического создания БД ---

	// Шаг 1: Подключение к служебной БД 'postgres' для проверки и создания целевой БД
	dsnPostgres := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Port,
	)

	db, err := gorm.Open(postgres.Open(dsnPostgres), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Временное отключение логов
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к служебной БД 'postgres': %w", err)
	}

	// Шаг 2: Проверка существования нужной БД
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = ?)"
	if err := db.Raw(query, cfg.Database.DBName).Scan(&exists).Error; err != nil {
		return nil, fmt.Errorf("не удалось проверить существование БД '%s': %w", cfg.Database.DBName, err)
	}

	// Шаг 3: Если БД не существует, создаем ее
	if !exists {
		appLogger.Info("Database not found. Creating...", "db_name", cfg.Database.DBName)
		// ВАЖНО: Имя БД здесь не параметризуется, так как CREATE DATABASE не поддерживает плейсхолдеры.
		// Это безопасно, так как имя берется из файла конфигурации.
		createDbQuery := fmt.Sprintf("CREATE DATABASE %s", cfg.Database.DBName)
		if err := db.Exec(createDbQuery).Error; err != nil {
			return nil, fmt.Errorf("не удалось создать БД '%s': %w", cfg.Database.DBName, err)
		}
		appLogger.Info("Database created successfully.", "db_name", cfg.Database.DBName)
	} else {
		appLogger.Info("Database already exists.", "db_name", cfg.Database.DBName)
	}

	// Закрываем соединение со служебной БД
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	// --- КОНЕЦ: Логика автоматического создания БД ---

	// Шаг 4: Основное подключение к целевой базе данных
	dsnApp := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
	)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	appDb, err := gorm.Open(postgres.Open(dsnApp), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных '%s': %w", cfg.Database.DBName, err)
	}

	if err := autoMigrate(appDb); err != nil {
		return nil, fmt.Errorf("ошибка выполнения автомиграций: %w", err)
	}

	return &Repository{
		CncMachineRepository: cnc_machine.NewCncMachineRepository(appDb),
	}, nil
}

func autoMigrate(db *gorm.DB) error {
	// AutoMigrate безопасно создает таблицу, если она не существует,
	// и добавляет новые колонки, если они появились в модели.
	return db.AutoMigrate(&entities.CncMachine{})
}
