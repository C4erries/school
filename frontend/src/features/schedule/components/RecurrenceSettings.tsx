import React from 'react';
import { GlassInput } from '../../../shared/components/GlassInput';
import { RotateCw, Calendar } from 'lucide-react';
import { WEEKDAYS } from '../utils/recurrence';

interface RecurrenceSettingsProps {
  isRecurring: boolean;
  onToggleRecurring: (val: boolean) => void;
  selectedDays: string[];
  onToggleDay: (day: string) => void;
  untilMode: 'never' | 'date';
  onChangeUntilMode: (mode: 'never' | 'date') => void;
  untilDate: string;
  onChangeUntilDate: (date: string) => void;
}

export const RecurrenceSettings: React.FC<RecurrenceSettingsProps> = ({
  isRecurring,
  onToggleRecurring,
  selectedDays,
  onToggleDay,
  untilMode,
  onChangeUntilMode,
  untilDate,
  onChangeUntilDate,
}) => {
  return (
    <div className="rounded-2xl p-3 sm:p-4 bg-white/40 border border-white/60 space-y-3 transition-all">
      {/* Toggle row */}
      <div className="flex items-center justify-between">
        <label
          htmlFor="recurring-toggle"
          className="flex items-center gap-2 cursor-pointer select-none"
        >
          <div className="p-1 rounded-lg bg-indigo-50 text-indigo-600">
            <RotateCw className="w-4 h-4" />
          </div>
          <div>
            <span className="text-xs font-bold text-slate-900 block">
              Повторять еженедельно
            </span>
            <span className="text-[11px] text-slate-500">
              Создать регулярную серию занятий по расписанию
            </span>
          </div>
        </label>
        <input
          id="recurring-toggle"
          type="checkbox"
          checked={isRecurring}
          onChange={(e) => onToggleRecurring(e.target.checked)}
          className="w-4 h-4 rounded text-indigo-600 border-slate-300 focus:ring-indigo-500 cursor-pointer"
        />
      </div>

      {isRecurring && (
        <div className="space-y-3 pt-2 border-t border-black/[0.05] animate-in fade-in">
          {/* Day chips */}
          <div>
            <span className="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-0.5">
              Дни недели повторения
            </span>
            <div className="grid grid-cols-7 gap-1 sm:gap-1.5">
              {WEEKDAYS.map((w) => {
                const isSelected = selectedDays.includes(w.key);
                return (
                  <button
                    key={w.key}
                    type="button"
                    onClick={() => onToggleDay(w.key)}
                    className={`py-2 px-1 rounded-xl text-xs font-bold transition-all text-center ${
                      isSelected
                        ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/30 scale-102'
                        : 'bg-white/60 hover:bg-white text-slate-600 border border-black/[0.06]'
                    }`}
                  >
                    {w.label}
                  </button>
                );
              })}
            </div>
            {selectedDays.length === 0 && (
              <p className="text-[11px] text-rose-500 mt-1 ml-0.5">
                Выберите хотя бы один день недели
              </p>
            )}
          </div>

          {/* Horizon selector */}
          <div>
            <span className="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-0.5">
              Горизонт повторений
            </span>
            <div className="grid grid-cols-2 gap-2 p-1 rounded-xl bg-black/[0.04]">
              <button
                type="button"
                onClick={() => onChangeUntilMode('never')}
                className={`py-1.5 px-3 rounded-lg text-xs font-semibold transition-all ${
                  untilMode === 'never'
                    ? 'bg-white text-indigo-600 shadow-sm'
                    : 'text-slate-600 hover:text-slate-900'
                }`}
              >
                Бессрочно
              </button>
              <button
                type="button"
                onClick={() => onChangeUntilMode('date')}
                className={`py-1.5 px-3 rounded-lg text-xs font-semibold transition-all ${
                  untilMode === 'date'
                    ? 'bg-white text-indigo-600 shadow-sm'
                    : 'text-slate-600 hover:text-slate-900'
                }`}
              >
                До даты...
              </button>
            </div>

            {untilMode === 'date' && (
              <div className="mt-2.5 animate-in fade-in">
                <GlassInput
                  label="Дата окончания серии"
                  type="date"
                  value={untilDate}
                  onChange={(e) => onChangeUntilDate(e.target.value)}
                  icon={<Calendar className="w-4 h-4 text-slate-400" />}
                  required
                />
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
