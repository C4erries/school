import React, { useState, useMemo } from 'react';
import { GlassCard } from '../../../shared/components/GlassCard';
import { AnalyticsDynamicsPoint } from '../../../types/analytics';
import { Clock, Wallet, BarChart3 } from 'lucide-react';

interface AnalyticsDynamicsChartProps {
  data: AnalyticsDynamicsPoint[];
  isLoading?: boolean;
}

export const AnalyticsDynamicsChart: React.FC<AnalyticsDynamicsChartProps> = ({
  data,
  isLoading = false,
}) => {
  const [metric, setMetric] = useState<'income' | 'hours'>('income');
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      maximumFractionDigits: 0,
    }).format(val);
  };

  const chartPoints = useMemo(() => {
    if (!data || data.length === 0) return [];
    return data;
  }, [data]);

  // Вычисление масштабов SVG
  const width = 800;
  const height = 240;
  const paddingX = 40;
  const paddingTop = 25;
  const paddingBottom = 40;

  const chartW = width - paddingX * 2;
  const chartH = height - paddingTop - paddingBottom;

  const maxVal = useMemo(() => {
    if (chartPoints.length === 0) return 100;
    const vals = chartPoints.map((p) =>
      metric === 'income' ? p.net_income : p.completed_hours
    );
    const m = Math.max(...vals, 0);
    return m === 0 ? (metric === 'income' ? 10000 : 10) : m * 1.15;
  }, [chartPoints, metric]);

  const barWidth = useMemo(() => {
    if (chartPoints.length === 0) return 30;
    const step = chartW / chartPoints.length;
    return Math.min(Math.max(step * 0.55, 12), 48);
  }, [chartPoints, chartW]);

  const hoveredPoint = hoveredIndex !== null ? chartPoints[hoveredIndex] : null;

  return (
    <GlassCard className="p-5 sm:p-6 space-y-4">
      {/* Шапка графика */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-xl bg-indigo-500/10 text-indigo-600 flex items-center justify-center">
              <BarChart3 className="w-4 h-4" />
            </div>
            <h3 className="text-base font-bold text-slate-900 tracking-tight">
              Динамика нагрузки и выручки
            </h3>
          </div>
          <p className="text-xs text-slate-500">
            Распределение отработанных часов и полученного чистого дохода во времени
          </p>
        </div>

        {/* Переключатель метрики */}
        <div className="flex items-center gap-1 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05] self-start sm:self-auto">
          <button
            type="button"
            onClick={() => setMetric('income')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              metric === 'income'
                ? 'bg-white text-emerald-700 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            <Wallet className="w-3.5 h-3.5" />
            <span>Доход</span>
          </button>
          <button
            type="button"
            onClick={() => setMetric('hours')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              metric === 'hours'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            <Clock className="w-3.5 h-3.5" />
            <span>Часы</span>
          </button>
        </div>
      </div>

      {/* SVG холст графика */}
      <div className="relative w-full overflow-hidden select-none">
        {isLoading ? (
          <div className="h-[240px] flex items-center justify-center text-slate-400 text-sm">
            Загрузка графика...
          </div>
        ) : chartPoints.length === 0 ? (
          <div className="h-[240px] flex items-center justify-center text-slate-400 text-sm">
            Нет данных за указанный период
          </div>
        ) : (
          <div className="w-full relative">
            <svg
              viewBox={`0 0 ${width} ${height}`}
              className="w-full h-auto overflow-visible"
              preserveAspectRatio="xMidYMid meet"
              onMouseLeave={() => setHoveredIndex(null)}
            >
              <defs>
                <linearGradient id="barGradientIncome" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#10b981" stopOpacity="0.85" />
                  <stop offset="100%" stopColor="#059669" stopOpacity="0.4" />
                </linearGradient>
                <linearGradient id="barGradientHours" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#6366f1" stopOpacity="0.85" />
                  <stop offset="100%" stopColor="#4f46e5" stopOpacity="0.4" />
                </linearGradient>
                <linearGradient id="areaGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop
                    offset="0%"
                    stopColor={metric === 'income' ? '#10b981' : '#6366f1'}
                    stopOpacity="0.2"
                  />
                  <stop
                    offset="100%"
                    stopColor={metric === 'income' ? '#10b981' : '#6366f1'}
                    stopOpacity="0.0"
                  />
                </linearGradient>
              </defs>

              {/* Горизонтальные направляющие линии */}
              {[0, 0.25, 0.5, 0.75, 1].map((pct, i) => {
                const y = paddingTop + chartH * (1 - pct);
                const val = maxVal * pct;
                const label =
                  metric === 'income'
                    ? `${Math.round(val / 1000)}k ₽`
                    : `${Math.round(val)} ч`;
                return (
                  <g key={i}>
                    <line
                      x1={paddingX}
                      y1={y}
                      x2={width - paddingX}
                      y2={y}
                      stroke="rgba(0,0,0,0.06)"
                      strokeDasharray="4 4"
                    />
                    <text
                      x={paddingX - 8}
                      y={y + 3}
                      textAnchor="end"
                      fontSize="10"
                      fill="#94a3b8"
                      className="font-medium select-none"
                    >
                      {label}
                    </text>
                  </g>
                );
              })}

              {/* Столбцы графика */}
              {chartPoints.map((p, idx) => {
                const xCenter =
                  paddingX + (chartW / chartPoints.length) * (idx + 0.5);
                const x = xCenter - barWidth / 2;
                const val = metric === 'income' ? p.net_income : p.completed_hours;
                const barH = Math.max((val / maxVal) * chartH, 4);
                const y = paddingTop + chartH - barH;
                const isHovered = hoveredIndex === idx;

                return (
                  <g
                    key={idx}
                    className="cursor-pointer transition-opacity"
                    onMouseEnter={() => setHoveredIndex(idx)}
                  >
                    {/* Невидимая область для легкого наведения */}
                    <rect
                      x={paddingX + (chartW / chartPoints.length) * idx}
                      y={paddingTop}
                      width={chartW / chartPoints.length}
                      height={chartH}
                      fill="transparent"
                    />

                    {/* Подсветка колонки при наведении */}
                    {isHovered && (
                      <rect
                        x={paddingX + (chartW / chartPoints.length) * idx}
                        y={paddingTop}
                        width={chartW / chartPoints.length}
                        height={chartH}
                        fill="rgba(99, 102, 241, 0.05)"
                        rx="6"
                      />
                    )}

                    {/* Сам столбец */}
                    <rect
                      x={x}
                      y={y}
                      width={barWidth}
                      height={barH}
                      rx="6"
                      fill={
                        metric === 'income'
                          ? 'url(#barGradientIncome)'
                          : 'url(#barGradientHours)'
                      }
                      className="transition-all duration-300"
                      opacity={hoveredIndex === null || isHovered ? 1 : 0.45}
                    />

                    {/* Текстовая подпись оси X */}
                    <text
                      x={xCenter}
                      y={height - 12}
                      textAnchor="middle"
                      fontSize="10"
                      fill={isHovered ? '#1e293b' : '#64748b'}
                      fontWeight={isHovered ? 'bold' : 'normal'}
                      className="select-none"
                    >
                      {p.label.length > 14 ? p.label.slice(0, 13) + '…' : p.label}
                    </text>
                  </g>
                );
              })}
            </svg>

            {/* Интерактивный тултип */}
            {hoveredPoint && (
              <div
                className="absolute top-2 right-4 p-3 rounded-2xl liquid-glass border border-white/80 shadow-lg text-xs space-y-1.5 animate-in fade-in duration-150 pointer-events-none z-10"
              >
                <div className="font-bold text-slate-800 border-b border-black/[0.06] pb-1">
                  {hoveredPoint.label}
                </div>
                <div className="flex items-center justify-between gap-4 text-emerald-800 font-semibold">
                  <span>Чистый доход:</span>
                  <span>{formatCurrency(hoveredPoint.net_income)}</span>
                </div>
                <div className="flex items-center justify-between gap-4 text-indigo-800 font-medium">
                  <span>Отработано часов:</span>
                  <span>{hoveredPoint.completed_hours} ч</span>
                </div>
                <div className="flex items-center justify-between gap-4 text-slate-600 text-[11px]">
                  <span>Проведено уроков:</span>
                  <span>{hoveredPoint.completed_count}</span>
                </div>
                {hoveredPoint.cancelled_count > 0 && (
                  <div className="flex items-center justify-between gap-4 text-rose-600 text-[11px]">
                    <span>Отменено:</span>
                    <span>{hoveredPoint.cancelled_count}</span>
                  </div>
                )}
              </div>
            )}
          </div>
        )}
      </div>
    </GlassCard>
  );
};
