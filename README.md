# School — SaaS-Ассистент Репетитора (Tutor Assistant)

> **Автономное single-player рабочее место для частных преподавателей и репетиторов**: CRM учеников, интерактивное расписание Apple Calendar, учет в часах и форматные абонементы, финансовый дашборд и дизайн-система Apple Liquid Glass.

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react)](https://react.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.5-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Valkey](https://img.shields.io/badge/Valkey-8.0-FF4438?style=flat&logo=redis)](https://valkey.io/)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind-3.4-38B2AC?style=flat&logo=tailwind-css)](https://tailwindcss.com/)
[![Docker Compose](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![E2E Tests](https://img.shields.io/badge/E2E_Tests-43_Passed_(100%25)-brightgreen?style=flat)](tests/)

---

## 📌 Текущее состояние платформы

- ✅ **Фаза 0 (Фундамент)**: Go DDD архитектура, Docker Compose, Nginx, Vite HMR, Viper, golang-migrate.
- ✅ **Фаза 1 (Аутентификация)**: регистрация, логин, JWT с ротацией refresh-токенов в Valkey, RBAC.
- ✅ **Фаза 2 (Спринты 2, 2.1, 2.1.2, 2.1.3)**: 
  - **CRM учеников**: ведение базы клиентов, тарифные сетки ставок (`индив` / `пара` / `группа`), динамические партнерские теги школ с комиссией %, полнотекстовый живой поиск, табы статуса («Активные» / «В архиве» / «Все»), сохранение персонального дефолтного прайса.
  - **Интерактивное расписание**: режимы Неделя (7 колонок), День, Список. Поддержка нахлёстов занятий (парные/групповые), быстрое назначение в 1 клик, подтверждение проведения прямо на карточке (`✓`), строгое разделение формата (`individual`/`pair`/`group`) и локации (`оффлайн` кабинет / `онлайн` звонок).
  - **Почасовой учет и абонементы ([ADR-007](docs/decisions/0007-hourly-rates-and-format-subscriptions.md))**: покупка абонементов по форматам в часах, автоматическое списание точной длительности (включая дробные 1.5 ч), ручная корректировка баланса с аудитом причин ([`client_balance_adjustments`](backend/migrations/000006_sprint_2_1_3_features.up.sql)).
  - **Финансовый дашборд**: расчет Gross Potential Revenue (потенциал расписания), Net Tutor Income (чистый доход репетитора за вычетом комиссий школ) и средней ставки в час.
  - **Дизайн Apple Liquid Glass ([ADR-006](docs/decisions/0006-app-layout-and-glass-conventions.md))**: анимированный канвас `<LiquidBackground />`, сквозной `<AppLayout />`, полупрозрачные стеклянные карточки `bg-white/30 backdrop-blur-md border border-white/40`.
  - **Производительность ([ADR-008](docs/decisions/0008-frontend-decomposition-and-valkey-csc.md))**: Client-Side Caching (CSC) через Valkey для справочников с инвалидацией по RESP3, декомпозиция фронтенда на модули `features/` (< 250 строк).
  - **Сетевой доступ**: Nginx настроен для локальной сети LAN / Wi-Fi (`http://192.168.0.21/`).
  - **E2E автотесты**: **43 passed** в Docker (`make test-e2e` 100% pass).

---

## 🛠️ Стек технологий

| Компонент | Технологии | Назначение |
|---|---|---|
| **Backend** | Go 1.24, `net/http` stdlib, `squirrel`, `slog`, `viper` | Hexagonal / DDD микро-монолит, REST API `/api/v1/` |
| **API Contract** | OpenAPI 3.0, `oapi-codegen` | Спецификация контрактов и кодогенерация типов |
| **Database** | PostgreSQL 16, `golang-migrate` (миграция 6) | Реляционные данные (пользователи, клиенты, уроки, абонементы) |
| **Cache / Sessions** | Valkey 8, `valkey-go` | Сессии, токены, Client-Side Caching (CSC) справочников |
| **Object Storage** | MinIO (S3-compatible) | Хранение файлов и аватарок |
| **Frontend** | React 18, TypeScript, Tailwind CSS, Vite, Lucide | Модульное SPA в стиле Apple Liquid Glass |
| **Proxy / Gateway** | Nginx (Alpine) | Reverse proxy, WebSocket HMR, маршрутизация LAN/Wi-Fi |
| **QA Automation** | Python 3.12, `pytest`, `httpx` | Изолированный Docker test-runner (43 теста) |

---

## 🚀 Быстрый старт

### 1. Запуск через Docker Compose

```bash
# 1. Склонировать репозиторий
git clone https://github.com/C4erries/school.git
cd school

# 2. Скопировать переменные окружения
cp .env.example .env

# 3. Собрать и запустить все контейнеры
docker compose up -d --build

# 4. Проверить статус контейнеров
docker compose ps
```

После старта веб-интерфейс доступен:
- Локально на хосте: **[http://localhost](http://localhost)**
- В локальной сети (Wi-Fi): **`http://<IP-вашего-хоста>/`** (например, `http://192.168.0.21/`)

---

## 🌐 Сетевые порты и эндпоинты

| Сервис | Контейнер | Внутренний порт | Внешний URL | Назначение |
|---|---|---|---|---|
| **Nginx (Reverse Proxy)** | `school_nginx` | 80 | **http://localhost** | Единая точка входа: `/api/v1/` $\rightarrow$ бэкенд, `/` $\rightarrow$ фронтенд |
| **Frontend SPA** | `school_frontend` | 3000 | **http://localhost:3000** | Vite Dev Server с поддержкой HMR |
| **Go Backend API** | `school_backend` | 8080 | **http://localhost:8080** | REST API сервис (`/health`, `/api/v1/health`) |
| **PostgreSQL** | `school_postgres` | 5432 | `localhost:5432` | База данных (`school` / `school_dev_password`) |
| **Valkey** | `school_valkey` | 6379 | `localhost:6379` | Кэш, токены, CSC |
| **MinIO Console** | `school_minio` | 9001 | **http://localhost:9001** | Веб-консоль хранилища (`minioadmin` / `minioadmin`) |
| **MinIO S3** | `school_minio` | 9000 | `localhost:9000` | S3 API |

---

## 🧪 Разработка и тестирование

### Команды Makefile

```bash
# Прогон бэкенд unit-тестов с проверкой гонок (-race)
make test

# Запуск полного сьюта E2E автотестов в Docker (43 теста)
make test-e2e

# Генерация серверных интерфейсов и моделей из OpenAPI
make oapi

# Применение миграций базы данных
make migrate-up

# Проверка линтером бэкенда
make lint
```

### Фронтенд (`frontend/`)

```bash
cd frontend

# Установка зависимостей
npm install

# Проверка типов и production сборка Vite
npm run build

# Линтер ESLint
npm run lint

# Локальный dev-сервер
npm run dev
```

---

## 🌿 Git Workflow

- **Ветка `main`** — стабильная, релизная ветка. Прямые коммиты запрещены.
- **Ветка `dev`** — основная рабочая ветка для разработки фичей и спринтов.
- **Pull Request** — изменения из `dev` попадают в `main` только через проверенный PR после успешного прогона тестов (`make test-e2e`, `npm run lint && npm run build`).

---

## 📚 Навигация по документации (Knowledge Base)

Все ключевые решения, спецификации и правила собраны в каталоге [`docs/`](docs/):

- **[docs/PROJECT_CONTEXT.md](docs/PROJECT_CONTEXT.md)** — **Главный Onboarding Hub**: пивот Tutor Assistant, правила и порядок чтения документации.
- **[docs/CURRENT_SPRINT.md](docs/CURRENT_SPRINT.md)** — **Единственная точка правды** о текущих задачах и бэклоге.
- **[docs/CONVENTIONS.md](docs/CONVENTIONS.md)** — Правила написания кода (Go, React, стиль Apple Liquid Glass, Git).
- **[docs/ROADMAP.md](docs/ROADMAP.md)** — Глобальная дорожная карта платформы по крупным фазам.
- **[docs/decisions/](docs/decisions/)** — Архитектурные решения (ADR-001 ... ADR-008):
  - [ADR-005: Tutor Assistant Pivot](docs/decisions/0005-tutor-assistant-pivot.md)
  - [ADR-006: Единый AppLayout и правила верстки Apple Liquid Glass](docs/decisions/0006-app-layout-and-glass-conventions.md)
  - [ADR-007: Почасовой учет, сетка ставок и форматные абонементы](docs/decisions/0007-hourly-rates-and-format-subscriptions.md)
  - [ADR-008: Декомпозиция монолитного фронтенда и Valkey CSC](docs/decisions/0008-frontend-decomposition-and-valkey-csc.md)
- **[docs/sprints/](docs/sprints/)** — Архив всех завершенных спринтов (0, 1, 2, 2.1, 2.1.2, 2.1.3).
