import React from 'react';
import { Calendar, PlusCircle } from 'lucide-react';
import { UpcomingLessonInfo } from '../../types/journal';

interface UpcomingLessonBannerProps {
  upcomingLesson: UpcomingLessonInfo | null | undefined;
  onPlanLesson?: (lessonId: string) => void;
}

export const UpcomingLessonBanner: React.FC<UpcomingLessonBannerProps> = ({
  upcomingLesson,
  onPlanLesson,
}) => {
  if (!upcomingLesson) return null;

  const startDate = new Date(upcomingLesson.start_time);
  const now = new Date();
  const isToday =
    startDate.getDate() === now.getDate() &&
    startDate.getMonth() === now.getMonth() &&
    startDate.getFullYear() === now.getFullYear();

  const tomorrow = new Date(now);
  tomorrow.setDate(tomorrow.getDate() + 1);
  const isTomorrow =
    startDate.getDate() === tomorrow.getDate() &&
    startDate.getMonth() === tomorrow.getMonth() &&
    startDate.getFullYear() === tomorrow.getFullYear();

  const timeStr = startDate.toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  });

  const dateLabel = isToday
    ? `Сегодня в ${timeStr}`
    : isTomorrow
    ? `Завтра в ${timeStr}`
    : `${startDate.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })} в ${timeStr}`;

  return (
    <div className="p-3.5 sm:p-4 rounded-2xl bg-indigo-500/10 border border-indigo-400/30 backdrop-blur-md flex items-center justify-between gap-3 shadow-sm transition-all animate-in fade-in">
      <div className="flex items-center gap-3 min-w-0">
        <div className="w-9 h-9 rounded-xl bg-indigo-600/15 border border-indigo-500/20 flex items-center justify-center text-indigo-700 shrink-0">
          <Calendar className="w-4 h-4" />
        </div>
        <div className="min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-xs font-bold text-indigo-950">
              Ближайший урок: {dateLabel}
            </span>
            <span className="text-[10px] font-semibold px-2 py-0.5 rounded-lg bg-indigo-100/60 text-indigo-700">
              {upcomingLesson.format === 'individual'
                ? 'Индивидуально'
                : upcomingLesson.format === 'pair'
                ? 'В паре'
                : 'Группа'}
            </span>
          </div>
          <p className="text-xs text-indigo-800/80 truncate mt-0.5">
            {upcomingLesson.topic ? (
              <span>Тема: {upcomingLesson.topic}</span>
            ) : (
              <span>{upcomingLesson.title || 'План занятия не задан'}</span>
            )}
          </p>
        </div>
      </div>

      {onPlanLesson && (
        <button
          type="button"
          onClick={() => onPlanLesson(upcomingLesson.lesson_id)}
          className="shrink-0 inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold text-indigo-700 bg-white/60 hover:bg-white/80 border border-indigo-300/40 transition-colors shadow-sm"
        >
          <PlusCircle className="w-3.5 h-3.5" />
          <span className="hidden sm:inline">Задать план</span>
        </button>
      )}
    </div>
  );
};
