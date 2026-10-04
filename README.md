# School

Платформа для онлайн/оффлайн школы: видео-уроки, проверка заданий, расписание, финансы.

> ✅ **Фаза 0 (Фундамент) завершена.** Все базовые сервисы запущены и объединены через единую точку входа (Nginx reverse-proxy).

---

## Стек

- **Backend**: Go 1.24 (DDD, net/http, Viper)
- **Frontend**: React 18 + TypeScript + Vite (Hot Reload, Strict Mode)
- **База данных**: PostgreSQL 16 + Valkey 8 (in-memory / Redis-совместимый)
- **Объектное хранилище**: MinIO (S3-compatible)
- **Инфраструктура**: Docker Compose + Nginx (Reverse Proxy & WebSocket)

---

## Быстрый старт

### 1. Запуск всей платформы через Docker Compose

```bash
# 1. Скопировать переменные окружения
cp .env.example .env

# 2. Собрать и запустить все 7 сервисов
docker compose up -d --build

# 3. Проверить статус контейнеров
docker compose ps
```

После старта веб-интерфейс доступен в браузере:
👉 **[http://localhost](http://localhost)**

---

## Сервисы, порты и URL

| Сервис | Контейнер | Внутренний порт | Внешний URL / Порт хоста | Описание |
|---|---|---|---|---|
| **Nginx (Единая точка входа)** | `school_nginx` | 80 | **http://localhost** (80) | Проксирует `/api/*` на backend, `/*` на frontend, поддерживает Vite HMR WebSocket |
| **Frontend Dev Server** | `school_frontend` | 3000 | **http://localhost:3000** (3000) | React SPA на Vite (dev-сервер с hot reload, health poll) |
| **Go Backend API** | `school_backend` | 8080 | **http://localhost:8080** (8080) | REST API (`/health`, `/api/v1/health`) |
| **MinIO Console** | `school_minio` | 9001 | **http://localhost:9001** (9001) | Веб-консоль S3 (`minioadmin` / `minioadmin`) |
| **MinIO S3 API** | `school_minio` | 9000 | **localhost:9000** (9000) | S3 API хранилища файлов |
| **PostgreSQL** | `school_postgres` | 5432 | **localhost:5432** (5432) | База данных (`school` / `school_dev_password`) |
| **Valkey** | `school_valkey` | 6379 | **localhost:6379** (6379) | Сессии и кэш (Redis-совместимый) |

---

## Проверка работоспособности (Health Checks)

### Проверка через единую точку входа (Nginx)

```bash
# Проверка здоровья API через Nginx (HTTP 200 OK + JSON)
curl -i http://localhost/api/v1/health

# Проверка главной страницы фронтенда через Nginx (HTTP 200 OK + HTML)
curl -i http://localhost/

# Проверка канонического редиректа /api -> /api/ (HTTP 308)
curl -i http://localhost/api
```

### Проверка прямого доступа к API бэкенда

```bash
# Healthcheck бэкенда напрямую
curl -i http://localhost:8080/health
curl -i http://localhost:8080/api/v1/health
```

---

## Разработка и тестирование

### Фронтенд (`frontend/`)

```bash
cd frontend

# Установка зависимостей
npm install

# Проверка типов TypeScript и сборка bundle
npm run build

# Проверка линтером (ESLint 9)
npm run lint

# Локальный запуск dev-сервера без Docker
npm run dev
```

### Бэкенд (`backend/`)

```bash
# Запуск unit-тестов
make test

# Сборка бинарника API
make build

# Запуск линтера golangci-lint
make lint

# Применение миграций
make migrate-up
```

---

## Документация

| Документ | Описание |
|----------|----------|
| [Видение проекта](docs/PROJECT_VISION.md) | Что строим и зачем |
| [Бизнес-спецификация](docs/PRODUCT_SPEC.md) | Бизнес-логика, сценарии ролей, правила слотов и денег |
| [Дизайн-система](docs/DESIGN_SYSTEM.md) | Спецификация UI/UX: стиль Apple Liquid Glass, палитра и правила стекла |
| [Roadmap](docs/ROADMAP.md) | Глобальный план по фазам |
| [Текущий спринт](docs/CURRENT_SPRINT.md) | Что делаем сейчас (статус задач) |
| [Конвенции](docs/CONVENTIONS.md) | Как пишем код (Go, React, архитектура) |
| [ADR](docs/decisions/) | Архитектурные решения |

---

## Структура проекта

```
school/
├── backend/          # Go 1.24 DDD бэкенд (cmd/api, internal: app, domain, infrastructure)
├── frontend/         # React 18 + Vite + TypeScript приложение
│   ├── src/          # Исходный код (app, api, shared, types)
│   ├── Dockerfile    # Dev Dockerfile с volume-пробросом
│   └── package.json  # Зависимости и скрипты
├── deploy/
│   └── nginx/        # Nginx конфигурация (nginx.conf reverse-proxy)
├── docs/             # Архитектурная и проектная документация
├── scripts/          # Вспомогательные скрипты
├── docker-compose.yml# Координация всех 7 сервисов
├── .env.example      # Пример переменных окружения
└── README.md
```
