# Notes Service

Микросервис заметок на Go для подготовки к собеседованию.
Использует чистую архитектуру, PostgreSQL, Redis, Prometheus и Grafana.

## Стек
- **Go 1.25**, **chi** (роутер)
- **PostgreSQL** (основное хранилище), **pgx** (драйвер)
- **Redis** (кэширование), **go-redis**
- **Prometheus** + **Grafana** (метрики и мониторинг)
- **Docker Compose** (окружение)

## Запуск
```bash
docker compose up -d          # PostgreSQL, Redis, Prometheus, Grafana
go run ./cmd/server/          # сам сервис
