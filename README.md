# Go Test

Это проект REST API на Go для управления товарами с кешированием, логированием и распределённой системой сообщений. Используются PostgreSQL для хранения данных, Redis для кеширования, ClickHouse для логирования, NATS для очереди сообщений и Docker для контейнеризации.

## Возможности
- Операции CRUD для сущности `goods` (создание, чтение, обновление, удаление).
- Кеширование в Redis с TTL 1 минута.
- Транзакционные обновления с блокировкой на уровне строк.
- Логирование в ClickHouse через NATS (пачками).
- Развёртывание через Docker.

## Требования
- Docker и Docker Compose.
- Git (для клонирования репозитория).
- Go (для локальной разработки, опционально).

## Установка

1. **Клонируйте репозиторий**:
   ```bash
   git clone https://github.com/EmelinDanila/go_test.git
   cd go_test
   ```

2. **Соберите и запустите через Docker Compose**:
   ```bash
   docker-compose up -d --build
   ```
   - Это запустит PostgreSQL, Redis, ClickHouse, NATS и приложение Go.

3. **Примените миграции базы данных**:
   - Скопируйте файл миграции:
     ```bash
     docker cp migrations/001_init.up.sql go_test-postgres-1:/tmp/migration.sql
     ```
   - Примените миграцию:
     ```bash
     docker exec go_test-postgres-1 psql -U postgres -d go_test_db -f /tmp/migration.sql
     ```

## Использование

### Конечные точки API
- **GET `/goods/{id}`**: Получить товар (кешируется в Redis).
  - Пример: `curl http://localhost:8080/goods/1`
- **POST `/goods`**: Создать товар.
  - Пример: `curl -X POST http://localhost:8080/goods -H "Content-Type: application/json" -d '{"project_id":1,"name":"Тест","description":"Тестовое описание"}'`
- **PATCH `/goods/{id}`**: Обновить товар.
  - Пример: `curl -X PATCH http://localhost:8080/goods/1 -H "Content-Type: application/json" -d '{"name":"Обновлённый тест"}'`
- **DELETE `/goods/{id}`**: Удалить товар.
  - Пример: `curl -X DELETE http://localhost:8080/goods/1`

## Настройка
- Измените `docker-compose.yml` для настройки портов или переменных окружения (например, `POSTGRES_PASSWORD`).
- Файл миграции находится в `migrations/001_init.up.sql`.