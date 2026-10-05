import React from 'react';
import { Client } from '../../../types/schedule';
import {
  Calendar as CalendarIcon,
  ChevronLeft,
  ChevronRight,
  Columns,
  CalendarDays,
  List,
} from 'lucide-react';

interface ScheduleToolbarProps {
  currentDate: Date;
  viewMode: 'day' | 'week' | 'list';
  weekDays: Date[];
  clients: Client[];
  onNavigate: (direction: -1 | 1) => void;
  onGoToday: () => void;
  onViewModeChange: (mode: 'day' | 'week' | 'list') => void;
  onQuickAssignClient: (client: Client) => void;
}

export const ScheduleToolbar: React.FC<ScheduleToolbarProps> = ({
  currentDate,
  viewMode,
  weekDays,
  clients,
  onNavigate,
  onGoToday,
  onViewModeChange,
  onQuickAssignClient,
}) => {
  return (
    <div className="space-y-4">
      {/* Быстрое назначение ученику */}
      {clients.length > 0 && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-500">
              Быстрое назначение ({clients.length})
            </span>
          </div>
          <div className="flex items-center gap-3 overflow-x-auto pb-2 scrollbar-none">
            {clients.map((client) => (
              <div
                key={client.id}
                onClick={() => onQuickAssignClient(client)}
                className="p-3 rounded-2xl bg-white/30 backdrop-blur-md hover:bg-white/60 border border-white/40 flex items-center gap-3 shrink-0 cursor-pointer transition-all shadow-sm hover:scale-[1.02]"
              >
                <div className="w-8 h-8 rounded-xl bg-indigo-500/15 border border-indigo-400/30 text-indigo-700 flex items-center justify-center font-bold text-xs shadow-sm">
                  {client.name.charAt(0)}
                </div>
                <div className="text-left pr-2">
                  <div className="text-xs font-bold text-slate-900 leading-tight">
                    {client.name}
                  </div>
                  <div className="text-[10px] text-indigo-600 font-semibold">
                    + Запланировать
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Панель навигации по календарю и переключатель режимов [ День | Неделя | Список ] */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4 p-3 rounded-3xl bg-white/30 backdrop-blur-md border border-white/40 shadow-sm">
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => onNavigate(-1)}
            className="p-2 rounded-xl hover:bg-black/5 text-slate-600 transition-colors"
            title="Назад"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>

          <div className="flex items-center gap-2 px-3.5 py-1.5 rounded-xl bg-white/60 border border-white/80 shadow-sm">
            <CalendarIcon className="w-4 h-4 text-indigo-600" />
            <span className="font-bold text-sm text-slate-900 capitalize">
              {viewMode === 'week' ? (
                <>
                  {weekDays[0].toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })} –{' '}
                  {weekDays[6].toLocaleDateString('ru-RU', {
                    day: 'numeric',
                    month: 'short',
                    year: 'numeric',
                  })}
                </>
              ) : (
                currentDate.toLocaleDateString('ru-RU', {
                  weekday: 'short',
                  day: 'numeric',
                  month: 'long',
                  year: 'numeric',
                })
              )}
            </span>
          </div>

          <button
            type="button"
            onClick={() => onNavigate(1)}
            className="p-2 rounded-xl hover:bg-black/5 text-slate-600 transition-colors"
            title="Вперёд"
          >
            <ChevronRight className="w-4 h-4" />
          </button>

          <button
            type="button"
            onClick={onGoToday}
            className="text-xs font-semibold px-3 py-1.5 rounded-xl text-indigo-600 hover:bg-indigo-50 transition-colors ml-1"
          >
            Сегодня
          </button>
        </div>

        {/* Переключатель вида [ День | Неделя | Список ] */}
        <div className="flex items-center gap-1.5 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05] self-end sm:self-auto">
          <button
            type="button"
            onClick={() => onViewModeChange('day')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              viewMode === 'day'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <Columns className="w-3.5 h-3.5" />
            <span>День</span>
          </button>

          <button
            type="button"
            onClick={() => onViewModeChange('week')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              viewMode === 'week'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <CalendarDays className="w-3.5 h-3.5" />
            <span>Неделя</span>
          </button>

          <button
            type="button"
            onClick={() => onViewModeChange('list')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              viewMode === 'list'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <List className="w-3.5 h-3.5" />
            <span>Список</span>
          </button>
        </div>
      </div>
    </div>
  );
};
