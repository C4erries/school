# Design System: Apple Liquid Glass

> Спецификация UI/UX дизайна, дизайн-токенов и визуального стиля платформы School.

---

## 1. Философия стиля

Дизайн платформы вдохновлен эстетикой **Apple visionOS / macOS Sonoma / iOS 18**:
* **Liquid Frosted Glass (Жидкое стекло)**: полупрозрачные панели с глубоким оптическим размытием (`backdrop-filter: blur(20px)`), сквозь которые мягко просвечивают фоновые цвета.
* **Световые грани (Specular Hairline)**: ультратонкие границы элементов `1px solid rgba(255, 255, 255, 0.45)`, имитирующие преломление света на полированном крае стекла.
* **Мягкие диффузные тени (Ambient Shadows)**: многослойные воздушные тени без резких черных краев — элементы "парят" над холстом.
* **Apple Squircle (Скругления)**: крупные, гармоничные скругления (`rounded-2xl` — 16px для карточек, `rounded-3xl` — 24px для модалок, `rounded-full` для бейджей и кнопок действий).
* **Типографика**: чистый, читаемый неогротеск с системным сглаживанием шрифтов (`Inter`, `SF Pro Display`, `-apple-system`).

---

## 2. Цветовая палитра (Design Tokens)

Вместо резких и агрессивных цветов интерфейс использует спокойные, природные оттенки Apple:

| Токен | Название | HEX | Назначение |
| :--- | :--- | :--- | :--- |
| `primary` | **Apple Electric Indigo** | `#4F46E5` | Основной цвет действий, активные табы, кнопки |
| `primary-hover` | **Deep Indigo** | `#4338CA` | Состояние hover для первичных элементов |
| `mint` | **Apple Mint** | `#10B981` | Успех: "Сдано", "Оплачено", "Проведено", бейджи |
| `mint-light` | **Mint Surface** | `rgba(16, 185, 129, 0.12)` | Фон бейджей и уведомлений об успехе |
| `amber` | **Warm Amber** | `#F59E0B` | Ожидание/внимание: "На проверке", "Перенос", слоты |
| `amber-light` | **Amber Surface** | `rgba(245, 158, 11, 0.12)` | Фон предупреждающих бейджей |
| `coral` | **Coral Sunset** | `#F43F5E` | Ошибки/отмены: "На доработку", "Долг", "Отмена" |
| `coral-light` | **Coral Surface** | `rgba(244, 63, 94, 0.12)` | Фон ошибок и деструктивных действий |
| `canvas` | **Apple Canvas Off-White** | `#F5F5F7` | Тёплый фоновый цвет рабочей области |
| `surface-glass`| **Liquid Glass Panel** | `rgba(255, 255, 255, 0.72)` | Стекло карточек и панелей в светлой теме |
| `surface-border`| **Glass Border** | `rgba(255, 255, 255, 0.45)` | Тонкая световая фаска стекла |

---

## 3. Базовые CSS-классы Liquid Glass

```css
/* Базовая стеклянная панель */
.liquid-glass {
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.5);
  box-shadow: 0 10px 30px -5px rgba(0, 0, 0, 0.05),
              0 1px 3px 0 rgba(0, 0, 0, 0.02);
}

/* Интерактивная стеклянная карточка с эффектом при наведении */
.liquid-glass-interactive {
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.5);
  box-shadow: 0 10px 30px -5px rgba(0, 0, 0, 0.05);
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.liquid-glass-interactive:hover {
  background: rgba(255, 255, 255, 0.88);
  transform: translateY(-2px);
  box-shadow: 0 20px 40px -10px rgba(0, 0, 0, 0.08),
              0 1px 4px 0 rgba(0, 0, 0, 0.03);
  border-color: rgba(255, 255, 255, 0.7);
}

/* Стеклянные инпуты */
.liquid-glass-input {
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  transition: all 0.2s ease;
}

.liquid-glass-input:focus {
  background: rgba(255, 255, 255, 0.95);
  border-color: #4F46E5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.15);
  outline: none;
}
```

---

## 4. UI Стек фронтенда

* **Стилизация**: Tailwind CSS (утилитарный CSS + кастомная палитра Apple Liquid Glass).
* **Иконки**: `lucide-react` (минималистичные контурные иконки в духе Apple SF Symbols).
* **Примитивы доступности**: Radix UI (headless диалоги, тултипы, поповеры, дропдауны).
* **Утилиты объединения классов**: `clsx`, `tailwind-merge`.

---

## 5. Графика и медиа-ассеты

* **Фоны и баннеры курсов**: качественные ландшафтные и архитектурные фотографии с Unsplash (спокойные градиенты, горы, геометрия).
* **Аватары**:
  * Загрузка пользовательских фото (квадрат 1:1, отображение в круглом маске `rounded-full`).
  * Дефолтные плейсхолдеры: мягкие градиентные круги с монохромными инициалами ученика/преподавателя.
