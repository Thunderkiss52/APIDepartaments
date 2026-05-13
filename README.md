# APIDepartaments

REST API для управления организационной структурой компании.

Проект поддерживает:
- подразделения;
- сотрудников;
- дерево подразделений;
- перемещение подразделений;
- удаление подразделений в режимах `cascade` и `reassign`.

## Стек

- Go
- `net/http`
- GORM
- PostgreSQL
- goose
- Docker / Docker Compose v2

## Структура проекта

- `cmd/api/main.go` - точка входа приложения
- `internal/app` - сборка зависимостей
- `internal/config` - загрузка переменных окружения
- `internal/db` - подключение к PostgreSQL с retry
- `internal/department` - модели, repository, service и HTTP handlers
- `internal/router` - маршрутизация на `net/http`
- `migrations` - миграции `goose`
- `tests` - `httptest` и unit-тесты сервисного слоя

## Быстрый запуск

1. Подними стек:

```bash
docker compose up --build
```

`.env` не обязателен для Docker-запуска: в `docker-compose.yml` уже заданы значения по умолчанию. Если нужно переопределить порт или доступ к БД, скопируй пример и измени значения:

```bash
cp .env.example .env
```

После старта API доступно по адресу:

```text
http://localhost:8080
```

## Docker Compose

В `docker-compose.yml` есть 3 сервиса:

- `postgres` - PostgreSQL;
- `migrate` - запуск миграций `goose`;
- `app` - само API.

`postgres` поднимается с `healthcheck`, `migrate` ждёт готовности базы, а `app` стартует после успешного завершения миграций.

## Миграции

Миграции находятся в папке `migrations/`.

Ручной запуск миграций:

```bash
docker compose run --rm migrate
```

Автоматический запуск тоже предусмотрен через `docker compose up --build`.

## Переменные окружения

Файл `.env.example`:

```env
APP_PORT=8080

DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=org_structure
DB_SSLMODE=disable
```

## Модель данных

### Department

- `id: int`
- `name: string`
- `parent_id: int | null`
- `created_at: datetime`

### Employee

- `id: int`
- `department_id: int`
- `full_name: string`
- `position: string`
- `hired_at: date | null`
- `created_at: datetime`

## API

### 0. Проверить API

`GET /health`

Пример:

```bash
curl http://localhost:8080/health
```

Успех:
- `200 OK`

### 0.1. Получить список подразделений

`GET /departments`

Пример:

```bash
curl http://localhost:8080/departments
```

Ответ содержит массив `departments`.

### 1. Создать подразделение

`POST /departments/`

Тело:

```json
{
  "name": "IT",
  "parent_id": null
}
```

Пример:

```bash
curl -X POST http://localhost:8080/departments/ \
  -H "Content-Type: application/json" \
  -d '{"name":"IT","parent_id":null}'
```

Успех:
- `201 Created`

Ошибки:
- `400 Bad Request` - пустое имя, слишком длинное имя, некорректный `parent_id`
- `404 Not Found` - указан несуществующий `parent_id`
- `409 Conflict` - имя уже занято в рамках того же родителя

### 2. Создать сотрудника в подразделении

`POST /departments/{id}/employees/`

Тело:

```json
{
  "full_name": "Ivan Ivanov",
  "position": "Backend Developer",
  "hired_at": "2026-05-13"
}
```

Пример:

```bash
curl -X POST http://localhost:8080/departments/1/employees/ \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Ivan Ivanov","position":"Backend Developer","hired_at":"2026-05-13"}'
```

Успех:
- `201 Created`

Ошибки:
- `400 Bad Request` - пустое/слишком длинное `full_name` или `position`, неверный формат `hired_at`
- `404 Not Found` - подразделение не найдено

### 3. Получить подразделение с сотрудниками и деревом

`GET /departments/{id}`

Query-параметры:

- `depth` - глубина дерева, по умолчанию `1`, максимум `5`
- `include_employees` - `true` или `false`, по умолчанию `true`

Пример:

```bash
curl "http://localhost:8080/departments/1?depth=2&include_employees=true"
```

Ответ содержит:

- `department` - текущий отдел
- `employees` - сотрудники отдела
- `children` - вложенные подразделения рекурсивно до указанной глубины

Ошибки:
- `400 Bad Request` - `depth < 0`, `depth > 5` или некорректный query
- `404 Not Found` - подразделение не найдено

### 4. Обновить подразделение / переместить подразделение

`PATCH /departments/{id}`

Тело:

```json
{
  "name": "Engineering",
  "parent_id": 2
}
```

Можно передать:
- только `name`
- только `parent_id`
- оба поля

Можно также явно сбросить родителя:

```json
{
  "parent_id": null
}
```

Пример:

```bash
curl -X PATCH http://localhost:8080/departments/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Engineering","parent_id":2}'
```

Ограничения:
- нельзя сделать подразделение родителем самого себя
- нельзя переместить подразделение внутрь собственного поддерева
- нельзя указать несуществующий `parent_id`
- имя должно быть от 1 до 200 символов после `TrimSpace`

Ошибки:
- `400 Bad Request` - невалидные данные
- `404 Not Found` - подразделение или новый родитель не найдены
- `409 Conflict` - цикл или конфликт имени

### 5. Удалить подразделение

`DELETE /departments/{id}`

Query-параметры:

- `mode=cascade`
- `mode=reassign&reassign_to_department_id=5`

#### cascade

Пример:

```bash
curl -X DELETE "http://localhost:8080/departments/1?mode=cascade"
```

Логика:
- удаляется выбранное подразделение;
- удаляются все дочерние подразделения;
- удаляются все сотрудники удаляемого поддерева;
- удаление идёт через БД/ORM.

#### reassign

Пример:

```bash
curl -X DELETE "http://localhost:8080/departments/1?mode=reassign&reassign_to_department_id=5"
```

Логика:
- удаляется выбранное подразделение;
- сотрудники удаляемого подразделения переводятся в `reassign_to_department_id`;
- дочерние подразделения удаляются каскадно вместе со своими сотрудниками.

Ограничения:
- `reassign_to_department_id` обязателен при `mode=reassign`
- нельзя переводить сотрудников в несуществующее подразделение
- нельзя переводить сотрудников в само удаляемое подразделение
- нельзя переводить сотрудников в подразделение, которое находится внутри удаляемого поддерева

Ответ:
- `204 No Content`

Ошибки:
- `400 Bad Request` - некорректный `mode` или отсутствует `reassign_to_department_id`
- `404 Not Found` - подразделение не найдено
- `409 Conflict` - попытка перевести в поддерево или нарушить ограничения

## Валидация

Правила:
- все входные строки триммятся через `strings.TrimSpace`
- `name`, `full_name`, `position` - длина от 1 до 200 символов
- `hired_at` - `YYYY-MM-DD`, либо `null`
- `depth` - от 0 до 5

## Формат ошибок

Пример:

```json
{
  "error": "validation_error"
}
```

## Тесты

Есть:
- `httptest` на создание подразделения
- unit-тесты на дерево, цикл и `reassign`

Запуск локально:

```bash
go test ./...
```

## Проверка руками

Рекомендуемый сценарий:

1. Поднять стек:

```bash
docker compose up --build
```

2. Проверить, что API отвечает:

```bash
curl http://localhost:8080/health
```

3. Посмотреть текущие подразделения:

```bash
curl http://localhost:8080/departments
```

4. Создать корневое подразделение. Если в базе уже есть `Engineering`, API вернёт `409 Conflict`; это нормальная защита от дублей. Для повторной проверки используй другое имя:

```bash
curl -X POST http://localhost:8080/departments/ \
  -H "Content-Type: application/json" \
  -d '{"name":"Engineering","parent_id":null}'
```

5. Создать сотрудника в подразделении:

```bash
curl -X POST http://localhost:8080/departments/1/employees/ \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Ivan Ivanov","position":"Backend Developer","hired_at":"2026-05-13"}'
```

6. Получить дерево:

```bash
curl "http://localhost:8080/departments/1?depth=2&include_employees=true"
```

7. Обновить подразделение:

```bash
curl -X PATCH http://localhost:8080/departments/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Engineering","parent_id":null}'
```

8. Проверить удаление:

```bash
curl -X DELETE "http://localhost:8080/departments/1?mode=cascade"
```

или:

```bash
curl -X DELETE "http://localhost:8080/departments/1?mode=reassign&reassign_to_department_id=5"
```

## Готовность к сдаче

Проект упакован так, чтобы его можно было поднять командой:

```bash
docker compose up --build
```

После этого API доступно на:

```text
http://localhost:8080
```
