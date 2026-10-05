import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../features/auth/useAuth';
import { GlassCard } from '../shared/components/GlassCard';
import { Badge } from '../shared/components/Badge';
import { checkBackendHealth } from '../api/health';
import { getFinancialDashboard } from '../api/schedule';
import { HealthResponse } from '../types/health';
import { FinancialDashboardStats } from '../types/schedule';
import {
  Calendar,
  Users,
  Wallet,
  Sparkles,
  Building2,
  ArrowRight,
} from 'lucide-react';

export const DashboardPage: React.FC = () => {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [finStats, setFinStats] = useState<FinancialDashboardStats | null>(null);

  useEffect(() => {
    const fetchHealth = async () => {
      try {
        const result = await checkBackendHealth();
        setHealth(result.data);
      } catch (err) {
        console.error('Health check error:', err);
      }
    };
    
    const fetchStats = async () => {
      if (user?.role === 'teacher') {
        const stats = await getFinancialDashboard();
        setFinStats(stats);
      }
    };

    fetchHealth();
    fetchStats();
    const interval = setInterval(fetchHealth, 15000);
    return () => clearInterval(interval);
  }, [user?.role]);

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
                {user?.email} {user?.phone ? `• ${user.phone}` : ''}
              </p>
            </div>

            <div className="flex items-center gap-3">
              {health && (
                <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 text-emerald-800 text-xs font-medium border border-emerald-500/20 backdrop-blur-md">
                  <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                  Система подключена
                </div>
              )}
            </div>
          </div>
        </GlassCard>

        {/* Разделы для Администратора / Владельца */}
        {(user?.role === 'owner' || user?.role === 'assistant') && (
          <div>
            <h2 className="text-lg font-bold text-slate-900 mb-4 flex items-center gap-2">
              <Sparkles className="w-5 h-5 text-indigo-600" />
              Управление платформой
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
              <GlassCard
                interactive
                onClick={() => navigate('/admin')}
                className="space-y-3 group"
              >
                <div className="w-10 h-10 rounded-2xl bg-indigo-500/10 text-indigo-600 flex items-center justify-center group-hover:scale-110 transition-transform">
                  <Building2 className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Кабинеты и привязка учеников</h3>
                  <p className="text-xs text-slate-500 mt-1">
                    Настройка школьных аудиторий, вместимость, цветовые метки и распределение учеников
                  </p>
                </div>
                <div className="pt-2 text-xs font-semibold text-indigo-600 flex items-center gap-1">
                  Перейти в панель администратора <ArrowRight className="w-3.5 h-3.5" />
                </div>
              </GlassCard>

              <GlassCard
                interactive
                onClick={() => navigate('/teacher/schedule')}
                className="space-y-3 group"
              >
                <div className="w-10 h-10 rounded-2xl bg-emerald-500/10 text-emerald-600 flex items-center justify-center group-hover:scale-110 transition-transform">
                  <Calendar className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Сетка расписания</h3>
                  <p className="text-xs text-slate-500 mt-1">
                    Просмотр расписания внахлёст, онлайн и оффлайн слоты
                  </p>
                </div>
                <div className="pt-2 text-xs font-semibold text-emerald-600 flex items-center gap-1">
                  Смотреть расписание <ArrowRight className="w-3.5 h-3.5" />
                </div>
              </GlassCard>
            </div>
          </div>
        )}

        {/* Разделы для Преподавателя */}
        {user?.role === 'teacher' && (
          <div>
            <h2 className="text-lg font-bold text-slate-900 mb-4 flex items-center gap-2">
              <Sparkles className="w-5 h-5 text-indigo-600" />
              Инструменты преподавателя
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
              <GlassCard
                interactive
                onClick={() => navigate('/teacher/schedule')}
                className="space-y-3 group"
              >
                <div className="w-10 h-10 rounded-2xl bg-indigo-500/10 text-indigo-600 flex items-center justify-center group-hover:scale-110 transition-transform">
                  <Calendar className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Расписание занятий</h3>
                  <p className="text-xs text-slate-500 mt-1">
                    Назначение уроков, сетка занятий, отметка «Проведён» и «Неявка»
                  </p>
                </div>
                <div className="pt-2 text-xs font-semibold text-indigo-600 flex items-center gap-1">
                  Открыть расписание <ArrowRight className="w-3.5 h-3.5" />
                </div>
              </GlassCard>

              <GlassCard
                interactive
                onClick={() => navigate('/teacher/clients')}
                className="space-y-3 group"
              >
                <div className="w-10 h-10 rounded-2xl bg-emerald-500/10 text-emerald-600 flex items-center justify-center group-hover:scale-110 transition-transform">
                  <Users className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 text-base">CRM Репетитора</h3>
                  <p className="text-xs text-slate-500 mt-1">
                    Управление учениками, абонементами и ставками
                  </p>
                </div>
                <div className="pt-2 text-xs font-semibold text-emerald-600 flex items-center gap-1">
                  Список учеников <ArrowRight className="w-3.5 h-3.5" />
                </div>
              </GlassCard>

              <GlassCard className="space-y-3">
                <div className="w-10 h-10 rounded-2xl bg-amber-500/10 text-amber-600 flex items-center justify-center">
                  <Wallet className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Финансы</h3>
                  <div className="mt-2 space-y-1">
                    <p className="text-xs text-slate-500 flex justify-between">
                      Gross Revenue: <span className="font-bold text-slate-800">{finStats?.gross_revenue || 0} ₽</span>
                    </p>
                    <p className="text-xs text-slate-500 flex justify-between">
                      Net Income: <span className="font-bold text-emerald-600">{finStats?.net_income || 0} ₽</span>
                    </p>
                    <p className="text-xs text-slate-500 flex justify-between">
                      Avg Rate: <span className="font-bold text-indigo-600">{finStats?.average_hourly_rate || 0} ₽/ч</span>
                    </p>
                  </div>
                </div>
              </GlassCard>
            </div>
          </div>
        )}



        {/* Подвал */}
        <footer className="pt-6 border-t border-slate-200/40 flex flex-col sm:flex-row items-center justify-between text-xs text-slate-400 gap-2">
          <div className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-emerald-500 inline-block" />
            <span>Система активна</span>
          </div>
          <div>
            Школьная платформа © 2026
          </div>
        </footer>
    </div>
  );
};
