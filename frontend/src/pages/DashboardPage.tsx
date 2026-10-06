import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../features/auth/useAuth';
import { GlassCard } from '../shared/components/GlassCard';
import { Badge } from '../shared/components/Badge';
import { getDashboardSummary } from '../api/dashboard';
import { DashboardSummary } from '../types/dashboard';
import { TodayLessonsWidget } from '../features/dashboard/components/TodayLessonsWidget';
import { FinanceSnapshotWidget } from '../features/dashboard/components/FinanceSnapshotWidget';
import { DashboardQuickActions } from '../features/dashboard/components/DashboardQuickActions';
import { Clock, Users } from 'lucide-react';

export const DashboardPage: React.FC = () => {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const fetchSummary = useCallback(async () => {
    try {
      const data = await getDashboardSummary();
      setSummary(data);
    } catch (err) {
      console.error('Failed to load dashboard summary:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchSummary();
  }, [fetchSummary]);

  const formatHours = (val?: number) => {
    const num = val ?? 0;
    return Number.isInteger(num) ? String(num) : num.toFixed(1);
  };

  const getRoleBadgeVariant = (role: string) => {
    switch (role) {
      case 'teacher':
        return 'indigo';
      case 'assistant':
        return 'amber';
      case 'owner':
        return 'coral';
      default:
        return 'neutral';
    }
  };

  const getRoleLabel = (role: string) => {
    switch (role) {
      case 'teacher':
        return 'Преподаватель';
      case 'assistant':
        return 'Ассистент';
      case 'owner':
        return 'Владелец школы';
      default:
        return role;
    }
  };

  const fin = summary?.financial_snapshot;
  const todayLessons = summary?.today_lessons || [];

  return (
    <div className="space-y-6">
      {/* Приветственный блок с профилем */}
      <GlassCard className="relative overflow-hidden">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="space-y-1.5">
            <div className="flex items-center gap-2.5">
              <h1 className="text-2xl font-bold text-slate-900 tracking-tight">
                Привет, {user?.full_name}!
              </h1>
              {user && (
                <Badge variant={getRoleBadgeVariant(user.role)}>
                  {getRoleLabel(user.role)}
                </Badge>
              )}
            </div>
            <p className="text-sm text-slate-500">
              Командный центр репетитора •{' '}
              {new Date().toLocaleDateString('ru-RU', {
                weekday: 'long',
                day: 'numeric',
                month: 'long',
              })}
            </p>
          </div>

          <div className="flex items-center gap-2">
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 text-emerald-800 text-xs font-medium border border-emerald-500/20 backdrop-blur-md">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              Онлайн платформа активна
            </div>
          </div>
        </div>
      </GlassCard>

      {/* Быстрые действия (Quick Actions) */}
      <DashboardQuickActions
        onNavigateToSchedule={() => navigate('/teacher/schedule')}
        onNavigateToFinance={() => navigate('/teacher/finance')}
        onNavigateToAnalytics={() => navigate('/teacher/analytics')}
      />

      {/* Основная сетка дашборда: 2 колонки (Уроки на сегодня + Финансовый статус) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Виджет 1: Ближайшие уроки на сегодня (Колонка 2/3) */}
        <div className="lg:col-span-2 space-y-4">
          <TodayLessonsWidget
            lessons={todayLessons}
            isLoading={isLoading}
            onNavigateToSchedule={() => navigate('/teacher/schedule')}
          />

          {/* Быстрые оперативные метрики */}
          <div className="grid grid-cols-2 gap-4">
            <GlassCard className="p-4 sm:p-5 flex items-center gap-3.5">
              <div className="w-10 h-10 rounded-2xl bg-indigo-500/15 text-indigo-600 flex items-center justify-center shrink-0 border border-indigo-400/20">
                <Clock className="w-5 h-5" />
              </div>
              <div>
                <div className="text-xs text-slate-500 uppercase font-semibold">
                  Нагрузка этой недели
                </div>
                <div className="text-xl font-bold text-slate-900 mt-0.5">
                  {formatHours(fin?.weekly_hours)} ч
                </div>
              </div>
            </GlassCard>

            <GlassCard className="p-4 sm:p-5 flex items-center gap-3.5">
              <div className="w-10 h-10 rounded-2xl bg-emerald-500/15 text-emerald-600 flex items-center justify-center shrink-0 border border-emerald-400/20">
                <Users className="w-5 h-5" />
              </div>
              <div>
                <div className="text-xs text-slate-500 uppercase font-semibold">
                  Активных учеников
                </div>
                <div className="text-xl font-bold text-slate-900 mt-0.5">
                  {fin?.active_clients_count ?? 0}
                </div>
              </div>
            </GlassCard>
          </div>
        </div>

        {/* Виджет 2: Финансовый статус месяца (Колонка 1/3) */}
        <div className="space-y-4">
          <FinanceSnapshotWidget
            financialSnapshot={fin}
            onNavigateToFinance={() => navigate('/teacher/finance')}
          />
        </div>
      </div>

      {/* Подвал */}
      <footer className="pt-6 border-t border-slate-200/40 flex flex-col sm:flex-row items-center justify-between text-xs text-slate-400 gap-2">
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-emerald-500 inline-block" />
          <span>Система активна • Tutor Assistant SaaS 2026</span>
        </div>
        <div>Школьная платформа © 2026</div>
      </footer>
    </div>
  );
};

export default DashboardPage;
