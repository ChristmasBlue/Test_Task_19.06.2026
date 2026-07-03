# Task Manager API

API для управления задачами в командах с поддержкой ролевой модели, истории изменений и сложных SQL-запросов.

## 📋 Оглавление

- [Стек технологий](#-стек-технологий)
- [Быстрый старт](#-быстрый-старт)
- [API Документация](#-api-документация)
- [Метрики](#-метрики)
- [Тесты](#-тесты)
- [Структура проекта](#-структура-проекта)
- [Переменные окружения](#-переменные-окружения)
- [Запуск миграций](#-запуск-миграций)

---

## 🛠 Стек технологиий

| Компонент | Технология |
|-----------|------------|
| **Язык** | Go 1.26+ |
| **HTTP фреймворк** | Gin |
| **База данных** | MySQL 8.0 |
| **Кеш** | Redis 7 |
| **Миграции** | Goose |
| **Метрики** | Prometheus |
| **Документация** | Swagger (OpenAPI) |
| **Тестирование** | testify + gomock |
| **Контейнеризация** | Docker + Docker Compose |

---

## 🚀 Быстрый старт

### Требования

- Go 1.26+
- Docker + Docker Compose
- MySQL 8.0 (если без Docker)
- Redis 7 (если без Docker)

### Запуск с Docker (рекомендуется)

```bash
# 1. Клонировать репозиторий
git clone <your-repo>
cd Test_task3

# 2. Запустить все сервисы
cd build/docker
docker-compose up -d

# 3. Проверить логи
docker-compose logs -f app
```
Приложение будет доступно на http://localhost:8080.

### Запуск без Docker

```bash
# 1. Установить зависимости
go mod download

# 2. Создать базу данных
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS taskmanager"

# 3. Применить миграции
goose -dir ./migrations mysql "root:root@tcp(localhost:3306)/taskmanager?parseTime=true" up

# 4. Запустить приложение
go run cmd/app/main.go
```

## 📚 API Документация
После запуска Swagger UI доступен по адресу:
```text
http://localhost:8080/swagger/index.html
```

### Основные ручки

### 🔐 Аутентификация
```text
Метод	Путь	            Описание
POST	/api/v1/register    Регистрация пользователя
POST	/api/v1/login	    Вход (JWT)
```

### 👥 Команды
```text
Метод   Путь                        Описание
POST    /api/v1/teams               Создать команду
GET     /api/v1/teams               Список команд пользователя
GET     /api/v1/teams/:id           Получить команду по ID
POST    /api/v1/teams/:id/invite    Пригласить пользователя
GET     /api/v1/teams/:id/members   Список участников
```

### 📋 Задачи
```text
Метод	    Путь                        Описание
POST        /api/v1/tasks               Создать задачу
GET         /api/v1/tasks               Список задач (фильтрация)
GET         /api/v1/tasks/:id           Получить задачу
PUT         /api/v1/tasks/:id           Обновить задачу
GET         /api/v1/teams/:id/tasks     Задачи команды
GET         /api/v1/tasks/:id/history   История изменений
```

### 💬 Комментарии
```text
Метод   Путь                            Описание
POST    /api/v1/tasks/:id/comments      Добавить комментарий
GET     /api/v1/tasks/:id/comments      Список комментариев
```

### 📊 Отчёты (сложные SQL)
```text
Метод   Путь                                Описание
GET     /api/v1/reports/team-stats          Статистика по командам (JOIN + агрегация)
GET     /api/v1/reports/top-creators        Топ-3 создателей задач (RANK)
GET     /api/v1/reports/invalid-assignee    Задачи с невалидным исполнителем
```

### 📊 Метрики
Prometheus метрики доступны на отдельном порту:

```text
http://localhost:9090/metrics
```
Собираются метрики:

- Количество запросов (http_requests_total)

- Ошибки (http_errors_total)

- Время ответа (http_request_duration_seconds)

## 🧪 Тесты
### Unit-тесты
```bash
# Запустить unit-тесты
go test ./internal/cases -v

# С покрытием
go test ./internal/cases -cover

# Детальный отчёт о покрытии
go test ./internal/cases -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Покрытие: 85.4% (выше требования ТЗ).

## Интеграционные тесты
```bash
# Все интеграционные тесты
go test ./... -v -run Integration

# Только без интеграционных
go test ./... -v -short
```

## 📁 Структура проекта

```text
Test_task3/
├── cmd/
│   └── app/
│       └── main.go                 # Точка входа
├── internal/
│   ├── adapters/
│   │   ├── config/                 # Конфигурация
│   │   └── storage/
│   │       ├── cache/              # Redis
│   │       └── database/           # MySQL
│   ├── cases/                      # Бизнес-логика (Service)
│   ├── entities/                   # Сущности
│   ├── ports/
│   │   └── http/
│   │       └── public/             # HTTP handlers
│   └── dto/                        # Data Transfer Objects
├── pkg/
│   ├── application/                # Запуск приложения
│   └── dto/                        # Общие DTO
├── tools/                          # Утилиты (JWT, хеширование)
├── migrations/                     # SQL миграции (Goose)
├── configs/                        # Конфигурационные файлы
├── docs/                           # Swagger документация
├── build/
│   └── docker/
│       ├── docker-compose.yml
│       └── taskmanager/
│           └── Dockerfile
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

### 🔧 Переменные окружения
```text
Переменная      Описание                            По умолчанию
CONFIG_PATH     Путь к конфигурационному файлу      configs/config.yaml
DB_DSN          Строка подключения к MySQL          —
REDIS_ADDR      Адрес Redis                         localhost:6379
GIN_MODE        Режим Gin (debug/release)           debug
```

### 📦 Запуск миграций

Через Docker

```bash
docker-compose exec migrate goose -dir /migrations mysql "..." up
```

Через CLI (локально)

```bash
# Установить goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Применить миграции
goose -dir ./migrations mysql "root:root@tcp(localhost:3306)/taskmanager?parseTime=true" up

# Откатить
goose -dir ./migrations mysql "root:root@tcp(localhost:3306)/taskmanager?parseTime=true" down

# Проверить статус
goose -dir ./migrations mysql "root:root@tcp(localhost:3306)/taskmanager?parseTime=true" status
```

