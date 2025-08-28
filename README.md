<div align="center">

# MTConnect Streamer

![alt text](https://img.shields.io/badge/Go-1.19+-00ADD8?logo=go)
![alt text](https://img.shields.io/badge/MTConnect-Compatible-blue)
![alt text](https://img.shields.io/badge/Apache%20Kafka-Integrated-blue?logo=apachekafka)
![alt text](https://img.shields.io/badge/PostgreSQL-Supported-336791?logo=postgresql)
![alt text](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)
![alt text](https://img.shields.io/badge/License-MIT-green)

*Сервис для сбора данных по протоколу MTConnect, их отправки в Apache Kafka и управления через REST API*

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

3️⃣ **Запуск Apache Kafka**

```bash
docker-compose up
```

После запуска [Веб-интерфейс Kafka](http://localhost:8081)

Либо просмотреть сообщения сервера можно в реальном времени командой:<br>
`docker-compose exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic mtconnect_data`

4️⃣ **Запуск приложения**

```
# Windows
./build/windows_mtc.exe

# Linux
./build/linux_mtc

# macOS
./build/macos_mtc
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
    "EndpointURL": "http://localhost:5001/Mazak",
    "Model": "Mazak VRX C600"
}'
```

```json
{
    "Status": "ok",
    "connectionInfo": {
        "SessionID": "870c5240-de93-4584-b411-37aa915cbc1d",
        "Config": {
            "EndpointURL": "http://localhost:5001/Mazak",
            "Model": "Mazak VRX C600"
        },
        "CreatedAt": "2025-08-28T12:19:21.2303802+03:00",
        "LastUsed": "2025-08-28T12:19:21.2303802+03:00",
        "UseCount": 1,
        "IsHealthy": true
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
    "Connections": [
        {
            "SessionID": "bba81cc3-ad26-4e0b-9336-a8cc8bf54238",
            "Config": {
                "EndpointURL": "http://localhost:5001/OKUMA",
                "Model": "Okuma MTConnect Adapter",
                "Manufacturer": "OKUMA"
            },
            "CreatedAt": "2025-08-28T12:19:00.4920414+03:00",
            "LastUsed": "2025-08-28T12:19:00.4920414+03:00",
            "UseCount": 1,
            "IsHealthy": true
        },
        {
            "SessionID": "870c5240-de93-4584-b411-37aa915cbc1d",
            "Config": {
                "EndpointURL": "http://localhost:5001/Mazak",
                "Model": "Mazak VRX C600"
            },
            "CreatedAt": "2025-08-28T12:19:21.2303802+03:00",
            "LastUsed": "2025-08-28T12:19:21.2303802+03:00",
            "UseCount": 1,
            "IsHealthy": true
        }
    ],
    "PoolSize": 2,
    "Status": "ok"
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
    "SessionID": "870c5240-de93-4584-b411-37aa915cbc1d"
}'
```

```json
{
    "Status": "healthy",
    "connectionInfo": {
        "SessionID": "870c5240-de93-4584-b411-37aa915cbc1d",
        "Config": {
            "EndpointURL": "http://localhost:5001/Mazak",
            "Model": "Mazak VRX C600"
        },
        "CreatedAt": "2025-08-28T12:19:21.2303802+03:00",
        "LastUsed": "2025-08-28T12:22:33.978361+03:00",
        "UseCount": 2,
        "IsHealthy": true
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
    "SessionID": "870c5240-de93-4584-b411-37aa915cbc1d",
    "Interval": 1000
}'
```

```json
{
    "Message": "Polling started for session 870c5240-de93-4584-b411-37aa915cbc1d",
    "Status": "ok"
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
    "SessionID": "870c5240-de93-4584-b411-37aa915cbc1d"
}'
```

```json
{
    "Message": "Polling stopped for session 870c5240-de93-4584-b411-37aa915cbc1d",
    "Status": "ok"
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
    "SessionID": "870c5240-de93-4584-b411-37aa915cbc1d"
}'
```

```json
{
    "Message": "Session 870c5240-de93-4584-b411-37aa915cbc1d disconnected successfully",
    "Status": "ok"
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
├── tools/
│   └── build/                    # Скрипт для сборки исполняемых файлов
├── build/                        # Папка с готовыми исполняемыми файлами
├── .env                          # Файл конфигурации
├── docker-compose.yml            # Файл для запуска Kafka и Kafka-UI
├── LICENSE
└── README.md
```

## 🆘 Поддержка

- 🐛 [Создайте issue](https://github.com/iwtcode/MTConnect/issues)
- 📧 Напишите на email: iwtcode@gmail.com

## 📝 Лицензия

Проект распространяется под [лицензией MIT](LICENSE)

Copyright (c) 2025 iwtcode