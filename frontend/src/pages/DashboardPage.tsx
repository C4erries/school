import React, { useState, useEffect } from 'react';
import { useAuth } from '../features/auth/useAuth';
import { GlassCard } from '../shared/components/GlassCard';
import { GlassButton } from '../shared/components/GlassButton';
import { Badge } from '../shared/components/Badge';
import { checkBackendHealth } from '../api/health';
import { HealthResponse } from '../types/health';
import {
  GraduationCap,
  LogOut,
  Calendar,
  Users,
  Wallet,
  BookOpen,
  CheckCircle2,
  Clock,
  Activity,
  Sparkles,
} from 'lucide-react';

export const DashboardPage: React.FC = () => {
  const { user, logout } = useAuth();
  const [health, setHealth] = useState<HealthResponse | null>(null);

  useEffect(() => {
    const fetchHealth = async () => {
      try {
        const result = await checkBackendHealth();
        setHealth(result.data);
      } catch (err) {
        console.error('Health check error:', err);
      }
    };

    fetchHealth();
    const interval = setInterval(fetchHealth, 15000);
    return () => clearInterval(interval);
  }, []);

  const getRoleBadgeVariant = (role: string) => {
    switch (role) {
      case 'student':
        return 'mint';
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
      case 'student':
        return 'Ученик';
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

  return (
    <div className="min-h-screen bg-[#F5F5F7] text-slate-800 p-4 sm:p-8 relative overflow-hidden">
      {/* Оптические ауры Apple Liquid Glass */}
      <div className="absolute top-0 right-1/4 w-[500px] h-[500px] bg-indigo-500/10 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute bottom-10 left-10 w-[450px] h-[450px] bg-emerald-500/10 rounded-full blur-3xl pointer-events-none" />

      <div className="max-w-5xl mx-auto space-y-8 relative z-10">
        {/* Верхняя панель (Navbar) */}
        <header className="flex items-center justify-between p-4 sm:p-5 rounded-3xl liquid-glass shadow-sm">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-2xl bg-indigo-600/10 flex items-center justify-center text-indigo-600">
              <GraduationCap className="w-5 h-5" />
            </div>
            <div>
              <span className="font-bold text-base tracking-tight text-slate-900">
                School Platform
              </span>
              <span className="block text-xs text-slate-500">
                Личный кабинет
              </span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            {health && (
              <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 text-emerald-700 text-xs font-medium border border-emerald-500/20">
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                API: {health.status} ({health.service})
              </div>
            )}

            <GlassButton
              variant="secondary"
              size="sm"
              onClick={logout}
              icon={<LogOut className="w-4 h-4 text-slate-500" />}
            >
              Выйти
            </GlassButton>
          </div>
        </header>

        {/* Приветственный блок с профилем */}
        <GlassCard className="relative overflow-hidden shadow-sm">
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
                {user?.email} {user?.phone ? `• ${user.phone}` : ''}
              </p>
            </div>

            <div className="flex items-center gap-3">
              <div className="text-xs text-slate-500 bg-white/60 px-4 py-2 rounded-2xl border border-white/80">
                ID: <span className="font-mono text-slate-700">{user?.id.slice(0, 8)}...</span>
              </div>
            </div>
          </div>
        </GlassCard>

        {/* Разделы по ролям */}
        {user?.role === 'teacher' || user?.role === 'owner' ? (
          <div>
            <h2 className="text-lg font-bold text-slate-900 mb-4 flex items-center gap-2">
              <Sparkles className="w-5 h-5 text-indigo-600" />
              Инструменты преподавателя
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
              <GlassCard interactive className="space-y-3">
                <div className="w-10 h-10 rounded-2xl bg-indigo-500/10 text-indigo-600 flex items-center justify-center">
                  <Calendar className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Расписание занятий</h3>
                  <p className="text-xs text-slate-500 mt-1">
                    Слоты, перенос уроков и календарь оффлайн/онлайн
                  </p>
                </div>
                <div className="pt-2 text-xs font-semibold text-indigo-600">
                  Фаза 2 (Спринт) →
                </div>
              </GlassCard>

              <GlassCard interactive className="space-y-3">
                <div className="w-10 h-10 rounded-2xl bg-emerald-500/10 text-emerald-600 flex items-center justify-center">
                  <Users className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Ученики и группы</h3>
                  <p className="text-xs text-slate-500 mt-1">
                    Список закрепленных учеников, посещаемость
                  </p>
                </div>
                <div className="pt-2 text-xs font-semibold text-emerald-600">
                  Активно →
                </div>
              </GlassCard>

              <GlassCard interactive className="space-y-3">
                <div className="w-10 h-10 rounded-2xl bg-amber-500/10 text-amber-600 flex items-center justify-center">
                  <Wallet className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Финансовый дашборд</h3>
                  <p className="text-xs text-slate-500 mt-1">
                    Оплата за проведенные часы, расчет процента школы
                  </p>
                </div>
                <div className="pt-2 text-xs font-semibold text-amber-600">
                  Фаза 3 →
                </div>
              </GlassCard>
            </div>
          </div>
        ) : (
          <div>
            <h2 className="text-lg font-bold text-slate-900 mb-4 flex items-center gap-2">
              <BookOpen className="w-5 h-5 text-indigo-600" />
              Мое обучение
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
              <GlassCard interactive className="space-y-3">
                <div className="w-10 h-10 rounded-2xl bg-indigo-500/10 text-indigo-600 flex items-center justify-center">
                  <Calendar className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Мои занятия</h3>
                  <p className="text-xs text-slate-500 mt-1">
                    Запись на урок и расписание репетитора
                  </p>
                </div>
                <div className="pt-2 text-xs font-semibold text-indigo-600">
                  Слоты доступны →
                </div>
              </GlassCard>

              <GlassCard interactive className="space-y-3">
                <div className="w-10 h-10 rounded-2xl bg-emerald-500/10 text-emerald-600 flex items-center justify-center">
                  <CheckCircle2 className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Домашние задания</h3>
                  <p className="text-xs text-slate-500 mt-1">
                    Материалы уроков и отправка решений
                  </p>
                </div>
                <div className="pt-2 text-xs font-semibold text-emerald-600">
                  0 к сдаче →
                </div>
              </GlassCard>

              <GlassCard interactive className="space-y-3">
                <div className="w-10 h-10 rounded-2xl bg-amber-500/10 text-amber-600 flex items-center justify-center">
                  <Clock className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Баланс занятий</h3>
                  <p className="text-xs text-slate-500 mt-1">
                    Оплаченные пакеты уроков и история списаний
                  </p>
                </div>
                <div className="pt-2 text-xs font-semibold text-amber-600">
                  Пакет активен →
                </div>
              </GlassCard>
            </div>
          </div>
        )}

        {/* Статус системы */}
        <footer className="pt-6 border-t border-slate-200/60 flex flex-col sm:flex-row items-center justify-between text-xs text-slate-400 gap-2">
          <div className="flex items-center gap-2">
            <Activity className="w-4 h-4 text-emerald-500" />
            <span>Сессия валидирована через Valkey и JWT HS256</span>
          </div>
          <div>
            Школьная платформа © 2026 • Apple Liquid Glass UI
          </div>
        </footer>
      </div>
    </div>
  );
};
