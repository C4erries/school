import React from 'react';
import { GlassCard } from '../../../shared/components/GlassCard';
import { AnalyticsFormatStat } from '../../../types/analytics';
import { Layers, User, Users, UserCheck } from 'lucide-react';

interface AnalyticsFormatsChartProps {
  formats: AnalyticsFormatStat[];
  isLoading?: boolean;
}

export const AnalyticsFormatsChart: React.FC<AnalyticsFormatsChartProps> = ({
  formats,
  isLoading = false,
}) => {
  const getFormatLabel = (fmt: string) => {
    switch (fmt) {
      case 'individual':
        return 'Индивидуальные';
      case 'pair':
        return 'Парные';
      case 'group':
        return 'Мини-группы';
      case 'online':
        return 'Онлайн';
      case 'offline':
        return 'Офлайн';
      default:
        return fmt;
    }
  };

  const getFormatIcon = (fmt: string) => {
    switch (fmt) {
      case 'individual':
        return <User className="w-4 h-4 text-indigo-600" />;
      case 'pair':
        return <UserCheck className="w-4 h-4 text-emerald-600" />;
      case 'group':
        return <Users className="w-4 h-4 text-amber-600" />;
      default:
        return <Layers className="w-4 h-4 text-slate-600" />;
    }
  };

  const getFormatColors = (fmt: string) => {
    switch (fmt) {
      case 'individual':
        return {
          bar: 'bg-gradient-to-r from-indigo-500 to-indigo-600',
          bg: 'bg-indigo-500/10',
          text: 'text-indigo-800',
        };
      case 'pair':
        return {
          bar: 'bg-gradient-to-r from-emerald-500 to-teal-500',
          bg: 'bg-emerald-500/10',
          text: 'text-emerald-800',
        };
      case 'group':
        return {
          bar: 'bg-gradient-to-r from-amber-500 to-orange-500',
          bg: 'bg-amber-500/10',
          text: 'text-amber-800',
        };
      default:
        return {
          bar: 'bg-gradient-to-r from-slate-500 to-slate-600',
          bg: 'bg-slate-500/10',
          text: 'text-slate-800',
        };
    }
  };

  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      maximumFractionDigits: 0,
    }).format(val);
  };

  return (
    <GlassCard className="p-5 sm:p-6 space-y-4">
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-xl bg-indigo-500/10 text-indigo-600 flex items-center justify-center">
              <Layers className="w-4 h-4" />
            </div>
            <h3 className="text-base font-bold text-slate-900 tracking-tight">
              Структура по форматам
            </h3>
          </div>
          <p className="text-xs text-slate-500">
            Распределение нагрузки и выручки по форматам занятий
          </p>
        </div>
      </div>

      {isLoading ? (
        <div className="py-8 text-center text-slate-400 text-sm">
          Загрузка структуры форматов...
        </div>
      ) : formats.length === 0 ? (
        <div className="py-8 text-center text-slate-400 text-sm">
          Нет проведенных уроков в выбранном периоде
        </div>
      ) : (
        <div className="space-y-4 pt-1">
          {formats.map((f) => {
            const colors = getFormatColors(f.format);
            const hoursShare = Math.round(f.hours_share_percent);
            const revenueShare = Math.round(f.revenue_share_percent);

            return (
              <div
                key={f.format}
                className="p-3.5 rounded-2xl bg-black/[0.02] border border-black/[0.04] space-y-2.5 transition-all hover:bg-black/[0.04]"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2.5">
                    <div className={`p-1.5 rounded-xl ${colors.bg}`}>
                      {getFormatIcon(f.format)}
                    </div>
                    <div>
                      <div className="font-semibold text-xs text-slate-900">
                        {getFormatLabel(f.format)}
                      </div>
                      <div className="text-[11px] text-slate-500">
                        {f.lessons_count} {f.lessons_count === 1 ? 'урок' : 'уроков'} • {f.completed_hours} ч
                      </div>
                    </div>
                  </div>

                  <div className="text-right">
                    <div className="font-bold text-xs text-slate-900">
                      {formatCurrency(f.net_income)}
                    </div>
                    <div className="text-[11px] text-emerald-700 font-medium">
                      {revenueShare}% выручки
                    </div>
                  </div>
                </div>

                {/* Прогресс-бар доли часов */}
                <div className="space-y-1">
                  <div className="flex items-center justify-between text-[10px] text-slate-500 font-medium">
                    <span>Доля нагрузки: {hoursShare}%</span>
                    <span>{f.completed_hours} ч</span>
                  </div>
                  <div className="w-full h-2 rounded-full bg-black/[0.06] overflow-hidden">
                    <div
                      className={`h-full rounded-full ${colors.bar} transition-all duration-500`}
                      style={{ width: `${Math.min(hoursShare, 100)}%` }}
                    />
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </GlassCard>
  );
};
