# School

Платформа для онлайн/оффлайн школы: видео-уроки, проверка заданий, расписание, финансы.

## Стек

- **Backend**: Go (DDD)
- **Frontend**: React + TypeScript
- **БД**: PostgreSQL + Redis
- **Файлы**: MinIO (S3-compatible)
- **Инфраструктура**: Docker Compose + Nginx

## Быстрый старт

```bash
# Скопировать переменные окружения
cp .env.example .env

# Запустить всё
docker compose up -d

# Открыть
open http://localhost
```

> ⚠️ Проект на стадии Фазы 0 — в разработке.

## Документация

| Документ | Описание |
|----------|----------|
| [Видение проекта](docs/PROJECT_VISION.md) | Что строим и зачем |
| [Roadmap](docs/ROADMAP.md) | Глобальный план по фазам |
| [Текущий спринт](docs/CURRENT_SPRINT.md) | Что делаем сейчас |
| [Конвенции](docs/CONVENTIONS.md) | Как пишем код |
| [ADR](docs/decisions/) | Архитектурные решения |

## Структура проекта

```
school/
├── backend/          # Go приложение
├── frontend/         # React приложение
├── deploy/           # Docker, nginx, compose
├── docs/             # Документация
├── scripts/          # Утилиты
├── docker-compose.yml
└── README.md
```
