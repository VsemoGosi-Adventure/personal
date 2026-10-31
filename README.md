# personal

Микросервис профилей и интересов проекта «Авантюра». JWT проверяет API Gateway; сервис получает UUID пользователя из заголовка `X-User-ID`.

## Запуск

Нужны PostgreSQL и переменная `DATABASE_URL` (строка подключения `pgx`). При старте сервис сам применяет ожидающие миграции. Необязательные переменные: `HTTP_ADDR` (по умолчанию `:8080`) и `SHUTDOWN_TIMEOUT` (по умолчанию `10s`).

```sh
go run ./cmd
```

## HTTP API

Все маршруты `/v1` требуют `X-User-ID`. `GET /healthz` доступен без него.

| Метод | Маршрут | Действие |
| --- | --- | --- |
| `GET` | `/v1/profile` | Прочитать свой профиль |
| `PUT` | `/v1/profile` | Создать или полностью заменить редактируемые поля профиля |
| `GET` | `/v1/interests` | Получить справочник интересов |
| `GET` | `/v1/profile/interests` | Получить свои интересы |
| `PUT` | `/v1/profile/interests/:interestID` | Добавить интерес (повторный запрос безопасен) |
| `DELETE` | `/v1/profile/interests/:interestID` | Удалить интерес (повторный запрос безопасен) |

Тело `PUT /v1/profile`: `{"username":"Alice","bio":"Hello","avatarUrl":null}`. `username` обязателен; `bio` и `avatarUrl` могут быть строкой или `null`. Ответы используют JSON с полями `userId`, `createdAt`, `updatedAt`. Ошибки имеют форму `{"error":{"code":"...","message":"..."}}`.

Хранение настроек приватности пока не реализовано: таблицы для него нет.

## Контейнер

```sh
docker build -t personal:local .
docker run --rm -p 8080:8080 -e DATABASE_URL='postgres://user:password@db-host:5432/personal' personal:local
```

При запуске (локально и в контейнере) сервис применяет ожидающие миграции из `migrations/` до старта HTTP-сервера. SQL-файлы встроены в бинарник; повторный запуск не применяет уже записанные версии. Для первой миграции PostgreSQL должен поддерживать PostGIS, а пользователь БД — иметь право создать расширение. По умолчанию сервис слушает порт `8080`.

## Swagger

Спецификация находится в `docs/swagger.json` и `docs/swagger.yaml`. После изменения комментариев к ручкам пересоздайте её командой:

```sh
swag init -g main.go -d cmd,internal/handler --parseInternal -o docs
```
