import React from 'react';
import { LessonJournalBundle } from '../../types/journal';
import { BookOpen, Star, Clock } from 'lucide-react';

interface LessonReportFeedItemProps {
  bundle: LessonJournalBundle;
}

export const LessonReportFeedItem: React.FC<LessonReportFeedItemProps> = ({ bundle }) => {
  const journal = bundle.journal;
  if (!journal) return null;

  const dateStr = new Date(journal.created_at).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  });

  return (
    <div className="p-4 sm:p-5 rounded-2xl bg-white/40 border border-white/60 backdrop-blur-md shadow-sm space-y-3 transition-all hover:bg-white/50">
      {/* Шапка отчета по уроку */}
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <div className="w-7 h-7 rounded-xl bg-indigo-600/10 text-indigo-700 flex items-center justify-center shrink-0">
            <BookOpen className="w-4 h-4" />
          </div>
          <div>
            <h4 className="text-sm font-bold text-slate-900 tracking-tight">
              {journal.topic}
            </h4>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {journal.performance_score && (
            <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-xl text-xs font-bold bg-amber-500/15 text-amber-900 border border-amber-400/30">
              <Star className="w-3.5 h-3.5 fill-amber-500 text-amber-500" />
              {journal.performance_score}/5
            </span>
          )}
          <div className="flex items-center gap-1 text-[11px] text-slate-500">
            <Clock className="w-3 h-3 text-slate-400" />
            <span>{dateStr}</span>
          </div>
        </div>
      </div>

      {/* Заметки преподавателя по уроку */}
      {journal.notes && (
        <div className="p-3 rounded-xl bg-white/30 border border-white/50 text-xs sm:text-sm text-slate-700 whitespace-pre-wrap leading-relaxed">
          {journal.notes}
        </div>
      )}

      {/* Выданные домашние задания */}
      {bundle.assigned_homeworks && bundle.assigned_homeworks.length > 0 && (
        <div className="space-y-1.5 pt-1">
          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-500">
            Выданное домашнее задание
          </span>
          <div className="space-y-1.5">
            {bundle.assigned_homeworks.map((hw) => {
              const isCompleted = hw.status === 'completed';
              return (
                <div
                  key={hw.id}
                  className="p-2.5 rounded-xl bg-white/30 border border-white/40 flex items-start justify-between gap-2 text-xs"
                >
                  <div className="space-y-0.5 min-w-0">
                    <span className="font-semibold text-slate-800 block truncate">
                      {hw.title}
                    </span>
                    {hw.description && (
                      <p className="text-slate-500 text-[11px] line-clamp-2">
                        {hw.description}
                      </p>
                    )}
                  </div>
                  <span
                    className={`shrink-0 inline-flex items-center gap-1 px-2 py-0.5 rounded-lg text-[10px] font-semibold ${
                      isCompleted
                        ? 'bg-emerald-500/15 text-emerald-800 border border-emerald-400/30'
                        : 'bg-amber-500/15 text-amber-800 border border-amber-400/30'
                    }`}
                  >
                    {isCompleted ? 'Сдано' : 'Задано'}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};
