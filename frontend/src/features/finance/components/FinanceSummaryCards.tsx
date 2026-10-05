import React from 'react';
import { GlassCard } from '../../../shared/components/GlassCard';
import { FinanceSummary } from '../../../types/finance';
import {
  Wallet,
  AlertCircle,
  Building2,
  Users,
  ChevronLeft,
  ChevronRight,
  Calendar,
} from 'lucide-react';

interface FinanceSummaryCardsProps {
  summary: FinanceSummary | null;
  currentMonth: string; // "YYYY-MM"
  onMonthChange: (month: string) => void;
  isLoading?: boolean;
}

export const FinanceSummaryCards: React.FC<FinanceSummaryCardsProps> = ({
  summary,
  currentMonth,
  onMonthChange,
  isLoading = false,
}) => {
  const formatMonthTitle = (monthStr: string) => {
    try {
      const [year, month] = monthStr.split('-').map(Number);
      const date = new Date(year, month - 1, 1);
      const formatted = date.toLocaleDateString('ru-RU', {
        month: 'long',
        year: 'numeric',
      });
      return formatted.charAt(0).toUpperCase() + formatted.slice(1);
    } catch {
      return monthStr;
    }
  };

  const handlePrevMonth = () => {
    const [year, month] = currentMonth.split('-').map(Number);
    const date = new Date(year, month - 2, 1);
    const nextY = date.getFullYear();
    const nextM = String(date.getMonth() + 1).padStart(2, '0');
    onMonthChange(`${nextY}-${nextM}`);
  };

  const handleNextMonth = () => {
    const [year, month] = currentMonth.split('-').map(Number);
    const date = new Date(year, month, 1);
    const nextY = date.getFullYear();
    const nextM = String(date.getMonth() + 1).padStart(2, '0');
    onMonthChange(`${nextY}-${nextM}`);
  };

  const formatCurrency = (val?: number) => {
    const num = val ?? 0;
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      maximumFractionDigits: 0,
    }).format(num);
  };

  return (
    <div className="space-y-4">
      {/* Селектор месяца */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <div className="flex items-center p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05]">
            <button
              type="button"
              onClick={handlePrevMonth}
              className="p-1.5 rounded-xl hover:bg-white/60 text-slate-600 transition-colors"
              title="Предыдущий месяц"
              aria-label="Предыдущий месяц"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <div className="flex items-center gap-2 px-3 py-1 font-semibold text-sm text-slate-800">
              <Calendar className="w-4 h-4 text-indigo-500" />
              <span>{formatMonthTitle(currentMonth)}</span>
            </div>
            <button
              type="button"
              onClick={handleNextMonth}
              className="p-1.5 rounded-xl hover:bg-white/60 text-slate-600 transition-colors"
              title="Следующий месяц"
              aria-label="Следующий месяц"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
          <input
            type="month"
            value={currentMonth}
            onChange={(e) => e.target.value && onMonthChange(e.target.value)}
            className="text-xs px-2.5 py-1.5 rounded-xl bg-white/50 border border-black/[0.08] text-slate-600 focus:outline-none focus:ring-1 focus:ring-indigo-400"
          />
        </div>

        {summary && (
          <span className="text-xs text-slate-500">
            Отработано занятий на сумму:{' '}
            <strong className="text-slate-800 font-semibold">
              {formatCurrency(summary.total_earned)}
            </strong>
          </span>
        )}
      </div>

      {/* 4 метрики */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Метрика 1: Собрано оплат */}
        <GlassCard className="p-5 flex flex-col justify-between relative overflow-hidden group hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
              Поступило оплат
            </span>
            <div className="w-9 h-9 rounded-2xl bg-emerald-500/15 text-emerald-600 flex items-center justify-center border border-emerald-400/20">
              <Wallet className="w-4 h-4" />
            </div>
          </div>
          <div className="space-y-1">
            <div className="text-2xl font-bold tracking-tight text-slate-900">
              {isLoading ? (
                <div className="h-8 w-24 bg-slate-200/60 rounded-lg animate-pulse" />
              ) : (
                formatCurrency(summary?.total_payments)
              )}
            </div>
            <p className="text-xs text-emerald-700 font-medium">
              Кассовый приход за расчетный период
            </p>
          </div>
        </GlassCard>

        {/* Метрика 2: Сумма долгов учеников */}
        <GlassCard className="p-5 flex flex-col justify-between relative overflow-hidden group hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
              Долги учеников
            </span>
            <div className="w-9 h-9 rounded-2xl bg-rose-500/15 text-rose-600 flex items-center justify-center border border-rose-400/20">
              <AlertCircle className="w-4 h-4" />
            </div>
          </div>
          <div className="space-y-1">
            <div className="text-2xl font-bold tracking-tight text-slate-900">
              {isLoading ? (
                <div className="h-8 w-24 bg-slate-200/60 rounded-lg animate-pulse" />
              ) : (
                formatCurrency(summary?.total_debts)
              )}
            </div>
            <p className="text-xs text-rose-700 font-medium">
              {summary?.debtors_count ? (
                <span>{summary.debtors_count} {summary.debtors_count === 1 ? 'клиент с отрицательным балансом' : 'клиентов с отрицательным балансом'}</span>
              ) : (
                'Нет активных задолженностей'
              )}
            </p>
          </div>
        </GlassCard>

        {/* Метрика 3: Доля школ к выплате */}
        <GlassCard className="p-5 flex flex-col justify-between relative overflow-hidden group hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
              Комиссии школ
            </span>
            <div className="w-9 h-9 rounded-2xl bg-amber-500/15 text-amber-600 flex items-center justify-center border border-amber-400/20">
              <Building2 className="w-4 h-4" />
            </div>
          </div>
          <div className="space-y-1">
            <div className="text-2xl font-bold tracking-tight text-slate-900">
              {isLoading ? (
                <div className="h-8 w-24 bg-slate-200/60 rounded-lg animate-pulse" />
              ) : (
                formatCurrency(summary?.total_commissions)
              )}
            </div>
            <p className="text-xs text-amber-700 font-medium">
              Доля партнерских школ по тегам
            </p>
          </div>
        </GlassCard>

        {/* Метрика 4: Активные абонементы */}
        <GlassCard className="p-5 flex flex-col justify-between relative overflow-hidden group hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
              Абонементы
            </span>
            <div className="w-9 h-9 rounded-2xl bg-indigo-500/15 text-indigo-600 flex items-center justify-center border border-indigo-400/20">
              <Users className="w-4 h-4" />
            </div>
          </div>
          <div className="space-y-1">
            <div className="text-2xl font-bold tracking-tight text-slate-900">
              {isLoading ? (
                <div className="h-8 w-16 bg-slate-200/60 rounded-lg animate-pulse" />
              ) : (
                summary?.active_subscriptions_count ?? 0
              )}
            </div>
            <p className="text-xs text-indigo-700 font-medium">
              Учеников с положительным балансом
            </p>
          </div>
        </GlassCard>
      </div>
    </div>
  );
};
