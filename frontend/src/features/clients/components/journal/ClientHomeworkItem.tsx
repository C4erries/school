import React from 'react';
import { HomeworkAssignment, HomeworkStatus } from '../../../../types/journal';
import { Badge } from '../../../../shared/components/Badge';
import { Check, X, RotateCcw, Calendar, Trash2 } from 'lucide-react';

interface ClientHomeworkItemProps {
  homework: HomeworkAssignment;
  isUpdating: boolean;
  onStatusChange: (homeworkId: string, status: HomeworkStatus) => void;
  onDelete: (homeworkId: string) => void;
}

export const ClientHomeworkItem: React.FC<ClientHomeworkItemProps> = ({
  homework,
  isUpdating,
  onStatusChange,
  onDelete,
}) => {
  const renderStatusBadge = (status: HomeworkStatus) => {
    switch (status) {
      case 'completed':
        return <Badge variant="mint">Выполнено</Badge>;
      case 'not_done':
        return <Badge variant="danger">Не сдано</Badge>;
      case 'assigned':
      default:
        return <Badge variant="amber">В работе</Badge>;
    }
  };

  return (
    <div className="p-3.5 rounded-2xl bg-white/35 backdrop-blur-md border border-white/60 shadow-xs flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-white/50 transition-all">
      <div className="min-w-0 space-y-1">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="font-bold text-sm text-slate-900 leading-snug truncate">
            {homework.title}
          </span>
          {renderStatusBadge(homework.status)}
        </div>

        {homework.description && (
          <p className="text-xs text-slate-600 line-clamp-2 leading-relaxed">
            {homework.description}
          </p>
        )}

        <div className="flex items-center gap-3 text-[11px] text-slate-500 font-mono flex-wrap">
          {homework.due_date && (
            <span className="flex items-center gap-1">
              <Calendar className="w-3 h-3 text-slate-400" />
              Дедлайн: {new Date(homework.due_date).toLocaleDateString('ru-RU')}
            </span>
          )}
          <span>
            Выдано: {new Date(homework.created_at).toLocaleDateString('ru-RU')}
          </span>
        </div>
      </div>

      {/* Кнопки смены статуса и удаления */}
      <div className="flex items-center gap-1.5 shrink-0 self-end sm:self-auto">
        {homework.status !== 'completed' && (
          <button
            type="button"
            disabled={isUpdating}
            onClick={() => onStatusChange(homework.id, 'completed')}
            className="px-2.5 py-1.5 rounded-xl text-xs font-semibold bg-white/60 hover:bg-emerald-50 text-emerald-800 border border-emerald-300 hover:border-emerald-400 transition-all flex items-center gap-1 active:scale-95 shadow-2xs"
            title="Отметить сданным"
          >
            <Check className="w-3.5 h-3.5 stroke-[2.5]" />
            <span className="hidden xs:inline">Сдано</span>
          </button>
        )}

        {homework.status !== 'not_done' && (
          <button
            type="button"
            disabled={isUpdating}
            onClick={() => onStatusChange(homework.id, 'not_done')}
            className="px-2.5 py-1.5 rounded-xl text-xs font-semibold bg-white/60 hover:bg-rose-50 text-rose-800 border border-rose-300 hover:border-rose-400 transition-all flex items-center gap-1 active:scale-95 shadow-2xs"
            title="Отметить не сданным"
          >
            <X className="w-3.5 h-3.5 stroke-[2.5]" />
            <span className="hidden xs:inline">Не сдано</span>
          </button>
        )}

        {homework.status !== 'assigned' && (
          <button
            type="button"
            disabled={isUpdating}
            onClick={() => onStatusChange(homework.id, 'assigned')}
            className="px-2 py-1.5 rounded-xl text-xs font-semibold bg-white/60 hover:bg-amber-50 text-amber-800 border border-amber-300 hover:border-amber-400 transition-all flex items-center gap-1 active:scale-95 shadow-2xs"
            title="Вернуть в работу"
          >
            <RotateCcw className="w-3 h-3" />
          </button>
        )}

        <button
          type="button"
          disabled={isUpdating}
          onClick={() => onDelete(homework.id)}
          className="p-1.5 rounded-xl text-slate-400 hover:text-rose-600 hover:bg-rose-50/50 transition-colors"
          title="Удалить задание"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      </div>
    </div>
  );
};
