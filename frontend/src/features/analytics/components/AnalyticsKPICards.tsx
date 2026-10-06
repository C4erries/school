import React from 'react';
import { GlassCard } from '../../../shared/components/GlassCard';
import { Badge } from '../../../shared/components/Badge';
import { AnalyticsOverview } from '../../../types/analytics';
import { Clock, CheckCircle2, Wallet, TrendingUp } from 'lucide-react';

interface AnalyticsKPICardsProps {
  overview: AnalyticsOverview | null;
  isLoading?: boolean;
}

export const AnalyticsKPICards: React.FC<AnalyticsKPICardsProps> = ({
  overview,
  isLoading = false,
}) => {
  const formatCurrency = (val?: number) => {
    const num = Math.round(val ?? 0);
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      maximumFractionDigits: 0,
    }).format(num);
  };

  const formatHours = (val?: number) => {
    const num = val ?? 0;
    return Number.isInteger(num) ? String(num) : num.toFixed(1);
  };

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      {/* Метрика 1: Отработано часов */}
      <GlassCard className="p-5 flex flex-col justify-between relative overflow-hidden group hover:shadow-md transition-shadow">
        <div className="flex items-center justify-between mb-3">
          <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
            Отработано часов
          </span>
          <div className="w-9 h-9 rounded-2xl bg-indigo-500/15 text-indigo-600 flex items-center justify-center border border-indigo-400/20">
            <Clock className="w-4 h-4" />
          </div>
        </div>
        <div className="space-y-1.5">
          <div className="text-3xl font-bold tracking-tight text-slate-900 flex items-baseline gap-1.5">
            {isLoading ? (
              <div className="h-8 w-24 bg-slate-200/60 rounded-lg animate-pulse" />
            ) : (
              <>
                <span>{formatHours(overview?.completed_hours)}</span>
                <span className="text-base font-medium text-slate-500">ч</span>
              </>
            )}
          </div>
          <div className="flex items-center gap-2 pt-0.5">
            <Badge variant="indigo" className="text-[11px] py-0.5 px-2">
              {overview?.completed_lessons ?? 0} проведено
            </Badge>
            <span className="text-[11px] text-slate-500">
              из {overview?.total_lessons ?? 0} занятий
            </span>
          </div>
        </div>
      </GlassCard>

      {/* Метрика 2: Доходимость */}
      <GlassCard className="p-5 flex flex-col justify-between relative overflow-hidden group hover:shadow-md transition-shadow">
        <div className="flex items-center justify-between mb-3">
          <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
            Доходимость уроков
          </span>
          <div className="w-9 h-9 rounded-2xl bg-emerald-500/15 text-emerald-600 flex items-center justify-center border border-emerald-400/20">
            <CheckCircle2 className="w-4 h-4" />
          </div>
        </div>
        <div className="space-y-1.5">
          <div className="text-3xl font-bold tracking-tight text-slate-900 flex items-baseline gap-1">
            {isLoading ? (
              <div className="h-8 w-20 bg-slate-200/60 rounded-lg animate-pulse" />
            ) : (
              <>
                <span>{(overview?.completion_rate ?? 0).toFixed(1)}</span>
                <span className="text-base font-medium text-slate-500">%</span>
              </>
            )}
          </div>
          <div className="flex items-center gap-2 pt-0.5">
            <Badge
              variant={overview?.cancelled_lessons ? 'amber' : 'mint'}
              className="text-[11px] py-0.5 px-2"
            >
              {overview?.cancelled_lessons ?? 0} отмен
            </Badge>
            <span className="text-[11px] text-slate-500">
              {overview?.completion_rate && overview.completion_rate >= 90
                ? 'Отличный показатель'
                : 'Требует внимания'}
            </span>
          </div>
        </div>
      </GlassCard>

      {/* Метрика 3: Чистый доход */}
      <GlassCard className="p-5 flex flex-col justify-between relative overflow-hidden group hover:shadow-md transition-shadow">
        <div className="flex items-center justify-between mb-3">
          <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
            Чистый доход
          </span>
          <div className="w-9 h-9 rounded-2xl bg-emerald-500/15 text-emerald-600 flex items-center justify-center border border-emerald-400/20">
            <Wallet className="w-4 h-4" />
          </div>
        </div>
        <div className="space-y-1.5">
          <div className="text-3xl font-bold tracking-tight text-slate-900">
            {isLoading ? (
              <div className="h-8 w-28 bg-slate-200/60 rounded-lg animate-pulse" />
            ) : (
              formatCurrency(overview?.net_income)
            )}
          </div>
          <div className="flex items-center gap-2 pt-0.5">
            <Badge variant="mint" className="text-[11px] py-0.5 px-2">
              Валовой: {formatCurrency(overview?.gross_revenue)}
            </Badge>
          </div>
        </div>
      </GlassCard>

      {/* Метрика 4: Средняя ставка в час */}
      <GlassCard className="p-5 flex flex-col justify-between relative overflow-hidden group hover:shadow-md transition-shadow">
        <div className="flex items-center justify-between mb-3">
          <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
            Эффективная ставка
          </span>
          <div className="w-9 h-9 rounded-2xl bg-amber-500/15 text-amber-600 flex items-center justify-center border border-amber-400/20">
            <TrendingUp className="w-4 h-4" />
          </div>
        </div>
        <div className="space-y-1.5">
          <div className="text-3xl font-bold tracking-tight text-slate-900 flex items-baseline gap-1">
            {isLoading ? (
              <div className="h-8 w-24 bg-slate-200/60 rounded-lg animate-pulse" />
            ) : (
              <>
                <span>{formatCurrency(overview?.effective_hourly_rate)}</span>
                <span className="text-base font-medium text-slate-500">/ ч</span>
              </>
            )}
          </div>
          <div className="flex items-center gap-2 pt-0.5">
            <span className="text-[11px] text-slate-500">
              Реальный заработок за час работы
            </span>
          </div>
        </div>
      </GlassCard>
    </div>
  );
};
