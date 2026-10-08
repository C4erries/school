import React from 'react';
import { HomeworkAssignment, HomeworkStatus } from '../../../../types/journal';
import { Badge } from '../../../../shared/components/Badge';
import { Check, X, CheckSquare, Calendar, Loader2 } from 'lucide-react';

interface LessonDueHomeworksProps {
  homeworks: HomeworkAssignment[];
  onUpdateStatus: (homeworkId: string, status: HomeworkStatus) => Promise<void>;
  isUpdatingId?: string | null;
}

export const LessonDueHomeworks: React.FC<LessonDueHomeworksProps> = ({
  homeworks,
  onUpdateStatus,
  isUpdatingId,
}) => {
  if (homeworks.length === 0) {
    return null;
  }

  const renderStatusBadge = (status: HomeworkStatus) => {
    switch (status) {
      case 'completed':
        return <Badge variant="mint">Сдано</Badge>;
      case 'not_done':
        return <Badge variant="danger">Не выполнено</Badge>;
      case 'assigned':
      default:
        return <Badge variant="amber">В работе</Badge>;
    }
  };

  return (
    <div className="space-y-2.5">
      <div className="flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider text-slate-600">
        <CheckSquare className="w-3.5 h-3.5 text-indigo-600" />
        <span>Экспресс-проверка ДЗ к этому уроку</span>
      </div>

      <div className="space-y-2">
        {homeworks.map((hw) => {
          const isLoading = isUpdatingId === hw.id;
          return (
            <div
              key={hw.id}
              className="p-3 rounded-2xl bg-white/40 backdrop-blur-md border border-white/60 shadow-xs flex flex-col sm:flex-row sm:items-center justify-between gap-3"
            >
              <div className="min-w-0 space-y-1">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="font-bold text-sm text-slate-900 truncate">
                    {hw.title}
                  </span>
                  {renderStatusBadge(hw.status)}
                </div>

                {hw.description && (
                  <p className="text-xs text-slate-600 line-clamp-2">
                    {hw.description}
                  </p>
                )}

                {hw.due_date && (
                  <div className="flex items-center gap-1 text-[11px] text-slate-500 font-mono">
                    <Calendar className="w-3 h-3 text-slate-400" />
                    <span>
                      Дедлайн: {new Date(hw.due_date).toLocaleDateString('ru-RU')}
                    </span>
                  </div>
                )}
              </div>

              {/* Быстрые кнопки смены статуса */}
              <div className="flex items-center gap-1.5 shrink-0 self-end sm:self-auto">
                <button
                  type="button"
                  disabled={isLoading || hw.status === 'completed'}
                  onClick={() => onUpdateStatus(hw.id, 'completed')}
                  className={`px-3 py-1.5 rounded-xl text-xs font-semibold flex items-center gap-1 transition-all border shadow-xs ${
                    hw.status === 'completed'
                      ? 'bg-emerald-500 text-white border-emerald-400 opacity-90'
                      : 'bg-white/60 hover:bg-emerald-50 text-emerald-800 border-emerald-300 hover:border-emerald-400 active:scale-95'
                  }`}
                  title="Отметить выполненным"
                >
                  {isLoading ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : (
                    <Check className="w-3.5 h-3.5 stroke-[2.5]" />
                  )}
                  <span>Сдано</span>
                </button>

                <button
                  type="button"
                  disabled={isLoading || hw.status === 'not_done'}
                  onClick={() => onUpdateStatus(hw.id, 'not_done')}
                  className={`px-3 py-1.5 rounded-xl text-xs font-semibold flex items-center gap-1 transition-all border shadow-xs ${
                    hw.status === 'not_done'
                      ? 'bg-rose-500 text-white border-rose-400 opacity-90'
                      : 'bg-white/60 hover:bg-rose-50 text-rose-800 border-rose-300 hover:border-rose-400 active:scale-95'
                  }`}
                  title="Отметить не сданным"
                >
                  <X className="w-3.5 h-3.5 stroke-[2.5]" />
                  <span>Не сдано</span>
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
