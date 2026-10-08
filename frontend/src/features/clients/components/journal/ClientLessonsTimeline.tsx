import React, { useState, useEffect, useCallback } from 'react';
import { LessonJournal } from '../../../../types/journal';
import { getClientJournals } from '../../../../api/journal';
import { BookOpen, Calendar, Star, Loader2, AlertCircle, RefreshCw } from 'lucide-react';

interface ClientLessonsTimelineProps {
  clientId: string;
}

const SCORE_BADGES: Record<
  number,
  { label: string; bg: string; text: string; border: string }
> = {
  1: {
    label: '1 • Сложно',
    bg: 'bg-rose-500/15',
    text: 'text-rose-700',
    border: 'border-rose-400/30',
  },
  2: {
    label: '2 • С трудом',
    bg: 'bg-orange-500/15',
    text: 'text-orange-700',
    border: 'border-orange-400/30',
  },
  3: {
    label: '3 • Удовл.',
    bg: 'bg-amber-500/15',
    text: 'text-amber-700',
    border: 'border-amber-400/30',
  },
  4: {
    label: '4 • Хорошо',
    bg: 'bg-indigo-500/15',
    text: 'text-indigo-700',
    border: 'border-indigo-400/30',
  },
  5: {
    label: '5 • Отлично',
    bg: 'bg-emerald-500/15',
    text: 'text-emerald-700',
    border: 'border-emerald-400/30',
  },
};

export const ClientLessonsTimeline: React.FC<ClientLessonsTimelineProps> = ({
  clientId,
}) => {
  const [journals, setJournals] = useState<LessonJournal[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    setIsLoading(true);
    setErrorMessage(null);
    try {
      const data = await getClientJournals(clientId);
      // Сортировка от свежих к старым
      const sorted = [...data].sort(
        (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      );
      setJournals(sorted);
    } catch (err: unknown) {
      setErrorMessage(
        err instanceof Error ? err.message : 'Ошибка загрузки истории занятий'
      );
    } finally {
      setIsLoading(false);
    }
  }, [clientId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  if (isLoading) {
    return (
      <div className="py-16 flex flex-col items-center justify-center space-y-3">
        <Loader2 className="w-8 h-8 text-indigo-600 animate-spin" />
        <p className="text-sm font-medium text-slate-500">Загрузка истории занятий...</p>
      </div>
    );
  }

  if (errorMessage) {
    return (
      <div className="p-4 rounded-2xl bg-rose-500/15 border border-rose-400/30 backdrop-blur-md flex items-center justify-between gap-3 text-xs text-rose-800">
        <div className="flex items-center gap-2">
          <AlertCircle className="w-4 h-4 text-rose-600 shrink-0" />
          <span>{errorMessage}</span>
        </div>
        <button
          type="button"
          onClick={loadData}
          className="p-1.5 rounded-xl hover:bg-rose-500/20 text-rose-700 transition-colors"
        >
          <RefreshCw className="w-3.5 h-3.5" />
        </button>
      </div>
    );
  }

  if (journals.length === 0) {
    return (
      <div className="py-12 text-center rounded-2xl bg-white/20 backdrop-blur-md border border-white/40 p-6">
        <BookOpen className="w-10 h-10 text-slate-300 mx-auto mb-2.5" />
        <h4 className="text-sm font-bold text-slate-800 mb-1">
          Записей в дневнике пока нет
        </h4>
        <p className="text-xs text-slate-500 max-w-xs mx-auto">
          Отчеты появятся здесь после сохранения тем и оценок в расписании уроков.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-3 relative before:absolute before:inset-0 before:left-3.5 before:w-0.5 before:bg-indigo-200/60 pl-8">
      {journals.map((journal) => {
        const scoreBadge =
          journal.performance_score && SCORE_BADGES[journal.performance_score];
        const dateStr = new Date(journal.created_at).toLocaleDateString('ru-RU', {
          day: 'numeric',
          month: 'short',
          year: 'numeric',
        });

        return (
          <div key={journal.id} className="relative group">
            {/* Точка таймлайна */}
            <div className="absolute -left-8 top-3.5 w-7 h-7 rounded-full bg-white/80 backdrop-blur-md border-2 border-indigo-500 shadow-xs flex items-center justify-center text-indigo-600">
              <BookOpen className="w-3.5 h-3.5" />
            </div>

            {/* Карточка отчета */}
            <div className="p-3.5 sm:p-4 rounded-2xl bg-white/35 backdrop-blur-md border border-white/60 shadow-xs space-y-2 hover:bg-white/50 transition-all">
              <div className="flex items-start justify-between gap-2 flex-wrap">
                <div className="min-w-0">
                  <h4 className="font-bold text-sm text-slate-900 leading-snug">
                    {journal.topic}
                  </h4>
                  <div className="flex items-center gap-1.5 text-[11px] text-slate-500 font-mono mt-0.5">
                    <Calendar className="w-3 h-3 text-slate-400" />
                    <span>{dateStr}</span>
                  </div>
                </div>

                {scoreBadge && (
                  <span
                    className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-xl text-xs font-bold border shadow-xs ${scoreBadge.bg} ${scoreBadge.text} ${scoreBadge.border}`}
                  >
                    <Star className="w-3 h-3 fill-current stroke-none" />
                    <span>{scoreBadge.label}</span>
                  </span>
                )}
              </div>

              {journal.notes && (
                <div className="text-xs text-slate-700 bg-white/40 backdrop-blur-md p-2.5 rounded-xl border border-white/50 leading-relaxed whitespace-pre-wrap">
                  {journal.notes}
                </div>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
};
