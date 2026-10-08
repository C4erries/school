# ADR 013: Дневник занятий (Lesson Journal) и Управление домашними заданиями (Homework Management)

## Статус
Принято (Accepted)

## Контекст и Проблематика
1. **Потребность преподавателя в фиксации учебного контента**:
   В Фазах 2.1–2.3 был полностью построен фундамент расписания, абонементов, финансовой аналитики и циклических серий (RFC 5545 RRULE). Однако репетитору критически важно вести содержательную сторону обучения:
   - Фиксировать пройденные темы по каждому ученику;
   - Отмечать уровень усвоения материала (простая и быстрая шкала оценки понимания 1–5 без бюрократии);
   - Вести приватные заметки по ходу урока (слабые места, психологические особенности, пробелы в знаниях);
   - Задавать домашние задания и контролировать их выполнение к следующему уроку.
2. **Разделение жизненных циклов сущностей**:
   - **Урок (`Lesson`)**: календарно-пространственный слот (время, кабинет, формат, факт присутствия).
   - **Отчет по уроку (`LessonJournal`)**: дидактический срез конкретного занятия (1:1 к уроку).
   - **Домашнее задание (`HomeworkAssignment`)**: самостоятельная задача ученика, которая задается на одном занятии, выполняется к определенной дате и имеет собственный жизненный цикл (`assigned` → `completed` / `not_done`).
3. **UX-требования к оперативности работы**:
   - Быстрый темп работы репетитора (перерыв между уроками 5–10 минут) требует, чтобы проведение урока (`✓`) оставалось мгновенным в 1 клик без принудительного открытия тяжелых форм.
   - Дневник урока должен открываться отдельной кнопкой на карточке занятия в любой момент времени (до урока для планирования, во время урока или после него).
   - В карточке ученика в CRM должна быть доступна сквозная ретроспектива: хронологический таймлайн тем и история домашних заданий со статусами сдачи.

---

## Решение

### 1. Архитектурная модель данных: Двухуровневый академический хаб (Вариант 3)

Вводятся две специализированные таблицы с поддержкой изоляции репетиторов и каскадной целостности:

```sql
-- 1. Дневник урока (содержательный отчет по проведенному занятию)
CREATE TABLE IF NOT EXISTS lesson_journals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id UUID NOT NULL UNIQUE REFERENCES lessons(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic VARCHAR(255) NOT NULL,
    notes TEXT,
    performance_score INT CHECK (performance_score BETWEEN 1 AND 5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lesson_journals_lesson_id ON lesson_journals (lesson_id);
CREATE INDEX IF NOT EXISTS idx_lesson_journals_client_id ON lesson_journals (client_id);
CREATE INDEX IF NOT EXISTS idx_lesson_journals_teacher_id ON lesson_journals (teacher_id);

-- 2. Домашние задания (задачи ученика с жизненным циклом и дедлайном)
CREATE TABLE IF NOT EXISTS homework_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_lesson_id UUID REFERENCES lessons(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    due_date DATE,
    status VARCHAR(50) NOT NULL DEFAULT 'assigned' 
        CHECK (status IN ('assigned', 'completed', 'not_done')),
    review_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_homework_assignments_client_id ON homework_assignments (client_id);
CREATE INDEX IF NOT EXISTS idx_homework_assignments_teacher_id ON homework_assignments (teacher_id);
CREATE INDEX IF NOT EXISTS idx_homework_assignments_assigned_lesson_id ON homework_assignments (assigned_lesson_id);
CREATE INDEX IF NOT EXISTS idx_homework_assignments_status ON homework_assignments (status);
```

### 2. Взаимодействие с виртуальными слотами (`LessonSeries` по ADR-012)
- Если преподаватель открывает и сохраняет дневник для виртуального занятия регулярной серии:
  - Сервис расписания вычисляет детерминированный `VirtualLessonID(series.ID, slot.OriginalStartTime)`.
  - Урок **автоматически материализуется** в таблице `lessons` (с заполнением `series_id` и `original_start_time`).
  - Создается связанная запись `lesson_journals(lesson_id)`.
  - Реляционная целостность `FOREIGN KEY` гарантируется без разрыва ссылок.

### 3. Пользовательский интерфейс (UX)
1. **Расписание (Интерактивный календарь)**:
   - Карточка урока получает отдельную аккуратную кнопку «Дневник / ДЗ» (иконка `BookOpen`).
   - Кнопка доступна всегда независимо от статуса проведения (`scheduled` или `completed`).
   - Обычный клик по галочке `✓` моментально завершает занятие и списывает абонемент (никаких блокирующих попапов).
   - Клик по кнопке «Дневник» открывает `LessonJournalModal`:
     * Тема занятия (input).
     * Оценка понимания темы (легкие кликабельные чипы `[1] [2] [3] [4] [5]` с цветовой индикацией).
     * Заметки репетитора (textarea).
     * Быстрое назначение ДЗ (заголовок, описание, дедлайн к следующему занятию).
     * Экспресс-проверка предыдущего ДЗ ученика (смена статуса «Сдано» / «Не сделано» в 1 клик прямо в окне текущего урока).
2. **Карточка ученика (CRM)**:
   - В карточке ученика (`ClientCard.tsx`) добавляется кнопка «Дневник & ДЗ».
   - Открывает модальную панель со сквозной историей:
     * Вкладка **«История уроков»**: хронологический список пройденных тем с оценками понимания и заметками преподавателя.
     * Вкладка **«Домашние задания»**: список всех ДЗ с фильтрами («Все», «В работе», «Сдано», «Не выполнено») и кнопками быстрой смены статуса.

### 4. Подготовка к будущим фазам развития платформы
- **Фаза 4 (Личный кабинет ученика)**: ученик сможет видеть свои выданные ДЗ и темы уроков без изменений схемы БД.
- **Фаза 5 (Проверка домашних заданий)**: таблица `homework_assignments` легко расширяется таблицей прикрепленных фотографий тетрадей (`homework_submissions`), сохраняя преемственность данных.

---

## Последствия
- Репетитор получает профессиональный инструмент для систематизации обучения и подготовки к занятиям.
- Сохранена высокая скорость проведения занятий в расписании без навязчивых модальных окон.
- Архитектура чистая, соблюдает SRP и правила DDD.
- 100% совместимость с существующей системой виртуальных слотов и абонементов.

