# personal

Микросервис профилей, интересов и аутентификации проекта «Авантюра». Сервис выпускает access JWT и ротируемые refresh-токены. API Gateway вызывает `/v1/auth/verify`, затем передает проверенный `X-User-ID` защищенным маршрутам.

## Запуск

Нужны PostgreSQL, `DATABASE_URL` (строка подключения `pgx`) и `JWT_SECRET` (не менее 32 байт; один и тот же секрет на всех экземплярах сервиса). При старте сервис сам применяет ожидающие миграции. Необязательные переменные: `HTTP_ADDR` (по умолчанию `:8080`) и `SHUTDOWN_TIMEOUT` (по умолчанию `10s`).

```sh
go run ./cmd
```

## HTTP API

Маршруты профиля и интересов требуют `X-User-ID` от доверенного API Gateway. Маршруты `/v1/auth/*` и `GET /healthz` доступны без него. Не открывайте защищенные маршруты напрямую внешним клиентам: `X-User-ID` принимается как доверенный заголовок.

| Метод | Маршрут | Действие |
| --- | --- | --- |
| `GET` | `/v1/profile` | Прочитать свой профиль |
| `PUT` | `/v1/profile` | Создать или полностью заменить редактируемые поля профиля |
| `GET` | `/v1/interests` | Получить справочник интересов |
| `GET` | `/v1/profile/interests` | Получить свои интересы |
| `PUT` | `/v1/profile/interests/:interestID` | Добавить интерес (повторный запрос безопасен) |
| `DELETE` | `/v1/profile/interests/:interestID` | Удалить интерес (повторный запрос безопасен) |
| `POST` | `/v1/auth/register` | Регистрация и выдача пары токенов (201) |
| `POST` | `/v1/auth/login` | Вход и выдача пары токенов |
| `POST` | `/v1/auth/refresh` | Однократная ротация refresh-токена |
| `POST` | `/v1/auth/logout` | Сброс refresh-токена (204) |
| `GET` | `/v1/auth/verify` | Проверка Bearer JWT для Traefik ForwardAuth; ответ `X-User-ID` |

Тело `PUT /v1/profile`: `{"username":"Alice","bio":"Hello","avatarUrl":null}`. `username` обязателен; `bio` и `avatarUrl` могут быть строкой или `null`. Ответы используют JSON с полями `userId`, `createdAt`, `updatedAt`. Ошибки имеют форму `{"error":{"code":"...","message":"..."}}`.

Хранение настроек приватности пока не реализовано: таблицы для него нет.

Регистрация принимает `{"email":"alice@example.com","password":"password123","username":"Alice"}`. Вход принимает `email` и `password`; обновление и выход — `refresh_token`. Ответ с токенами: `{"access_token":"...","refresh_token":"...","expires_in":900}`. Access JWT живет 15 минут, refresh-токен — 30 дней; в БД хранится только SHA-256 хеш refresh-токена. Повторное использование старого refresh-токена после ротации отклоняется. Выход сбрасывает refresh-токен; уже выданный access JWT остается действительным до истечения срока.

## Контейнер

```sh
docker build -t personal:local .
docker run --rm -p 8080:8080 -e DATABASE_URL='postgres://user:password@db-host:5432/personal' -e JWT_SECRET='replace-with-a-random-secret-of-at-least-32-bytes' personal:local
```

При запуске (локально и в контейнере) сервис применяет ожидающие миграции из `migrations/` до старта HTTP-сервера. SQL-файлы встроены в бинарник; повторный запуск не применяет уже записанные версии. Для первой миграции PostgreSQL должен поддерживать PostGIS, а пользователь БД — иметь право создать расширение. По умолчанию сервис слушает порт `8080`.

## Swagger

Спецификация находится в `docs/swagger.json` и `docs/swagger.yaml`. После изменения комментариев к ручкам пересоздайте её командой:

```sh
swag init -g main.go -d cmd,internal/handler --parseDependency --parseInternal -o docs
```
