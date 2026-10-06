import React from 'react';
import { Calendar } from 'lucide-react';

export type PeriodPreset = 'month' | 'quarter' | 'half_year' | 'year' | 'custom';

interface AnalyticsPeriodSelectorProps {
  from: string; // YYYY-MM-DD
  to: string;   // YYYY-MM-DD
  interval: 'week' | 'month';
  preset: PeriodPreset;
  onPresetChange: (preset: PeriodPreset) => void;
  onDateChange: (from: string, to: string) => void;
  onIntervalChange: (interval: 'week' | 'month') => void;
}

export const AnalyticsPeriodSelector: React.FC<AnalyticsPeriodSelectorProps> = ({
  from,
  to,
  interval,
  preset,
  onPresetChange,
  onDateChange,
  onIntervalChange,
}) => {
  const presets: { id: PeriodPreset; label: string }[] = [
    { id: 'month', label: 'Этот месяц' },
    { id: 'quarter', label: '3 месяца' },
    { id: 'half_year', label: 'Полгода' },
    { id: 'year', label: 'Год' },
  ];

  return (
    <div className="flex flex-col lg:flex-row items-stretch lg:items-center justify-between gap-4 p-3 sm:p-4 rounded-3xl liquid-glass">
      {/* Быстрые пресеты */}
      <div className="flex flex-wrap items-center gap-1.5 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05]">
        {presets.map((p) => {
          const isActive = preset === p.id;
          return (
            <button
              key={p.id}
              type="button"
              onClick={() => onPresetChange(p.id)}
              className={`px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
                isActive
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
              }`}
            >
              {p.label}
            </button>
          );
        })}
      </div>

      {/* Пользовательский диапазон дат и шаг группировки */}
      <div className="flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-2 p-1 rounded-2xl bg-black/[0.03] border border-black/[0.05]">
          <div className="flex items-center gap-1.5 px-2.5 py-1 text-slate-500 text-xs">
            <Calendar className="w-3.5 h-3.5 text-indigo-500" />
            <span className="hidden sm:inline font-medium">Диапазон:</span>
          </div>
          <input
            type="date"
            value={from}
            onChange={(e) => onDateChange(e.target.value, to)}
            className="text-xs px-2.5 py-1 rounded-xl bg-white/70 border border-black/[0.08] text-slate-700 focus:outline-none focus:ring-1 focus:ring-indigo-400 font-medium"
            title="Дата начала"
          />
          <span className="text-slate-400 text-xs">—</span>
          <input
            type="date"
            value={to}
            onChange={(e) => onDateChange(from, e.target.value)}
            className="text-xs px-2.5 py-1 rounded-xl bg-white/70 border border-black/[0.08] text-slate-700 focus:outline-none focus:ring-1 focus:ring-indigo-400 font-medium"
            title="Дата окончания"
          />
        </div>

        {/* Шаг динамики: неделя/месяц */}
        <div className="flex items-center gap-1 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05]">
          <button
            type="button"
            onClick={() => onIntervalChange('week')}
            className={`px-3 py-1 rounded-xl text-xs font-semibold transition-all ${
              interval === 'week'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            По неделям
          </button>
          <button
            type="button"
            onClick={() => onIntervalChange('month')}
            className={`px-3 py-1 rounded-xl text-xs font-semibold transition-all ${
              interval === 'month'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            По месяцам
          </button>
        </div>
      </div>
    </div>
  );
};
