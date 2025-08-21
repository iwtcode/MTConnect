<div align="center">

# MTConnect Parser

![MTConnect](https://img.shields.io/badge/MTConnect-Compatible-blue)
![Go](https://img.shields.io/badge/Go-1.19+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green)

*Парсер для преобразования протокола MTConnect XML в структурированный JSON*

</div>

### ✨ Ключевые возможности

- 🔄 **Актуальность**: Постоянный мониторинг состояния оборудования с частотой в 1 секунду
- 🏭 **Ассинхронность**: Ассинхронная обработка множества MTConnect-эндпоинтов с использование горутин
- 🔧 **Универсальность**: Извлечение и сохранение метаинформации из /probe, универсальной для всех производителей
- 🌐 **REST API**: Простой HTTP API для получения данных в формате JSON

## 🏗️ Архитектура

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   MTConnect     │    │   MTConnect      │    │      REST       │
│   Endpoints     │───▸│   Parser         │───▸│      API        │
│                 │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              ▾
                       ┌──────────────────┐
                       │  Data Store      │
                       │  (Thread-Safe)   │
                       └──────────────────┘
```

## 📦 Установка

1️⃣ **Клонирование репозитория**

```bash
git clone https://github.com/iwtcode/MTConnect.git
cd MTConnect
```

2️⃣ **Запуск приложения на Windows / Linux / MacOS**

```
./build/windows_mtc.exe
./build/linux_mtc
./build/macos_mtc
```

## ⚙️ Конфигурация

Добавьте в файл `config/config.json` свои эндпоинты:

```json
{
  "server_port": "8080",
  "endpoints": [
    "http://localhost:5001/Mazak",
    "http://localhost:5001/OKUMA",
    "https://smstestbed.nist.gov/vds"
  ]
}
```

### Параметры конфигурации

| Параметр | Описание | Пример |
|---|---|---|
| `server_port` | Порт для HTTP сервера | `"8080"` |
| `endpoints` | Список MTConnect эндпоинтов | `[`<br>&nbsp;&nbsp;&nbsp;&nbsp;`"http://machine1",`<br>&nbsp;&nbsp;&nbsp;&nbsp;`"http://machine2"`<br>`]` |

## 🔌 API

### Получение данных станка

```http
GET /api/{machineId}/current
```

**Пример запроса:**
```bash
curl -H "Accept: application/json" http://localhost:8080/api/Mazak/current
```

**Пример ответа:**
```json
{
  "MachineId": "Mazak",
  "Id": "Mazak",
  "Timestamp": "2025-08-21T13:03:34.401887Z",
  "IsEnabled": true,
  "IsInEmergency": false,
  "MachineState": "ACTIVE",
  "ProgramMode": "MANUAL",
  "TmMode": "UNAVAILABLE",
  "HandleRetraceStatus": true,
  "AxisMovementStatus": {
    "B": "STOPPED",
    "C2": "TRAVELING",
    "X": "TRAVELING",
    "Y": "HOMING",
    "Z": "TRAVELING"
  },
  "MstbStatus": "UNAVAILABLE",
  "EmergencyStatus": "ARMED",
  "AlarmStatus": "FAULT",
  "EditStatus": "ACTIVE",
  "ManualMode": true,
  "WriteStatus": "ACTIVE",
  "LabelSkipStatus": "UNAVAILABLE",
  "WarningStatus": "WARNING",
  "BatteryStatus": "UNAVAILABLE",
  "activeToolNumber": "57",
  "toolOffsetNumber": "UNAVAILABLE",
  "AxisInfos": [
    {
      "id": "x",
      "name": "X",
      "type": "LINEAR",
      "data": {
        "axis_feedrate": "1966.37",
        "load": "1.07",
        "position": "754.7812",
        "temperature": "74"
      }
    },
    {
      "id": "y",
      "name": "Y",
      "type": "LINEAR",
      "data": {
        "axis_feedrate": "2674.31",
        "load": "48.57",
        "position": "407.0532",
        "temperature": "87"
      }
    },
    {
      "id": "z",
      "name": "Z",
      "type": "LINEAR",
      "data": {
        "axis_feedrate": "884.22",
        "load": "21.8",
        "position": "925.8021",
        "specification_limit": "",
        "temperature": "52"
      }
    }
  ],
  "FeedRate": {
    "ACTUAL": "2695.19"
  },
  "FeedOverride": {
    "PROGRAMMED": "108",
    "RAPID": "95"
  },
  "Alarms": [
    {
      "componentId": "a",
      "componentName": "base",
      "dataItemId": "servo_cond",
      "level": "FAULT",
      "timestamp": "2025-08-21T13:03:33.309229Z",
      "type": "ACTUATOR"
    },
    {
      "componentId": "a",
      "componentName": "base",
      "dataItemId": "spindle_cond",
      "level": "WARNING",
      "timestamp": "2025-08-21T13:03:34.376533Z",
      "type": "SYSTEM"
    }
  ],
  "hasAlarms": true,
  "PartsCount": {
    "ALL": "28"
  },
  "AccumulatedTime": {
    "x:AUTO": "00:23:34",
    "x:CUT": "00:23:34",
    "x:TOTAL": "00:23:34",
    "x:TOTALCUTTIME": "00:23:34"
  },
  "CurrentProgram": {
    "PROGRAM": "PROG-1096",
    "PROGRAM_COMMENT": "PROG-8591",
    "LINE_NUMBER": "6",
    "LINE_LABEL": "N283"
  },
  "SpindleInfos": [
    {
      "id": "br",
      "name": "B",
      "type": "ROTARY",
      "data": {
        "angle": "139.3119",
        "angular_velocity": "6.37283333333333",
        "load": "62.72",
        "rotary_mode": "INDEX"
      }
    },
    {
      "id": "c",
      "name": "C1",
      "type": "ROTARY",
      "data": {
        "load": "43.73",
        "rotary_mode": "SPINDLE",
        "rotary_velocity": "2427.3",
        "temperature": "95"
      }
    },
    {
      "id": "c2",
      "name": "C2",
      "type": "ROTARY",
      "data": {
        "angle": "237.9649",
        "angular_velocity": "39.9876666666667",
        "load": "36.67",
        "rotary_mode": "INDEX"
      }
    }
  ],
  "ContourFeedRate": "UNAVAILABLE",
  "JogOverride": "UNAVAILABLE"
}
```

## 🔧 Структура проекта

```
MTConnect/
├── build/
│   ├── windows_mtc.exe # Точка входа для Windows
│   ├── linux_mtc       # Точка входа для Linux
│   └── macos_mtc       # Точка входа для MacOS
├── cmd/
│   └── main.go         # Точка входа для Golang
├── config/
│   ├── config.go       # Конфигурационные структуры
│   └── config.json     # Файл конфигурации
├── internal/
|   └── build/
|       └── build.go    # Скрипт для сборки
│   └── mtconnect/
│       ├── client.go   # HTTP клиент для MTConnect
│       ├── models.go   # Структуры данных
│       └── parser.go   # Логика парсинга
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## 🆘 Поддержка

- 🐛 [Создайте issue](https://github.com/iwtcode/MTConnect/issues)
- 📧 Напишите на email: iwtcode@gmail.com

## 📝 Лицензия

Проект распространяется под [лицензией MIT](LICENSE)

Copyright (c) 2025 iwtcode