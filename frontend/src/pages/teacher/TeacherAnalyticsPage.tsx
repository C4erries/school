import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  getAnalyticsOverview,
  getAnalyticsDynamics,
  getAnalyticsFormats,
  getAnalyticsClients,
  getTagStats,
} from '../../api/analytics';
import {
  AnalyticsOverview,
  AnalyticsDynamicsPoint,
  AnalyticsFormatStat,
  AnalyticsClientStat,
  TagStat,
} from '../../types/analytics';
import { GlassButton } from '../../shared/components/GlassButton';
import {
  AnalyticsPeriodSelector,
  PeriodPreset,
} from '../../features/analytics/components/AnalyticsPeriodSelector';
import { AnalyticsKPICards } from '../../features/analytics/components/AnalyticsKPICards';
import { AnalyticsDynamicsChart } from '../../features/analytics/components/AnalyticsDynamicsChart';
import { AnalyticsFormatsChart } from '../../features/analytics/components/AnalyticsFormatsChart';
import { AnalyticsClientsTable } from '../../features/analytics/components/AnalyticsClientsTable';
import { AnalyticsTagsDistribution } from '../../features/analytics/components/AnalyticsTagsDistribution';
import { AnalyticsForecastTab } from '../../features/analytics/components/AnalyticsForecastTab';
import { CalendarSyncModal } from '../../features/calendar/components/CalendarSyncModal';
import { Calendar, RefreshCw, AlertCircle, BarChart3, Sparkles } from 'lucide-react';

export const TeacherAnalyticsPage: React.FC = () => {
  // Режим страницы: История (Факт) vs Прогноз (План)
  const [activeTab, setActiveTab] = useState<'fact' | 'forecast'>('fact');

  // Вычисление начальных дат для пресетов
  const computeDatesForPreset = useCallback((preset: PeriodPreset): { from: string; to: string } => {
    const now = new Date();
    const toStr = now.toISOString().split('T')[0];
    const y = now.getFullYear();
    const m = now.getMonth();

    if (preset === 'quarter') {
      const start = new Date(now);
      start.setDate(start.getDate() - 90);
      return { from: start.toISOString().split('T')[0], to: toStr };
    }
    if (preset === 'half_year') {
      const start = new Date(now);
      start.setDate(start.getDate() - 180);
      return { from: start.toISOString().split('T')[0], to: toStr };
    }
    if (preset === 'year') {
      const start = new Date(now);
      start.setDate(start.getDate() - 365);
      return { from: start.toISOString().split('T')[0], to: toStr };
    }

    // Default: 'month' (с 1 числа текущего месяца)
    const startOfMonth = new Date(y, m, 1);
    const startStr = `${startOfMonth.getFullYear()}-${String(startOfMonth.getMonth() + 1).padStart(2, '0')}-01`;
    return { from: startStr, to: toStr };
  }, []);

  const [preset, setPreset] = useState<PeriodPreset>('month');
  const initialDates = useMemo(() => computeDatesForPreset('month'), [computeDatesForPreset]);
  const [fromDate, setFromDate] = useState<string>(initialDates.from);
  const [toDate, setToDate] = useState<string>(initialDates.to);
  const [interval, setInterval] = useState<'week' | 'month'>('week');
  const [sortClients, setSortClients] = useState<'hours' | 'revenue' | 'cancellations'>('hours');

  // Состояния данных факта
  const [overview, setOverview] = useState<AnalyticsOverview | null>(null);
  const [dynamics, setDynamics] = useState<AnalyticsDynamicsPoint[]>([]);
  const [formats, setFormats] = useState<AnalyticsFormatStat[]>([]);
  const [clients, setClients] = useState<AnalyticsClientStat[]>([]);
  const [tags, setTags] = useState<TagStat[]>([]);

  // Состояния загрузки и ошибок
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSyncModalOpen, setIsSyncModalOpen] = useState(false);

  const handlePresetChange = (newPreset: PeriodPreset) => {
    setPreset(newPreset);
    if (newPreset !== 'custom') {
      const dates = computeDatesForPreset(newPreset);
      setFromDate(dates.from);
      setToDate(dates.to);
      if (newPreset === 'year' || newPreset === 'half_year') {
        setInterval('month');
      } else {
        setInterval('week');
      }
    }
  };

  const handleCustomDateChange = (from: string, to: string) => {
    setPreset('custom');
    setFromDate(from);
    setToDate(to);
  };

  const loadData = useCallback(async () => {
    setIsLoading(true);
    setErrorMessage(null);

    const fromIso = `${fromDate}T00:00:00Z`;
    const toIso = `${toDate}T23:59:59Z`;

    try {
      const [ovData, dynData, fmtData, clData, tagData] = await Promise.all([
        getAnalyticsOverview({ from: fromIso, to: toIso }).catch(() => null),
        getAnalyticsDynamics({ from: fromIso, to: toIso, interval }).catch(() => []),
        getAnalyticsFormats({ from: fromIso, to: toIso }).catch(() => []),
        getAnalyticsClients({ from: fromIso, to: toIso, sort: sortClients }).catch(() => []),
        getTagStats({ from: fromIso, to: toIso }).catch(() => []),
      ]);

      setOverview(ovData);
      setDynamics(dynData);
      setFormats(fmtData);
      setClients(clData);
      setTags(tagData);
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Ошибка загрузки аналитики');
    } finally {
      setIsLoading(false);
    }
  }, [fromDate, toDate, interval, sortClients]);

  useEffect(() => {
    if (activeTab === 'fact') {
      loadData();
    }
  }, [loadData, activeTab]);

  return (
    <div className="space-y-6">
      {/* Шапка страницы */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">
            Статистика и аналитика
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Умный анализ проведенных занятий, планирование нагрузки и финансовый прогноз
          </p>
        </div>

        <div className="flex items-center gap-2.5 self-stretch sm:self-auto">
          <GlassButton
            variant="secondary"
            onClick={() => setIsSyncModalOpen(true)}
            icon={<Calendar className="w-4 h-4 text-indigo-600" />}
          >
            Синхронизация с календарем
          </GlassButton>
          {activeTab === 'fact' && (
            <GlassButton
              variant="secondary"
              onClick={loadData}
              isLoading={isLoading}
              icon={<RefreshCw className="w-3.5 h-3.5 text-slate-600" />}
            >
              Обновить
            </GlassButton>
          )}
        </div>
      </div>

      {/* Segmented Control переключатель в стиле Apple Liquid Glass */}
      <div className="p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05] inline-flex items-center gap-1 backdrop-blur-md">
        <button
          type="button"
          onClick={() => setActiveTab('fact')}
          className={`flex items-center gap-2 px-5 py-2.5 rounded-xl text-xs sm:text-sm font-semibold transition-all select-none ${
            activeTab === 'fact'
              ? 'bg-white text-indigo-600 shadow-sm'
              : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
          }`}
        >
          <BarChart3 className="w-4 h-4" />
          <span>История (Факт)</span>
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('forecast')}
          className={`flex items-center gap-2 px-5 py-2.5 rounded-xl text-xs sm:text-sm font-semibold transition-all select-none ${
            activeTab === 'forecast'
              ? 'bg-white text-indigo-600 shadow-sm'
              : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
          }`}
        >
          <Sparkles className="w-4 h-4 text-purple-600" />
          <span>Прогноз (План)</span>
        </button>
      </div>

      {/* Вкладка 1: История (Факт) */}
      {activeTab === 'fact' && (
        <div className="space-y-6">
          {errorMessage && (
            <div className="p-4 rounded-3xl bg-rose-500/10 border border-rose-500/20 backdrop-blur-md flex items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
                <p className="text-sm font-medium text-rose-900">{errorMessage}</p>
              </div>
              <GlassButton
                variant="secondary"
                size="sm"
                onClick={loadData}
                icon={<RefreshCw className="w-3.5 h-3.5" />}
              >
                Повторить
              </GlassButton>
            </div>
          )}

          {/* Селектор периодов */}
          <AnalyticsPeriodSelector
            from={fromDate}
            to={toDate}
            interval={interval}
            preset={preset}
            onPresetChange={handlePresetChange}
            onDateChange={handleCustomDateChange}
            onIntervalChange={setInterval}
          />

          {/* 4 ключевые KPI карточки */}
          <AnalyticsKPICards overview={overview} isLoading={isLoading} />

          {/* Графики динамики и форматов */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2">
              <AnalyticsDynamicsChart data={dynamics} isLoading={isLoading} />
            </div>
            <div className="lg:col-span-1">
              <AnalyticsFormatsChart formats={formats} isLoading={isLoading} />
            </div>
          </div>

          {/* Доходность и нагрузка по тегам (ADR-009) */}
          <AnalyticsTagsDistribution tags={tags} isLoading={isLoading} />

          {/* Таблица рейтинга и показателей учеников */}
          <AnalyticsClientsTable
            clients={clients}
            sort={sortClients}
            onSortChange={setSortClients}
            isLoading={isLoading}
          />
        </div>
      )}

      {/* Вкладка 2: Прогноз (План) */}
      {activeTab === 'forecast' && <AnalyticsForecastTab />}

      {/* Модальное окно синхронизации календаря */}
      <CalendarSyncModal
        isOpen={isSyncModalOpen}
        onClose={() => setIsSyncModalOpen(false)}
      />
    </div>
  );
};

export default TeacherAnalyticsPage;
