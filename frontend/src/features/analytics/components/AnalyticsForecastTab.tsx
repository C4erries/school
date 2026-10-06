import React, { useState, useEffect, useCallback } from 'react';
import { GlassCard } from '../../../shared/components/GlassCard';
import { Badge } from '../../../shared/components/Badge';
import { AnalyticsForecast } from '../../../types/analytics';
import { getForecast } from '../../../api/analytics';
import {
  Sparkles,
  Calendar,
  Clock,
  Wallet,
  TrendingUp,
  Percent,
  RefreshCw,
  AlertCircle,
} from 'lucide-react';

type ForecastPeriod = 'this_week' | 'next_week' | 'this_month' | 'next_month';

export const AnalyticsForecastTab: React.FC = () => {
  const [period, setPeriod] = useState<ForecastPeriod>('this_week');
  const [forecast, setForecast] = useState<AnalyticsForecast | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const computeDates = useCallback((p: ForecastPeriod): { from: string; to: string } => {
    const now = new Date();
    const currentDay = now.getDay();
    const diffToMonday = (currentDay === 0 ? -6 : 1) - currentDay;

    if (p === 'this_week') {
      const mon = new Date(now);
      mon.setDate(now.getDate() + diffToMonday);
      mon.setHours(0, 0, 0, 0);

      const sun = new Date(mon);
      sun.setDate(mon.getDate() + 6);
      sun.setHours(23, 59, 59, 999);

      return {
        from: mon.toISOString(),
        to: sun.toISOString(),
      };
    }

    if (p === 'next_week') {
      const nextMon = new Date(now);
      nextMon.setDate(now.getDate() + diffToMonday + 7);
      nextMon.setHours(0, 0, 0, 0);

      const nextSun = new Date(nextMon);
      nextSun.setDate(nextMon.getDate() + 6);
      nextSun.setHours(23, 59, 59, 999);

      return {
        from: nextMon.toISOString(),
        to: nextSun.toISOString(),
      };
    }

    if (p === 'this_month') {
      const start = new Date(now.getFullYear(), now.getMonth(), 1, 0, 0, 0, 0);
      const end = new Date(now.getFullYear(), now.getMonth() + 1, 0, 23, 59, 59, 999);
      return {
        from: start.toISOString(),
        to: end.toISOString(),
      };
    }

    // next_month
    const start = new Date(now.getFullYear(), now.getMonth() + 1, 1, 0, 0, 0, 0);
    const end = new Date(now.getFullYear(), now.getMonth() + 2, 0, 23, 59, 59, 999);
    return {
      from: start.toISOString(),
      to: end.toISOString(),
    };
  }, []);

  const loadForecast = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const dates = computeDates(period);
      const data = await getForecast({ from: dates.from, to: dates.to });
      setForecast(data);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Ошибка загрузки прогноза');
    } finally {
      setIsLoading(false);
    }
  }, [period, computeDates]);

  useEffect(() => {
    loadForecast();
  }, [loadForecast]);

  const formatCurrency = (val?: number) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      maximumFractionDigits: 0,
    }).format(Math.round(val ?? 0));
  };

  const formatHours = (val?: number) => {
    const num = val ?? 0;
    return Number.isInteger(num) ? String(num) : num.toFixed(1);
  };

  const formatLabels: Record<string, string> = {
    individual: 'Индивидуальные',
    pair: 'Пары',
    group: 'Группы',
  };

  const formatColors: Record<string, string> = {
    individual: 'bg-indigo-500',
    pair: 'bg-purple-500',
    group: 'bg-amber-500',
  };

  return (
    <div className="space-y-6">
      {/* Селектор периодов прогнозирования */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4 p-3 sm:p-4 rounded-3xl liquid-glass">
        <div className="flex flex-wrap items-center gap-1.5 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05]">
          <button
            type="button"
            onClick={() => setPeriod('this_week')}
            className={`px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              period === 'this_week'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            Текущая неделя
          </button>
          <button
            type="button"
            onClick={() => setPeriod('next_week')}
            className={`px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              period === 'next_week'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            Следующая неделя
          </button>
          <button
            type="button"
            onClick={() => setPeriod('this_month')}
            className={`px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              period === 'this_month'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            Текущий месяц
          </button>
          <button
            type="button"
            onClick={() => setPeriod('next_month')}
            className={`px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              period === 'next_month'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            Следующий месяц
          </button>
        </div>

        <button
          type="button"
          onClick={loadForecast}
          disabled={isLoading}
          className="self-end sm:self-auto px-3 py-1.5 rounded-xl text-xs font-medium text-slate-600 hover:text-slate-900 hover:bg-black/5 flex items-center gap-1.5"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
          <span>Пересчитать прогноз</span>
        </button>
      </div>

      {error && (
        <div className="p-4 rounded-3xl bg-rose-500/10 border border-rose-500/20 backdrop-blur-md flex items-center gap-3">
          <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
          <p className="text-sm font-medium text-rose-900">{error}</p>
        </div>
      )}

      {/* Карточки прогноза */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Часы */}
        <GlassCard className="p-5 flex flex-col justify-between">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
              Ожидаемая нагрузка
            </span>
            <div className="w-9 h-9 rounded-2xl bg-indigo-500/15 text-indigo-600 flex items-center justify-center border border-indigo-400/20">
              <Clock className="w-4 h-4" />
            </div>
          </div>
          <div className="space-y-1">
            <div className="text-3xl font-bold text-slate-900 flex items-baseline gap-1.5">
              {isLoading ? (
                <div className="h-8 w-20 bg-slate-200/60 rounded-lg animate-pulse" />
              ) : (
                <>
                  <span>{formatHours(forecast?.scheduled_hours)}</span>
                  <span className="text-base font-medium text-slate-500">ч</span>
                </>
              )}
            </div>
            <div className="pt-1">
              <Badge variant="indigo" className="text-[11px] py-0.5 px-2">
                {forecast?.scheduled_lessons ?? 0} запланировано уроков
              </Badge>
            </div>
          </div>
        </GlassCard>

        {/* Валовая выручка */}
        <GlassCard className="p-5 flex flex-col justify-between">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
              Потенциальная выручка (Gross)
            </span>
            <div className="w-9 h-9 rounded-2xl bg-slate-500/15 text-slate-700 flex items-center justify-center border border-slate-400/20">
              <TrendingUp className="w-4 h-4" />
            </div>
          </div>
          <div className="space-y-1">
            <div className="text-3xl font-bold text-slate-900">
              {isLoading ? (
                <div className="h-8 w-24 bg-slate-200/60 rounded-lg animate-pulse" />
              ) : (
                formatCurrency(forecast?.gross_potential_revenue)
              )}
            </div>
            <p className="text-[11px] text-slate-500 pt-1">
              Сумма по ставкам учеников до вычета партнерских комиссий
            </p>
          </div>
        </GlassCard>

        {/* Партнерские комиссии */}
        <GlassCard className="p-5 flex flex-col justify-between">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
              Комиссии партнерских школ
            </span>
            <div className="w-9 h-9 rounded-2xl bg-amber-500/15 text-amber-600 flex items-center justify-center border border-amber-400/20">
              <Percent className="w-4 h-4" />
            </div>
          </div>
          <div className="space-y-1">
            <div className="text-3xl font-bold text-amber-600">
              {isLoading ? (
                <div className="h-8 w-24 bg-slate-200/60 rounded-lg animate-pulse" />
              ) : (
                `- ${formatCurrency(forecast?.partner_commission_expected)}`
              )}
            </div>
            <p className="text-[11px] text-slate-500 pt-1">
              Ожидаемые отчисления партнерским школам по тегам
            </p>
          </div>
        </GlassCard>

        {/* Чистый доход (Net) */}
        <GlassCard className="p-5 flex flex-col justify-between relative overflow-hidden ring-2 ring-emerald-500/30">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-emerald-800">
              Ожидаемый чистый доход
            </span>
            <div className="w-9 h-9 rounded-2xl bg-emerald-500/20 text-emerald-700 flex items-center justify-center border border-emerald-400/30">
              <Wallet className="w-4 h-4" />
            </div>
          </div>
          <div className="space-y-1">
            <div className="text-3xl font-bold text-emerald-600">
              {isLoading ? (
                <div className="h-8 w-28 bg-slate-200/60 rounded-lg animate-pulse" />
              ) : (
                formatCurrency(forecast?.net_potential_income)
              )}
            </div>
            <div className="pt-1">
              <Badge variant="mint" className="text-[11px] py-0.5 px-2">
                Чистыми преподавателю
              </Badge>
            </div>
          </div>
        </GlassCard>
      </div>

      {/* Структура планируемых часов по форматам */}
      <GlassCard className="p-5 sm:p-6 space-y-4">
        <div className="flex items-center gap-3 border-b border-black/[0.05] pb-3">
          <div className="w-10 h-10 rounded-2xl bg-purple-500/15 text-purple-600 flex items-center justify-center border border-purple-400/20">
            <Sparkles className="w-5 h-5" />
          </div>
          <div>
            <h3 className="font-bold text-lg text-slate-900">
              Структура планируемой нагрузки по форматам
            </h3>
            <p className="text-xs text-slate-500">
              Распределение запланированных часов и ожидаемого дохода
            </p>
          </div>
        </div>

        {isLoading ? (
          <div className="p-8 text-center text-slate-400">
            <div className="w-6 h-6 border-2 border-indigo-600 border-t-transparent rounded-full animate-spin mx-auto mb-2" />
            <span>Загрузка форматов...</span>
          </div>
        ) : !forecast?.by_format || forecast.by_format.length === 0 ? (
          <div className="p-8 text-center text-slate-400 space-y-1.5">
            <Calendar className="w-8 h-8 text-slate-300 mx-auto" />
            <p className="text-sm font-semibold text-slate-700">Нет запланированных уроков</p>
            <p className="text-xs text-slate-400">
              На выбранный период в расписании нет занятий со статусом «Запланирован»
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            {/* Горизонтальный стек-бар */}
            {forecast.scheduled_hours > 0 && (
              <div className="h-4 rounded-xl overflow-hidden flex bg-black/[0.04] p-0.5 gap-0.5">
                {forecast.by_format.map((item) => {
                  const percent = (item.hours / forecast.scheduled_hours) * 100;
                  if (percent <= 0) return null;
                  return (
                    <div
                      key={item.format}
                      className={`h-full rounded-lg transition-all ${
                        formatColors[item.format] || 'bg-indigo-500'
                      }`}
                      style={{ width: `${percent}%` }}
                      title={`${formatLabels[item.format] || item.format}: ${formatHours(item.hours)} ч (${percent.toFixed(0)}%)`}
                    />
                  );
                })}
              </div>
            )}

            {/* Карточки по каждому формату */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 pt-2">
              {forecast.by_format.map((item) => {
                const percent =
                  forecast.scheduled_hours > 0
                    ? ((item.hours / forecast.scheduled_hours) * 100).toFixed(0)
                    : '0';

                return (
                  <div
                    key={item.format}
                    className="p-4 rounded-2xl bg-black/[0.02] border border-black/[0.04] flex flex-col justify-between space-y-3"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span
                          className={`w-3 h-3 rounded-full ${
                            formatColors[item.format] || 'bg-indigo-500'
                          }`}
                        />
                        <span className="font-bold text-sm text-slate-800">
                          {formatLabels[item.format] || item.format}
                        </span>
                      </div>
                      <Badge variant="neutral" className="text-[10px] py-0 px-2">
                        {percent}% часов
                      </Badge>
                    </div>

                    <div className="flex items-baseline justify-between pt-1 border-t border-black/[0.04]">
                      <div>
                        <div className="text-[10px] text-slate-400 uppercase font-semibold">
                          Часы
                        </div>
                        <div className="text-lg font-bold text-slate-900">
                          {formatHours(item.hours)} ч
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-[10px] text-slate-400 uppercase font-semibold">
                          Выручка
                        </div>
                        <div className="text-lg font-bold text-slate-800">
                          {formatCurrency(item.revenue)}
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </GlassCard>
    </div>
  );
};
