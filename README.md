# Notes Service

Микросервис заметок на Go для подготовки к собеседованию.
Использует чистую архитектуру, PostgreSQL, Redis, Prometheus и Grafana.

## Стек
- **Go 1.26**, **chi** (роутер)
- **PostgreSQL** (основное хранилище), **pgx** (драйвер)
- **Redis** (кэширование), **go-redis**
- **Prometheus** + **Grafana** (метрики и мониторинг)
- **Docker Compose** (окружение)
- **OTPL** + **Jaeger** (трейсинг)
- **Swagger** (документация)
- **Mock testing** (тестирование)


## Фичи для отказоустойчивости / оптимизации 
- **Gracefull shutdown** 
- **Health Checker (PostgreSQL, Redis)**, работа через **BlackBox** 
- **Rate Limiter (Token Buckets)**
- **Pagination (Offset/Limit)**, В планах сделать через Cursor
- **Cache invalidation** при Update, Delete

## Запуск
```bash
docker compose up -d          # PostgreSQL, Redis, Prometheus, Grafana
go run ./cmd/server/          # сам сервис

