import React from 'react';
import { GlassCard } from '../../../shared/components/GlassCard';
import { Badge } from '../../../shared/components/Badge';
import { AnalyticsClientStat } from '../../../types/analytics';
import { Users, ArrowUpDown, Clock, Wallet, AlertCircle } from 'lucide-react';

interface AnalyticsClientsTableProps {
  clients: AnalyticsClientStat[];
  sort: 'hours' | 'revenue' | 'cancellations';
  onSortChange: (sort: 'hours' | 'revenue' | 'cancellations') => void;
  isLoading?: boolean;
}

export const AnalyticsClientsTable: React.FC<AnalyticsClientsTableProps> = ({
  clients,
  sort,
  onSortChange,
  isLoading = false,
}) => {
  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      maximumFractionDigits: 0,
    }).format(val);
  };

  return (
    <GlassCard className="p-5 sm:p-6 space-y-4">
      {/* Шапка таблицы */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-xl bg-indigo-500/10 text-indigo-600 flex items-center justify-center">
              <Users className="w-4 h-4" />
            </div>
            <h3 className="text-base font-bold text-slate-900 tracking-tight">
              Показатели учеников
            </h3>
          </div>
          <p className="text-xs text-slate-500">
            Рейтинг учеников по объему занятий, выручке и дисциплине посещаемости
          </p>
        </div>

        {/* Быстрые переключатели сортировки */}
        <div className="flex items-center gap-1 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05] self-start sm:self-auto">
          <button
            type="button"
            onClick={() => onSortChange('hours')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              sort === 'hours'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            <Clock className="w-3.5 h-3.5" />
            <span>По часам</span>
          </button>
          <button
            type="button"
            onClick={() => onSortChange('revenue')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              sort === 'revenue'
                ? 'bg-white text-emerald-700 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            <Wallet className="w-3.5 h-3.5" />
            <span>По доходу</span>
          </button>
          <button
            type="button"
            onClick={() => onSortChange('cancellations')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              sort === 'cancellations'
                ? 'bg-white text-rose-700 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            <AlertCircle className="w-3.5 h-3.5" />
            <span>По отменам</span>
          </button>
        </div>
      </div>

      {/* Таблица */}
      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs border-collapse">
          <thead>
            <tr className="border-b border-black/[0.06] text-slate-500 font-semibold uppercase tracking-wider text-[10px]">
              <th className="py-3 px-3">#</th>
              <th className="py-3 px-3">Ученик</th>
              <th className="py-3 px-3 text-right">
                <button
                  type="button"
                  onClick={() => onSortChange('hours')}
                  className="inline-flex items-center gap-1 hover:text-slate-900"
                >
                  Часы
                  <ArrowUpDown className="w-3 h-3" />
                </button>
              </th>
              <th className="py-3 px-3 text-right">
                <button
                  type="button"
                  onClick={() => onSortChange('revenue')}
                  className="inline-flex items-center gap-1 hover:text-slate-900"
                >
                  Доход
                  <ArrowUpDown className="w-3 h-3" />
                </button>
              </th>
              <th className="py-3 px-3 text-center">Проведено</th>
              <th className="py-3 px-3 text-center">
                <button
                  type="button"
                  onClick={() => onSortChange('cancellations')}
                  className="inline-flex items-center gap-1 hover:text-slate-900"
                >
                  Отмены
                  <ArrowUpDown className="w-3 h-3" />
                </button>
              </th>
              <th className="py-3 px-3 text-right">Посещаемость</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-black/[0.04]">
            {isLoading ? (
              <tr>
                <td colSpan={7} className="py-8 text-center text-slate-400">
                  Загрузка показателей учеников...
                </td>
              </tr>
            ) : clients.length === 0 ? (
              <tr>
                <td colSpan={7} className="py-8 text-center text-slate-400">
                  Нет данных об активности учеников за выбранный период
                </td>
              </tr>
            ) : (
              clients.map((c, idx) => (
                <tr
                  key={c.client_id}
                  className="hover:bg-black/[0.02] transition-colors"
                >
                  <td className="py-3.5 px-3 font-medium text-slate-400">
                    {idx + 1}
                  </td>
                  <td className="py-3.5 px-3">
                    <span className="font-semibold text-slate-900 text-sm">
                      {c.client_name}
                    </span>
                  </td>
                  <td className="py-3.5 px-3 text-right font-semibold text-slate-800">
                    {c.completed_hours} ч
                  </td>
                  <td className="py-3.5 px-3 text-right font-bold text-slate-900">
                    {formatCurrency(c.net_income)}
                  </td>
                  <td className="py-3.5 px-3 text-center text-slate-600">
                    {c.completed_count}
                  </td>
                  <td className="py-3.5 px-3 text-center">
                    {c.cancelled_count > 0 ? (
                      <span className="font-semibold text-rose-600">
                        {c.cancelled_count}
                      </span>
                    ) : (
                      <span className="text-slate-400">0</span>
                    )}
                  </td>
                  <td className="py-3.5 px-3 text-right">
                    <Badge
                      variant={
                        c.attendance_rate >= 90
                          ? 'mint'
                          : c.attendance_rate >= 75
                          ? 'amber'
                          : 'coral'
                      }
                      className="text-[11px] py-0.5 px-2"
                    >
                      {c.attendance_rate.toFixed(0)}%
                    </Badge>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </GlassCard>
  );
};
