import React, { useState, useEffect, useCallback } from 'react';
import { HomeworkAssignment, HomeworkStatus } from '../../../../types/journal';
import {
  listClientHomework,
  createHomework,
  updateHomeworkStatus,
  deleteHomework,
} from '../../../../api/journal';
import { GlassButton } from '../../../../shared/components/GlassButton';
import { ClientHomeworkItem } from './ClientHomeworkItem';
import {
  Plus,
  Loader2,
  CheckSquare,
  AlertCircle,
  ChevronUp,
} from 'lucide-react';

interface ClientHomeworkListProps {
  clientId: string;
}

type FilterTab = 'all' | HomeworkStatus;

export const ClientHomeworkList: React.FC<ClientHomeworkListProps> = ({
  clientId,
}) => {
  const [filter, setFilter] = useState<FilterTab>('all');
  const [homeworks, setHomeworks] = useState<HomeworkAssignment[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [updatingId, setUpdatingId] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Форма добавления нового ДЗ
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [newTitle, setNewTitle] = useState('');
  const [newDueDate, setNewDueDate] = useState('');
  const [newDescription, setNewDescription] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const loadHomework = useCallback(async () => {
    setIsLoading(true);
    setErrorMessage(null);
    try {
      const data = await listClientHomework(
        clientId,
        filter === 'all' ? undefined : filter
      );
      setHomeworks(data);
    } catch (err: unknown) {
      setErrorMessage(
        err instanceof Error ? err.message : 'Ошибка загрузки заданий'
      );
    } finally {
      setIsLoading(false);
    }
  }, [clientId, filter]);

  useEffect(() => {
    loadHomework();
  }, [loadHomework]);

  const handleStatusChange = async (homeworkId: string, status: HomeworkStatus) => {
    setUpdatingId(homeworkId);
    try {
      const updated = await updateHomeworkStatus(homeworkId, { status });
      setHomeworks((prev) =>
        prev.map((item) => (item.id === homeworkId ? updated : item))
      );
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка смены статуса');
    } finally {
      setUpdatingId(null);
    }
  };

  const handleDelete = async (homeworkId: string) => {
    if (!window.confirm('Удалить это домашнее задание?')) return;
    setUpdatingId(homeworkId);
    try {
      await deleteHomework(homeworkId);
      setHomeworks((prev) => prev.filter((item) => item.id !== homeworkId));
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка удаления задания');
    } finally {
      setUpdatingId(null);
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTitle.trim()) return;

    setIsSubmitting(true);
    try {
      const created = await createHomework(clientId, {
        title: newTitle.trim(),
        description: newDescription.trim() || null,
        due_date: newDueDate || null,
      });
      setHomeworks((prev) => [created, ...prev]);
      setNewTitle('');
      setNewDueDate('');
      setNewDescription('');
      setIsFormOpen(false);
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка создания задания');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="space-y-4">
      {/* Кнопка создания нового ДЗ и фильтры */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-2.5">
        {/* Фильтры */}
        <div className="flex items-center p-1 rounded-2xl bg-black/[0.04] backdrop-blur-md border border-white/40 gap-1 overflow-x-auto">
          {(
            [
              { key: 'all', label: 'Все' },
              { key: 'assigned', label: 'В работе' },
              { key: 'completed', label: 'Выполнено' },
              { key: 'not_done', label: 'Не сдано' },
            ] as const
          ).map((t) => (
            <button
              key={t.key}
              type="button"
              onClick={() => setFilter(t.key)}
              className={`px-2.5 py-1 rounded-xl text-xs font-semibold whitespace-nowrap transition-all ${
                filter === t.key
                  ? 'bg-white/80 shadow-xs text-indigo-900 border border-white/80'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-white/30'
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>

        {/* Кнопка добавления ДЗ */}
        <GlassButton
          variant="secondary"
          size="sm"
          onClick={() => setIsFormOpen((prev) => !prev)}
          icon={isFormOpen ? <ChevronUp className="w-3.5 h-3.5" /> : <Plus className="w-3.5 h-3.5 text-indigo-600" />}
        >
          {isFormOpen ? 'Закрыть форму' : 'Выдать ДЗ'}
        </GlassButton>
      </div>

      {/* Форма создания задания */}
      {isFormOpen && (
        <form
          onSubmit={handleCreate}
          className="p-4 rounded-2xl bg-white/40 backdrop-blur-md border border-white/60 shadow-xs space-y-3 animate-in fade-in"
        >
          <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-700">
            <Plus className="w-3.5 h-3.5 text-indigo-600" />
            <span>Новое домашнее задание</span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
            <div className="sm:col-span-2 space-y-1">
              <label className="block text-[11px] font-semibold text-slate-600 ml-1">
                Заголовок / Что задано *
              </label>
              <input
                type="text"
                required
                placeholder="Например: Прорешать вариант ОГЭ № 4"
                value={newTitle}
                onChange={(e) => setNewTitle(e.target.value)}
                className="w-full rounded-xl px-3 py-2 text-xs text-slate-800 liquid-glass-input"
              />
            </div>

            <div className="space-y-1">
              <label className="block text-[11px] font-semibold text-slate-600 ml-1">
                Срок сдачи
              </label>
              <input
                type="date"
                value={newDueDate}
                onChange={(e) => setNewDueDate(e.target.value)}
                className="w-full rounded-xl px-2.5 py-2 text-xs text-slate-800 liquid-glass-input font-mono"
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="block text-[11px] font-semibold text-slate-600 ml-1">
              Пояснения и критерии (опционально)
            </label>
            <textarea
              rows={2}
              placeholder="Формулы расписать подробно, прислать фото до вечера пятницы..."
              value={newDescription}
              onChange={(e) => setNewDescription(e.target.value)}
              className="w-full rounded-xl p-2.5 text-xs text-slate-800 liquid-glass-input resize-none"
            />
          </div>

          <div className="flex justify-end gap-2 pt-1">
            <GlassButton
              type="button"
              variant="secondary"
              size="sm"
              onClick={() => setIsFormOpen(false)}
            >
              Отмена
            </GlassButton>
            <GlassButton
              type="submit"
              variant="primary"
              size="sm"
              disabled={isSubmitting || !newTitle.trim()}
              icon={isSubmitting ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Plus className="w-3.5 h-3.5" />}
            >
              {isSubmitting ? 'Сохранение...' : 'Выдать задание'}
            </GlassButton>
          </div>
        </form>
      )}

      {/* Список домашних заданий */}
      {isLoading ? (
        <div className="py-16 flex flex-col items-center justify-center space-y-3">
          <Loader2 className="w-8 h-8 text-indigo-600 animate-spin" />
          <p className="text-sm font-medium text-slate-500">Загрузка домашних заданий...</p>
        </div>
      ) : errorMessage ? (
        <div className="p-4 rounded-2xl bg-rose-500/15 border border-rose-400/30 backdrop-blur-md flex items-center gap-2 text-xs text-rose-800">
          <AlertCircle className="w-4 h-4 text-rose-600 shrink-0" />
          <span>{errorMessage}</span>
        </div>
      ) : homeworks.length === 0 ? (
        <div className="py-12 text-center rounded-2xl bg-white/20 backdrop-blur-md border border-white/40 p-6">
          <CheckSquare className="w-10 h-10 text-slate-300 mx-auto mb-2.5" />
          <h4 className="text-sm font-bold text-slate-800 mb-1">
            Заданий не найдено
          </h4>
          <p className="text-xs text-slate-500 max-w-xs mx-auto">
            {filter === 'all'
              ? 'Ученику пока не выдано ни одного домашнего задания.'
              : 'Нет заданий в выбранном статусе.'}
          </p>
        </div>
      ) : (
        <div className="space-y-2.5">
          {homeworks.map((hw) => (
            <ClientHomeworkItem
              key={hw.id}
              homework={hw}
              isUpdating={updatingId === hw.id}
              onStatusChange={handleStatusChange}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}
    </div>
  );
};
