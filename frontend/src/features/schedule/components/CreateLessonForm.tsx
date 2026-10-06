import React from 'react';
import { GlassInput } from '../../../shared/components/GlassInput';
import { Classroom, Client, SubscriptionFormat } from '../../../types/schedule';
import { RecurrenceSettings } from './RecurrenceSettings';

interface CreateLessonFormProps {
  clients: Client[];
  classrooms: Classroom[];
  selectedClientId: string;
  onSelectClient: (id: string) => void;
  lessonTitle: string;
  onChangeTitle: (title: string) => void;
  lessonDate: string;
  onChangeDate: (date: string) => void;
  lessonStartTime: string;
  onChangeStartTime: (time: string) => void;
  lessonDuration: string;
  onChangeDuration: (dur: string) => void;
  allStartTimeOptions: string[];
  selectedFormat: SubscriptionFormat;
  onChangeFormat: (fmt: SubscriptionFormat) => void;
  locationType: 'offline' | 'online';
  onChangeLocationType: (loc: 'offline' | 'online') => void;
  selectedClassroomId: string;
  onChangeClassroomId: (id: string) => void;
  onlineLink: string;
  onChangeOnlineLink: (link: string) => void;
  lessonComment: string;
  onChangeComment: (comment: string) => void;
  // Recurrence
  isRecurring: boolean;
  onToggleRecurring: (val: boolean) => void;
  selectedDays: string[];
  onToggleDay: (day: string) => void;
  untilMode: 'never' | 'date';
  onChangeUntilMode: (mode: 'never' | 'date') => void;
  untilDate: string;
  onChangeUntilDate: (date: string) => void;
}

export const CreateLessonForm: React.FC<CreateLessonFormProps> = ({
  clients,
  classrooms,
  selectedClientId,
  onSelectClient,
  lessonTitle,
  onChangeTitle,
  lessonDate,
  onChangeDate,
  lessonStartTime,
  onChangeStartTime,
  lessonDuration,
  onChangeDuration,
  allStartTimeOptions,
  selectedFormat,
  onChangeFormat,
  locationType,
  onChangeLocationType,
  selectedClassroomId,
  onChangeClassroomId,
  onlineLink,
  onChangeOnlineLink,
  lessonComment,
  onChangeComment,
  isRecurring,
  onToggleRecurring,
  selectedDays,
  onToggleDay,
  untilMode,
  onChangeUntilMode,
  untilDate,
  onChangeUntilDate,
}) => {
  return (
    <div className="space-y-4">
      {/* Выбор ученика */}
      <div>
        <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
          Ученик *
        </label>
        <select
          value={selectedClientId}
          onChange={(e) => onSelectClient(e.target.value)}
          className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 bg-white/70 border border-slate-200 focus:outline-none focus:border-indigo-500"
          required
        >
          <option value="">-- Выберите ученика --</option>
          {clients.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name} ({c.balance ?? 0} ч на балансе)
            </option>
          ))}
        </select>
      </div>

      {/* Тема урока */}
      <GlassInput
        label="Тема / Название занятия *"
        placeholder="например: Подготовка к ОГЭ: Алгебра"
        value={lessonTitle}
        onChange={(e) => onChangeTitle(e.target.value)}
        required
      />

      {/* Дата и время */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <GlassInput
          label="Дата *"
          type="date"
          value={lessonDate}
          onChange={(e) => onChangeDate(e.target.value)}
          required
        />
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
            Время начала *
          </label>
          <select
            value={lessonStartTime}
            onChange={(e) => onChangeStartTime(e.target.value)}
            className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 bg-white/70 border border-slate-200 focus:outline-none focus:border-indigo-500"
            required
          >
            {allStartTimeOptions.map((t) => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
            Длительность
          </label>
          <select
            value={lessonDuration}
            onChange={(e) => onChangeDuration(e.target.value)}
            className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 bg-white/70 border border-slate-200 focus:outline-none focus:border-indigo-500"
          >
            <option value="30">30 минут (0.5 ч)</option>
            <option value="45">45 минут (0.75 ч)</option>
            <option value="60">60 минут (1.0 ч)</option>
            <option value="90">90 минут (1.5 ч)</option>
            <option value="120">120 минут (2.0 ч)</option>
            <option value="150">150 минут (2.5 ч)</option>
            <option value="180">180 минут (3.0 ч)</option>
            <option value="240">240 минут (4.0 ч)</option>
            {!['30', '45', '60', '90', '120', '150', '180', '240'].includes(lessonDuration) && (
              <option value={lessonDuration}>
                {lessonDuration} минут ({(Number(lessonDuration) / 60).toFixed(1)} ч)
              </option>
            )}
          </select>
        </div>
      </div>

      {/* Регулярность (Повторять еженедельно) */}
      <RecurrenceSettings
        isRecurring={isRecurring}
        onToggleRecurring={onToggleRecurring}
        selectedDays={selectedDays}
        onToggleDay={onToggleDay}
        untilMode={untilMode}
        onChangeUntilMode={onChangeUntilMode}
        untilDate={untilDate}
        onChangeUntilDate={onChangeUntilDate}
      />

      {/* Формат занятия */}
      <div>
        <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
          Формат занятия
        </label>
        <div className="grid grid-cols-3 gap-2 p-1 rounded-2xl bg-black/[0.04]">
          <button
            type="button"
            onClick={() => onChangeFormat('individual')}
            className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
              selectedFormat === 'individual'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            Индивидуально
          </button>
          <button
            type="button"
            onClick={() => onChangeFormat('pair')}
            className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
              selectedFormat === 'pair'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            В паре
          </button>
          <button
            type="button"
            onClick={() => onChangeFormat('group')}
            className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
              selectedFormat === 'group'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            В группе
          </button>
        </div>
      </div>

      {/* Локация проведения */}
      <div>
        <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
          Локация проведения
        </label>
        <div className="grid grid-cols-2 gap-2 p-1 rounded-2xl bg-black/[0.04]">
          <button
            type="button"
            onClick={() => onChangeLocationType('offline')}
            className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
              locationType === 'offline'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            Оффлайн в школе
          </button>
          <button
            type="button"
            onClick={() => onChangeLocationType('online')}
            className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
              locationType === 'online'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            Онлайн урок
          </button>
        </div>
      </div>

      {/* Кабинет (если оффлайн) */}
      {locationType === 'offline' && (
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
            Аудитория школы
          </label>
          <select
            value={selectedClassroomId}
            onChange={(e) => onChangeClassroomId(e.target.value)}
            className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 bg-white/70 border border-slate-200 focus:outline-none focus:border-indigo-500"
          >
            <option value="">Без закрепления кабинета</option>
            {classrooms.map((room) => (
              <option key={room.id} value={room.id}>
                {room.name} (вместимость: {room.capacity})
              </option>
            ))}
          </select>
        </div>
      )}

      {/* Ссылка (если онлайн) */}
      {locationType === 'online' && (
        <GlassInput
          label="Ссылка на видеозвонок"
          placeholder="https://telemost.yandex.ru/j/..."
          value={onlineLink}
          onChange={(e) => onChangeOnlineLink(e.target.value)}
        />
      )}

      <GlassInput
        label="Заметка / ДЗ (необязательно)"
        placeholder="например: Проверить вариант №4"
        value={lessonComment}
        onChange={(e) => onChangeComment(e.target.value)}
      />
    </div>
  );
};
