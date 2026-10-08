import React from 'react';
import { Client } from '../../types/schedule';
import { Clock, ChevronRight } from 'lucide-react';

interface JournalClientItemProps {
  client: Client;
  isSelected: boolean;
  onSelect: (client: Client) => void;
}

export const JournalClientItem: React.FC<JournalClientItemProps> = ({
  client,
  isSelected,
  onSelect,
}) => {
  const indBal = client.balances?.individual_hours ?? client.balance ?? 0;

  // Форматирование даты последнего занятия
  const formatLastLesson = (isoStr?: string | null) => {
    if (!isoStr) return 'Уроков еще не было';
    const d = new Date(isoStr);
    const now = new Date();
    const diffDays = Math.floor((now.getTime() - d.getTime()) / (1000 * 60 * 60 * 24));

    if (diffDays === 0) {
      return `Сегодня в ${d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}`;
    }
    if (diffDays === 1) {
      return `Вчера в ${d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}`;
    }
    return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
  };

  // Получить инициалы
  const getInitials = (name: string) => {
    const parts = name.trim().split(/\s+/);
    if (parts.length >= 2) {
      return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
    }
    return name.slice(0, 2).toUpperCase();
  };

  return (
    <button
      type="button"
      onClick={() => onSelect(client)}
      className={`w-full text-left p-3.5 rounded-2xl transition-all duration-200 flex items-center justify-between gap-3 group border ${
        isSelected
          ? 'bg-white/60 border-white/80 shadow-md backdrop-blur-md'
          : 'bg-white/20 hover:bg-white/40 border-white/30 backdrop-blur-sm'
      }`}
    >
      <div className="flex items-center gap-3 min-w-0">
        <div
          className={`w-10 h-10 rounded-2xl flex items-center justify-center font-bold text-xs shrink-0 transition-colors shadow-sm ${
            isSelected
              ? 'bg-indigo-600 text-white'
              : 'bg-indigo-100/60 text-indigo-700 group-hover:bg-indigo-200/70'
          }`}
        >
          {getInitials(client.name)}
        </div>

        <div className="min-w-0">
          <div className="flex items-center gap-1.5">
            <span
              className={`text-sm font-bold truncate tracking-tight ${
                isSelected ? 'text-slate-900' : 'text-slate-800'
              }`}
            >
              {client.name}
            </span>
          </div>

          <div className="flex items-center gap-1 text-[11px] text-slate-500 mt-0.5">
            <Clock className="w-3 h-3 text-slate-400 shrink-0" />
            <span className="truncate">{formatLastLesson(client.last_lesson_at)}</span>
          </div>
        </div>
      </div>

      <div className="flex flex-col items-end gap-1 shrink-0">
        {indBal < 0 ? (
          <span className="px-2 py-0.5 rounded-lg text-[10px] font-bold bg-rose-500/15 text-rose-800 border border-rose-400/30">
            {indBal} ч
          </span>
        ) : indBal > 0 ? (
          <span className="px-2 py-0.5 rounded-lg text-[10px] font-bold bg-emerald-500/15 text-emerald-800 border border-emerald-400/30">
            {indBal} ч
          </span>
        ) : (
          <span className="px-1.5 py-0.5 rounded-lg text-[10px] font-medium bg-black/5 text-slate-500">
            0 ч
          </span>
        )}
        <ChevronRight
          className={`w-3.5 h-3.5 text-slate-400 transition-transform ${
            isSelected ? 'text-indigo-600 translate-x-0.5' : 'group-hover:translate-x-0.5'
          }`}
        />
      </div>
    </button>
  );
};
