<div align="center">

# MTConnect Streamer

![alt text](https://img.shields.io/badge/MTConnect-Compatible-blue)
![alt text](https://img.shields.io/badge/Apache%20Kafka-Integrated-blue?logo=apachekafka)
![alt text](https://img.shields.io/badge/Go-1.19+-00ADD8?logo=go)
![alt text](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)
![alt text](https://img.shields.io/badge/License-MIT-green)

*Сервис для сбора данных по протоколу MTConnect, их отправки в Apache Kafka и управления через REST API*

</div>

### ✨ Ключевые возможности
- 🚀 **Потоковая передача в Kafka**: Все данные со станков в реальном времени отправляются в топик Apache Kafka для дальнейшей обработки и аналитики
- 🕹️ **Управляемый опрос**: Запускайте и останавливайте мониторинг для каждого станка индивидуально через REST API с настраиваемым интервалом
- 🌐 **REST API**: Удобный HTTP API для получения актуальных данных, проверки доступности станков и управления процессами опроса
- 🐳 **Простота развертывания**: Готовая конфигурация docker-compose.yml для быстрого запуска Apache Kafka и сопутствующих сервисов
- 🎛️ **Веб-интерфейс для Kafka**: Встроенный Kafka UI для удобного просмотра топиков и сообщений
- 🔧 **Универсальность**: Автоматическое извлечение и кэширование метаинформации из /probe для корректной интерпретации данных с различных станков

## 🏗️ Архитектура

```
┌─────────────────┐      ┌─────────────────┐      ┌──────────────────┐
│   Управляющий   │─────▸│     Сервис      │◂─────│    MTConnect     │
│    REST API     │      │    MTConnect    │      │    Endpoints     │
│   (Gin-Gonic)   │      │    (Go App)     │      │   (XML-данные)   │
└─────────────────┘      └─────────────────┘      └──────────────────┘
        ▲                        │   │
        │                        │   └──────────────────┐
        │ (GET /current)         │ (Polling)            ▼
┌─────────────────┐              ▼             ┌──────────────────┐
│  Пользователь / │    ┌──────────────────┐    │   Apache Kafka   │
│     Система     │    │  In-Memory       │    │   (Потоковая     │
│                 │◂───│  Data Store      │───▸│   обработка)     │
└─────────────────┘    └──────────────────┘    └──────────────────┘
```

## 📦 Установка

1️⃣ **Клонирование репозитория**

```bash
git clone https://github.com/iwtcode/MTConnect.git
cd MTConnect
```

2️⃣ **Конфигурация приложения**

Откройте файл config.json и при необходимости измените его

```json
{
  "server_port": "8080",
  "kafka_brokers": ["localhost:9092"],
  "kafka_topic": "mtconnect_data",
  "endpoints": [
    "http://localhost:5001/Mazak",
    "http://localhost:5001/OKUMA",
    "https://smstestbed.nist.gov/vds"
  ]
}
```

| Параметр | Описание | Пример |
|---|---|---|
| `server_port` | Порт для HTTP сервера | `"8080"` |
| `kafka_brokers` | Список брокеров Kafka для подключения | `["localhost:9092"]` |	
| `kafka_topic` | Имя топика для отправки данных | `"mtconnect_data"` |
| `endpoints` | Список MTConnect эндпоинтов | `[`<br>&nbsp;&nbsp;&nbsp;&nbsp;`"http://machine1",`<br>&nbsp;&nbsp;&nbsp;&nbsp;`"http://machine2"`<br>`]` |

3️⃣ **Запуск Apache Kafka**

```bash
docker-compose up -d
```

После запуска веб-интерфейс Kafka будет доступен по адресу http://localhost:8081

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

## Проверка доступности станка

```http
GET /api/{machineId}/check
```

```bash
curl -X GET "http://localhost:8080/api/Mazak/check"
```

```json
{
  "status": "ok",
  "message": "станок Mazak доступен"
}
```

## Запуск опроса станка

```http
POST /api/{machineId}/polling/start?interval={ms}
```

```bash
curl -X POST "http://localhost:8080/api/Mazak/polling/start?interval=1000"
```

```json
{
  "status": "ok",
  "message": "опрос для станка Mazak запущен"
}
```

## Остановка опроса станка

```http
POST /api/{machineId}/polling/stop
```

```bash
curl -X POST "http://localhost:8080/api/Mazak/polling/stop"
```

```json
{
  "status": "ok",
  "message": "опрос для станка Mazak остановлен"
}
```

## Получение управляющей программы

```http
GET /api/{machineId}/program
```

```bash
curl -X GET "http://localhost:8080/api/Mazak/program"
```

```json
{
  "machineId": "Mazak",
  "message": "функционал получения управляющей программы пока не реализован",
  "program": "G0 X0 Y0..."
}
```

## Получение актуальных данных

```http
GET /api/{machineId}/current
```

```bash
curl http://localhost:8080/api/Mazak/current
```

```json
{
  "MachineId": "Mazak",
  "Id": "Mazak",
  "Timestamp": "2025-08-21T13:03:34.401887Z",
  "IsEnabled": true,
  "MachineState": "ACTIVE",
  "AxisInfos": [
    {
      "id": "x",
      "name": "X",
      "type": "LINEAR",
      "data": { "position": "754.7812" }
    }
  ],
  "Alarms": [],
  "hasAlarms": false,
  "PartsCount": { "ALL": "28" },
  "...": "..."
}
```

## 🔧 Структура проекта

```
MTConnect/
├── cmd/app/              # Главная точка входа приложения
├── internal/
│   ├── app/              # Сборка и запуск приложения (с использованием FX)
│   ├── config/           # Логика загрузки конфигурации
│   ├── adapters/
│   │   ├── handlers/     # Обработчики HTTP-запросов (Gin)
│   │   ├── producers/    # Продюсеры для внешних систем (Kafka)
│   │   └── repositories/ # Реализации репозиториев (in-memory)
│   ├── domain/           # Основные бизнес-сущности и модели
│   ├── interfaces/       # Go-интерфейсы для слоев
│   ├── services/         # Конкретные сервисы (опрос эндпоинтов, парсинг XML)
│   └── usecases/         # Сценарии использования (основная бизнес-логика)
├── tools/
│   └── build/            # Скрипт для сборки исполняемых файлов
├── build/                # Папка с готовыми исполняемыми файлами
├── config.json           # Файл конфигурации
├── docker-compose.yml    # Файл для запуска Kafka
├── LICENSE
└── README.md
```

## 🆘 Поддержка

- 🐛 [Создайте issue](https://github.com/iwtcode/MTConnect/issues)
- 📧 Напишите на email: iwtcode@gmail.com

## 📝 Лицензия

Проект распространяется под [лицензией MIT](LICENSE)

Copyright (c) 2025 iwtcode