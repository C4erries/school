import React, { useState, useEffect } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassInput } from '../../../shared/components/GlassInput';
import { GlassButton } from '../../../shared/components/GlassButton';
import { Classroom, Client, SubscriptionFormat } from '../../../types/schedule';
import { createLesson } from '../../../api/schedule';

interface CreateLessonModalProps {
  isOpen: boolean;
  onClose: () => void;
  clients: Client[];
  classrooms: Classroom[];
  initialClientId?: string;
  initialDate?: string;
  initialStartTime?: string;
  initialDuration?: string;
  initialTitle?: string;
  onCreated: () => Promise<void>;
}

export const CreateLessonModal: React.FC<CreateLessonModalProps> = ({
  isOpen,
  onClose,
  clients,
  classrooms,
  initialClientId = '',
  initialDate = new Date().toISOString().split('T')[0],
  initialStartTime = '14:00',
  initialDuration = '60',
  initialTitle = '',
  onCreated,
}) => {
  const [selectedClientId, setSelectedClientId] = useState(initialClientId);
  const [lessonTitle, setLessonTitle] = useState(initialTitle);
  const [lessonDate, setLessonDate] = useState(initialDate);
  const [lessonStartTime, setLessonStartTime] = useState(initialStartTime);
  const [lessonDuration, setLessonDuration] = useState(initialDuration);
  const [selectedFormat, setSelectedFormat] = useState<SubscriptionFormat>('individual');
  const [locationType, setLocationType] = useState<'offline' | 'online'>('offline');
  const [selectedClassroomId, setSelectedClassroomId] = useState('');
  const [onlineLink, setOnlineLink] = useState('https://telemost.yandex.ru/j/school-lesson');
  const [lessonComment, setLessonComment] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setSelectedClientId(initialClientId);
      setLessonTitle(initialTitle);
      setLessonDate(initialDate);
      setLessonStartTime(initialStartTime);
      setLessonDuration(initialDuration);
      setSelectedFormat('individual');
      setLocationType('offline');
      setSelectedClassroomId('');
      setLessonComment('');
    }
  }, [isOpen, initialClientId, initialTitle, initialDate, initialStartTime, initialDuration]);

  const handleCreateLesson = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedClientId || !lessonTitle.trim()) return;

    setIsSubmitting(true);
    try {
      const [hoursVal, minsVal] = lessonStartTime.split(':').map(Number);
      const start = new Date(lessonDate);
      start.setHours(hoursVal, minsVal, 0, 0);

      const end = new Date(start.getTime() + Number(lessonDuration) * 60 * 1000);

      await createLesson({
        client_id: selectedClientId,
        title: lessonTitle.trim(),
        format: selectedFormat,
        classroom_id: locationType === 'offline' ? selectedClassroomId || null : null,
        online_link: locationType === 'online' ? onlineLink : undefined,
        comment: lessonComment.trim() || undefined,
        start_time: start.toISOString(),
        end_time: end.toISOString(),
      });

      await onCreated();
      onClose();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка создания занятия');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title="Назначить занятие"
      description="Выберите ученика, дату, формат и кабинет школы"
      maxWidth="lg"
    >
      <form onSubmit={handleCreateLesson} className="space-y-4">
        {/* Выбор ученика */}
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
            Ученик *
          </label>
          <select
            value={selectedClientId}
            onChange={(e) => {
              setSelectedClientId(e.target.value);
              const found = clients.find((c) => c.id === e.target.value);
              if (found && (!lessonTitle || lessonTitle.startsWith('Урок'))) {
                setLessonTitle(`Урок: ${found.name}`);
              }
            }}
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
          onChange={(e) => setLessonTitle(e.target.value)}
          required
        />

        {/* Дата и время */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <GlassInput
            label="Дата *"
            type="date"
            value={lessonDate}
            onChange={(e) => setLessonDate(e.target.value)}
            required
          />
          <GlassInput
            label="Время начала *"
            type="time"
            value={lessonStartTime}
            onChange={(e) => setLessonStartTime(e.target.value)}
            required
          />
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
              Длительность
            </label>
            <select
              value={lessonDuration}
              onChange={(e) => setLessonDuration(e.target.value)}
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

        {/* Формат занятия */}
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
            Формат занятия
          </label>
          <div className="grid grid-cols-3 gap-2 p-1 rounded-2xl bg-black/[0.04]">
            <button
              type="button"
              onClick={() => setSelectedFormat('individual')}
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
              onClick={() => setSelectedFormat('pair')}
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
              onClick={() => setSelectedFormat('group')}
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
              onClick={() => setLocationType('offline')}
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
              onClick={() => setLocationType('online')}
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
              onChange={(e) => setSelectedClassroomId(e.target.value)}
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
            onChange={(e) => setOnlineLink(e.target.value)}
          />
        )}

        <GlassInput
          label="Заметка / ДЗ (необязательно)"
          placeholder="например: Проверить вариант №4"
          value={lessonComment}
          onChange={(e) => setLessonComment(e.target.value)}
        />

        <div className="pt-3 flex justify-end gap-3 border-t border-black/[0.05]">
          <GlassButton
            type="button"
            variant="secondary"
            onClick={onClose}
          >
            Отмена
          </GlassButton>
          <GlassButton
            type="submit"
            variant="primary"
            isLoading={isSubmitting}
          >
            Запланировать урок
          </GlassButton>
        </div>
      </form>
    </GlassModal>
  );
};
