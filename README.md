## Технологии

- Go 1.25
- PostgreSQL 17
- `github.com/jackc/pgx/v5`
- `github.com/phpdave11/gofpdf`
- `github.com/ilyakaznacheev/cleanenv`
- `go.uber.org/mock/gomock`
- Docker, Docker Compose
- `migrate/migrate` (в compose для SQL миграций)

## Структура проекта

```text
.
|-- src/
|   |-- cmd/                           
|   |   |-- main.go
|   |   `-- config.go
|   `-- internal/
|       |-- application/
|       |   |-- contracts/
|       |   |-- services/
|       |   `-- mocks/
|       |-- domain/
|       |   |-- models/
|       |   `-- errors.go
|       |-- infrastructure/
|       |   |-- data/
|       |   |-- tsv/
|       |   `-- reportgen/
|       `-- presentation/
|           |-- http_api/
|-- migrations/
|-- Dockerfile
|-- docker-compose.yml
`-- .env.example
```

## Конфигурация

Переменные окружения (см. `.env.example`):

- `APP_ENV` — окружение (`local`, `dev`, `prod`)
- `HTTP_ADDR` — адрес HTTP сервера (по умолчанию `:8080`)
- `WATCHER_DIR` — директория для входных `.tsv`
- `WATCHER_HEARTBEAT` — период сканирования (`3s`, `10s`, ...)
- `REPORTS_DIR` — директория для PDF
- `POSTGRES_DSN` — строка подключения
- `POSTGRES_MAX_CONNS`
- `POSTGRES_MIN_CONNS`
- `POSTGRES_MAX_CONN_LIFETIME`
- `POSTGRES_MAX_CONN_IDLE_TIME`
- `POSTGRES_HEALTH_CHECK_PERIOD`

## Схема БД

Миграции в папке `migrations/` создают:

- `file_data` — данные устройств из `.tsv`
- `processed_files` — учет обработанных файлов и ошибок обработки

## Запуск

### Через Docker Compose

Сервисы:

- `db` — PostgreSQL
- `migrate` — применяет SQL миграции
- `app` — приложение

Запуск:

```bash
docker compose up --build
```

## Пример использования

### 1. Положить входной `.tsv` файл в директорию watcher

### 2. Дождаться обработки

- данные окажутся в БД (`file_data`)
- в `processed_files` появится запись об обработке
- в `REPORTS_DIR` появятся PDF-отчеты по `unit_guid`

### 3. Получить данные через API

```bash
curl "http://localhost:8080/device-data/<unit_guid>?take=100&offset=0"
```

Пример ответа:

```json
{
  "data": [
    {
      "n": 1,
      "mqtt": "...",
      "invid": "...",
      "unit_guid": "...",
      "msg_id": "...",
      "text": "...",
      "context": "...",
      "class": "...",
      "level": 1,
      "area": "...",
      "addr": "...",
      "block": "...",
      "type": "...",
      "bit": "...",
      "invert_bit": "..."
    }
  ],
  "take": 100,
  "offset": 0,
  "total": 1
}
```

## Тесты

```bash
go test ./...
```

## Линт

```bash
golangci-lint run
```