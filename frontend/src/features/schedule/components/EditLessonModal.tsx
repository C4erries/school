import React, { useState, useEffect } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassInput } from '../../../shared/components/GlassInput';
import { GlassButton } from '../../../shared/components/GlassButton';
import { Classroom, Lesson, SubscriptionFormat } from '../../../types/schedule';
import { updateLesson, cancelLesson, markNoShow } from '../../../api/schedule';
import { Trash2, AlertTriangle } from 'lucide-react';

interface EditLessonModalProps {
  isOpen: boolean;
  onClose: () => void;
  lesson: Lesson | null;
  classrooms: Classroom[];
  clientDisplayName: string;
  onUpdated: () => Promise<void>;
}

export const EditLessonModal: React.FC<EditLessonModalProps> = ({
  isOpen,
  onClose,
  lesson,
  classrooms,
  clientDisplayName,
  onUpdated,
}) => {
  const [editTitle, setEditTitle] = useState('');
  const [editDate, setEditDate] = useState('');
  const [editStartTime, setEditStartTime] = useState('');
  const [editEndTime, setEditEndTime] = useState('');
  const [editFormat, setEditFormat] = useState<SubscriptionFormat>('individual');
  const [editLocationType, setEditLocationType] = useState<'offline' | 'online'>('offline');
  const [editClassroomId, setEditClassroomId] = useState('');
  const [editOnlineLink, setEditOnlineLink] = useState('');
  const [editComment, setEditComment] = useState('');
  const [cancelReason, setCancelReason] = useState('');
  const [isCancelling, setIsCancelling] = useState(false);
  const [isSubmittingEdit, setIsSubmittingEdit] = useState(false);

  useEffect(() => {
    if (lesson && isOpen) {
      setEditTitle(lesson.title);
      const start = new Date(lesson.start_time);
      const end = new Date(lesson.end_time);
      setEditDate(start.toISOString().split('T')[0]);
      setEditStartTime(
        `${String(start.getHours()).padStart(2, '0')}:${String(start.getMinutes()).padStart(2, '0')}`
      );
      setEditEndTime(
        `${String(end.getHours()).padStart(2, '0')}:${String(end.getMinutes()).padStart(2, '0')}`
      );
      const fmt: SubscriptionFormat =
        lesson.format === 'pair' || lesson.format === 'group' ? lesson.format : 'individual';
      setEditFormat(fmt);

      // Task 0: определение онлайн урока
      const isOnline =
        Boolean(lesson.online_link) ||
        Boolean(
          lesson.location_or_url &&
            (lesson.location_or_url.startsWith('http') || !lesson.classroom_id)
        ) ||
        (lesson.format as string) === 'online';

      setEditLocationType(isOnline ? 'online' : 'offline');
      setEditClassroomId(lesson.classroom_id || '');
      setEditOnlineLink(
        lesson.online_link ||
          (lesson.location_or_url?.startsWith('http') ? lesson.location_or_url : '') ||
          ''
      );
      setEditComment(lesson.comment || '');
      setCancelReason('');
      setIsCancelling(false);
    }
  }, [lesson, isOpen]);

  const handleSaveEdit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!lesson) return;

    setIsSubmittingEdit(true);
    try {
      const [startH, startM] = editStartTime.split(':').map(Number);
      const [endH, endM] = editEndTime.split(':').map(Number);
      const start = new Date(editDate);
      start.setHours(startH, startM, 0, 0);
      const end = new Date(editDate);
      end.setHours(endH, endM, 0, 0);

      const isOnline = editLocationType === 'online';

      await updateLesson(lesson.id, {
        title: editTitle.trim(),
        start_time: start.toISOString(),
        end_time: end.toISOString(),
        format: editFormat,
        classroom_id: isOnline ? null : editClassroomId || null,
        online_link: isOnline ? editOnlineLink : undefined,
        comment: editComment.trim() || undefined,
      });

      await onUpdated();
      onClose();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка сохранения занятия');
    } finally {
      setIsSubmittingEdit(false);
    }
  };

  const handleCancelLesson = async () => {
    if (!lesson) return;
    try {
      await cancelLesson(lesson.id, cancelReason.trim() || undefined);
      await onUpdated();
      onClose();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка отмены занятия');
    }
  };

  const handleNoShow = async () => {
    if (!lesson) return;
    if (!window.confirm('Отметить неявку ученика на занятие?')) return;
    try {
      await markNoShow(lesson.id);
      await onUpdated();
      onClose();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка отметки неявки');
    }
  };

  if (!lesson) return null;

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title={`Урок: ${clientDisplayName}`}
      description="Измените время, аудиторию или отмените занятие."
      maxWidth="lg"
    >
      <form onSubmit={handleSaveEdit} className="space-y-4">
        <GlassInput
          label="Тема занятия"
          value={editTitle}
          onChange={(e) => setEditTitle(e.target.value)}
          required
        />

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <GlassInput
            label="Дата"
            type="date"
            value={editDate}
            onChange={(e) => setEditDate(e.target.value)}
            required
          />
          <GlassInput
            label="Начало"
            type="time"
            value={editStartTime}
            onChange={(e) => setEditStartTime(e.target.value)}
            required
          />
          <GlassInput
            label="Конец"
            type="time"
            value={editEndTime}
            onChange={(e) => setEditEndTime(e.target.value)}
            required
          />
        </div>

        {/* Формат занятия */}
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
            Формат занятия
          </label>
          <div className="grid grid-cols-3 gap-2 p-1 rounded-2xl bg-black/[0.04]">
            <button
              type="button"
              onClick={() => setEditFormat('individual')}
              className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                editFormat === 'individual'
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Индивидуально
            </button>
            <button
              type="button"
              onClick={() => setEditFormat('pair')}
              className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                editFormat === 'pair'
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              В паре
            </button>
            <button
              type="button"
              onClick={() => setEditFormat('group')}
              className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                editFormat === 'group'
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
              onClick={() => setEditLocationType('offline')}
              className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                editLocationType === 'offline'
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Оффлайн в кабинете
            </button>
            <button
              type="button"
              onClick={() => setEditLocationType('online')}
              className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                editLocationType === 'online'
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Онлайн урок
            </button>
          </div>
        </div>

        {editLocationType === 'offline' ? (
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
              Кабинет
            </label>
            <select
              value={editClassroomId}
              onChange={(e) => setEditClassroomId(e.target.value)}
              className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 bg-white/70 border border-slate-200 focus:outline-none focus:border-indigo-500"
            >
              <option value="">Без закрепления кабинета</option>
              {classrooms.map((room) => (
                <option key={room.id} value={room.id}>
                  {room.name}
                </option>
              ))}
            </select>
          </div>
        ) : (
          <GlassInput
            label="Ссылка на созвон"
            placeholder="https://telemost.yandex.ru/j/..."
            value={editOnlineLink}
            onChange={(e) => setEditOnlineLink(e.target.value)}
          />
        )}

        <GlassInput
          label="Заметка"
          value={editComment}
          onChange={(e) => setEditComment(e.target.value)}
        />

        {/* Блок отмены урока */}
        {lesson.status === 'scheduled' && (
          <div className="p-3.5 rounded-2xl bg-rose-500/10 border border-rose-500/20 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold text-rose-800 flex items-center gap-1.5">
                <Trash2 className="w-3.5 h-3.5" /> Отмена занятия
              </span>
              {!isCancelling && (
                <div className="flex items-center gap-3">
                  <button
                    type="button"
                    onClick={handleNoShow}
                    className="text-xs font-semibold text-amber-700 hover:text-amber-800 flex items-center gap-1"
                  >
                    <AlertTriangle className="w-3.5 h-3.5" />
                    Неявка
                  </button>
                  <button
                    type="button"
                    onClick={() => setIsCancelling(true)}
                    className="text-xs font-semibold text-rose-600 hover:text-rose-700 underline"
                  >
                    Отменить этот урок
                  </button>
                </div>
              )}
            </div>

            {isCancelling && (
              <div className="space-y-2 pt-1 animate-in fade-in">
                <input
                  type="text"
                  placeholder="Укажите причину отмены (необязательно)"
                  value={cancelReason}
                  onChange={(e) => setCancelReason(e.target.value)}
                  className="w-full rounded-xl px-3 py-2 text-xs text-slate-800 bg-white/80 border border-rose-200 focus:outline-none"
                />
                <div className="flex justify-end gap-2">
                  <button
                    type="button"
                    onClick={() => setIsCancelling(false)}
                    className="px-3 py-1 rounded-lg text-xs font-medium text-slate-600 hover:bg-black/5"
                  >
                    Назад
                  </button>
                  <button
                    type="button"
                    onClick={handleCancelLesson}
                    className="px-3 py-1 rounded-lg text-xs font-bold text-white bg-rose-500 hover:bg-rose-600 shadow-sm"
                  >
                    Подтвердить отмену
                  </button>
                </div>
              </div>
            )}
          </div>
        )}

        <div className="pt-3 flex justify-end gap-3 border-t border-black/[0.05]">
          <GlassButton
            type="button"
            variant="secondary"
            onClick={onClose}
          >
            Закрыть
          </GlassButton>
          <GlassButton
            type="submit"
            variant="primary"
            isLoading={isSubmittingEdit}
          >
            Сохранить
          </GlassButton>
        </div>
      </form>
    </GlassModal>
  );
};
