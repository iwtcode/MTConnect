<div align="center">

# MTConnect Service

![alt text](https://img.shields.io/badge/Go-1.19+-00ADD8?logo=go)
![alt text](https://img.shields.io/badge/MTConnect-Compatible-blue)
![alt text](https://img.shields.io/badge/Apache%20Kafka-Integrated-blue?logo=apachekafka)
![alt text](https://img.shields.io/badge/PostgreSQL-Supported-336791?logo=postgresql)
![alt text](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)
![alt text](https://img.shields.io/badge/License-MIT-green)

*Сервис для сбора данных по протоколу MTConnect, отправки в Apache Kafka и управления через REST API*

</div>

### ✨ Ключевые возможности
- 🚀 **Потоковая передача в Kafka**: Все данные со станков в реальном времени отправляются в топик Apache Kafka для дальнейшей обработки и аналитики
- 🕹️ **Управляемый опрос**: Запускайте и останавливайте мониторинг для каждого станка индивидуально через REST API с настраиваемым интервалом
- 💾 **Персистентность**: Состояния подключений и опроса сохраняются в базе данных PostgreSQL, что позволяет автоматически восстанавливать их после перезапуска сервиса.
- 🌐 **REST API**: Удобный HTTP API для получения актуальных данных, проверки доступности станков и управления процессами опроса
- 🐳 **Простота развертывания**: Готовая конфигурация docker-compose.yml для быстрого запуска Apache Kafka и сопутствующих сервисов
- 🎛️ **Веб-интерфейс для Kafka**: Встроенный Kafka UI для удобного просмотра топиков и сообщений
- 🔧 **Универсальность**: Автоматическое извлечение и кэширование метаинформации из /probe для корректной интерпретации данных с различных станков

## 🏗️ Архитектура

```
┌─────────────────┐      ┌─────────────────┐      ┌──────────────────┐
│   Управляющий   ├─────▸│     Сервис      │◂─────┤    MTConnect     │
│    REST API     │      │    MTConnect    │      │    Endpoints     │
│   (Gin-Gonic)   │      │    (Go App)     │      │   (XML-данные)   │
└─────────────────┘      └───────┬───┬─────┘      └──────────────────┘
        ▴                        │   │      (Polling)
        │                        │   └─────────────────────┐
        │                        ▾                         ▾
┌───────┴─────────┐      ┌─────────────────┐      ┌──────────────────┐
│  Пользователь / │      │  PostgreSQL     │      │   Apache Kafka   │
│     Система     │      │  (Состояния     │      │   (Потоковая     │
│  (Управление)   │      │  подключений)   │      │   обработка)     │
└─────────────────┘      └─────────────────┘      └──────────────────┘
```

## 📦 Установка

1️⃣ **Клонирование репозитория**

```bash
git clone https://github.com/iwtcode/MTConnect.git
cd MTConnect
```

2️⃣ **Конфигурация приложения**

Откройте файл .env и при необходимости измените его

```dotenv
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=1234
DB_NAME=mtconnect_db

# App
APP_PORT=8080
GIN_MODE=debug

# Kafka
KAFKA_BROKER=localhost:9092
KAFKA_TOPIC=mtconnect_data

# Logger
LOGGER_ENABLE=true
LOGGER_LOGS_DIR=./logs
LOGGER_LOG_LEVEL=DEBUG
LOGGER_SAVING_DAYS=7
```

3️⃣ **Установка cppagent**
```bash
https://github.com/mtconnect/cppagent/releases/tag/v2.5.0.11
```

4️⃣ **Запуск симуляторов станков**

Скопируйте файл `pkg\agent\random_simulator.rb` в `agent-2.5.0.11-win64\simulator`
```bash
cd agent-2.5.0.11-win64\simulator
ruby random_simulator.rb
```

5️⃣ **Запуск MTConnect agent**
```bash
cd agent-2.5.0.11-win64\bin
agent.exe debug ..\demo\agent\agent.cfg
```

6️⃣ **Запуск Apache Kafka**

```bash
cd MTConnect
docker-compose up
```

После запуска [Веб-интерфейс Kafka](http://localhost:8081)

Либо просмотреть сообщения сервера можно в реальном времени командой:<br>
`docker-compose exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic mtconnect_data`

7️⃣ **Запуск приложения**

```
# Windows
./build/windows_mtc.exe

# Linux
./build/linux_mtc

# MacOS
./build/macos_mtc

# Golang
go run cmd/app/main.go
```

## 🔌 API

## Создание подключения

```http
POST /connect
```

```bash
curl -X POST http://localhost:8080/api/v1/connect \
-H "Content-Type: application/json" \
-d '{
    "endpoint_url": "http://localhost:5001/Mazak",
    "model": "Mazak VRX C600"
}'
```

```json
{
    "status": "ok",
    "connection_info": {
        "session_id": "870c5240-de93-4584-b411-37aa915cbc1d",
        "config": {
            "endpoint_url": "http://localhost:5001/Mazak",
            "model": "Mazak VRX C600",
            "manufacturer": "Mazak"
        },
        "created_at": "2025-08-28T12:19:21.2303802+03:00",
        "last_used": "2025-08-28T12:19:21.2303802+03:00",
        "use_count": 1,
        "is_healthy": true
    }
}
```

## Получение списка активных подключений

```http
GET /connect
```

```bash
curl http://localhost:8080/api/v1/connect
```

```json
{
    "status": "ok",
    "pool_size": 2,
    "connections": [
        {
            "session_id": "bba81cc3-ad26-4e0b-9336-a8cc8bf54238",
            "config": {
                "endpoint_url": "http://localhost:5001/OKUMA",
                "model": "Okuma MTConnect Adapter",
                "manufacturer": "OKUMA"
            },
            "created_at": "2025-08-28T12:19:00.4920414+03:00",
            "last_used": "2025-08-28T12:19:00.4920414+03:00",
            "use_count": 1,
            "is_healthy": true
        },
        {
            "session_id": "870c5240-de93-4584-b411-37aa915cbc1d",
            "config": {
                "endpoint_url": "http://localhost:5001/Mazak",
                "model": "Mazak VRX C600",
                "manufacturer": "Mazak"
            },
            "created_at": "2025-08-28T12:19:21.2303802+03:00",
            "last_used": "2025-08-28T12:19:21.2303802+03:00",
            "use_count": 1,
            "is_healthy": true
        }
    ]
}
```

## Проверка состояния подключения конкретного станка

```http
POST /connect/check
```

```bash
curl -X POST http://localhost:8080/api/v1/connect/check \
-H "Content-Type: application/json" \
-d '{
    "session_id": "870c5240-de93-4584-b411-37aa915cbc1d"
}'
```

```json
{
    "status": "healthy",
    "connection_info": {
        "session_id": "870c5240-de93-4584-b411-37aa915cbc1d",
        "config": {
            "endpoint_url": "http://localhost:5001/Mazak",
            "model": "Mazak VRX C600",
            "manufacturer": "Mazak"
        },
        "created_at": "2025-08-28T12:19:21.2303802+03:00",
        "last_used": "2025-08-28T12:22:33.978361+03:00",
        "use_count": 2,
        "is_healthy": true
    }
}
```

## Запуск сбора данных

```http
POST /polling/start
```

```bash
curl -X POST http://localhost:8080/api/v1/polling/start \
-H "Content-Type: application/json" \
-d '{
    "session_id": "870c5240-de93-4584-b411-37aa915cbc1d",
    "interval": 1000
}'
```

```json
{
    "status": "ok",
    "message": "Polling started for session 870c5240-de93-4584-b411-37aa915cbc1d"
}
```

## Остановка сбора данных

```http
POST /polling/stop
```

```bash
curl -X POST http://localhost:8080/api/v1/polling/stop \
-H "Content-Type: application/json" \
-d '{
    "session_id": "870c5240-de93-4584-b411-37aa915cbc1d"
}'
```

```json
{
    "status": "ok",
    "message": "Polling stopped for session 870c5240-de93-4584-b411-37aa915cbc1d"
}
```

## Удаление подключения

```http
DELETE /connect
```

```bash
curl -X DELETE http://localhost:8080/api/v1/connect \
-H "Content-Type: application/json" \
-d '{
    "session_id": "870c5240-de93-4584-b411-37aa915cbc1d"
}'
```

```json
{
    "status": "ok",
    "message": "Session 870c5240-de93-4584-b411-37aa915cbc1d disconnected successfully"
}
```

## 🔧 Структура проекта

```
MTConnect/
├── cmd/app/                      # Главная точка входа приложения (main.go)
├── internal/
│   ├── app/                      # Сборка и запуск приложения с помощью Fx для DI
│   ├── config/                   # Логика загрузки конфигурации из .env
│   ├── adapters/
│   │   ├── handlers/             # Обработчики HTTP-запросов (слой API на Gin)
│   │   └── repositories/         # Реализации репозиториев (PostgreSQL)
│   ├── domain/                   # Основные бизнес-сущности (entities) и модели (models)
│   ├── interfaces/               # Go-интерфейсы для всех слоев (контракты)
│   ├── services/                 
│   │   ├── kafka/                # Продюсер для Apache Kafka
│   │   └── mtconnect_service/    # Основная бизнес-логика: управление подключениями, опрос, парсинг
│   └── usecases/                 # Сценарии использования, связывающие API и сервисный слой
├── pkg/
│   ├── client/                   # Клиентская библиотека для API
│   └── models/                   # Публичные модели для использования в клиенте
├── tools/
│   └── build/                    # Скрипт для сборки исполняемых файлов
├── build/                        # Папка с готовыми исполняемыми файлами
├── .env                          # Файл конфигурации
├── docker-compose.yml            # Файл для запуска Kafka и Kafka-UI
├── LICENSE
└── README.md
```

## 🆘 Поддержка

- 🐛 [Создайте issue](https://github.com/iwtcode/github.com/iwtcode/MTConnect/issues)
- 📧 Напишите на email: iwtcode@gmail.com

## 📝 Лицензия

Проект распространяется под [лицензией MIT](LICENSE)

Copyright (c) 2025 iwtcode