# link-storage-service

Сервис коротких ссылок: создаёт короткие коды, отдаёт по ним оригинальный URL и считает переходы.

## Запуск

Поднять Postgres и Redis:

```
docker compose up -d
```

Запустить сервис:

```
go run ./cmd/link_storage_service
```

Сервис слушает `localhost:8080`.

## Настройки

Всё конфигурируется через переменные окружения, у каждой есть дефолт — можно запускать без настройки.
Основное: `ADDRESS`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `REDIS_ADDR`.
Полный список — в `internal/config/config.go`.

## Примеры запросов

Создать ссылку:

```
curl -X POST localhost:8080/links -d '{"url":"https://example.com"}'
```

Получить по короткому коду:

```
curl localhost:8080/links/abc123
```

Список с пагинацией:

```
curl "localhost:8080/links?limit=10&offset=0"
```

Статистика:

```
curl localhost:8080/links/abc123/stats
```

Удалить:

```
curl -X DELETE localhost:8080/links/abc123
```