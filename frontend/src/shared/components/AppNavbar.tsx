import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../../features/auth/useAuth';
import { GlassButton } from './GlassButton';
import { Badge } from './Badge';
import {
  GraduationCap,
  LogOut,
  Calendar,
  Building2,
  BookOpen,
  LayoutDashboard,
  Receipt,
  TrendingUp,
} from 'lucide-react';

export const AppNavbar: React.FC = () => {
  const { user, logout } = useAuth();
  const location = useLocation();

  const getRoleBadgeVariant = (role?: string) => {
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

  const getRoleLabel = (role?: string) => {
    switch (role) {
      case 'student':
        return 'Ученик';
      case 'teacher':
        return 'Преподаватель';
      case 'assistant':
        return 'Ассистент';
      case 'owner':
        return 'Администратор';
      default:
        return role || '';
    }
  };

  const isActive = (path: string) => location.pathname === path;

  return (
    <header className="flex flex-col md:flex-row items-center justify-between p-4 sm:p-5 rounded-3xl liquid-glass shadow-sm gap-4">
      <div className="flex items-center gap-3">
        <Link to="/dashboard" className="flex items-center gap-3 group">
          <div className="w-10 h-10 rounded-2xl bg-indigo-600/10 group-hover:bg-indigo-600/20 transition-colors flex items-center justify-center text-indigo-600">
            <GraduationCap className="w-5 h-5" />
          </div>
          <div>
            <span className="font-bold text-base tracking-tight text-slate-900 group-hover:text-indigo-600 transition-colors">
              Tutor Assistant
            </span>
            <span className="block text-xs text-slate-500">
              {getRoleLabel(user?.role)}
            </span>
          </div>
        </Link>
      </div>

      {/* Навигационные табы в стиле Apple Segmented Control */}
      <nav className="flex items-center gap-1.5 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05]">
        <Link
          to="/dashboard"
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
            isActive('/dashboard')
              ? 'bg-white text-indigo-600 shadow-sm'
              : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
          }`}
        >
          <LayoutDashboard className="w-3.5 h-3.5" />
          <span>Дашборд</span>
        </Link>

        {(user?.role === 'owner' || user?.role === 'assistant') && (
          <Link
            to="/admin"
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              isActive('/admin')
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            <Building2 className="w-3.5 h-3.5" />
            <span>Кабинеты и привязки</span>
          </Link>
        )}

        {(user?.role === 'teacher' || user?.role === 'owner') && (
          <Link
            to="/teacher/schedule"
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              isActive('/teacher/schedule')
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            <Calendar className="w-3.5 h-3.5" />
            <span>Расписание уроков</span>
          </Link>
        )}

        {(user?.role === 'teacher' || user?.role === 'owner') && (
          <Link
            to="/teacher/clients"
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              isActive('/teacher/clients')
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            <BookOpen className="w-3.5 h-3.5" />
            <span>Ученики</span>
          </Link>
        )}

        {(user?.role === 'teacher' || user?.role === 'owner') && (
          <Link
            to="/teacher/finance"
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              isActive('/teacher/finance')
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            <Receipt className="w-3.5 h-3.5" />
            <span>Бухгалтерия</span>
          </Link>
        )}

        {(user?.role === 'teacher' || user?.role === 'owner') && (
          <Link
            to="/teacher/analytics"
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              isActive('/teacher/analytics')
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
            }`}
          >
            <TrendingUp className="w-3.5 h-3.5" />
            <span>Статистика</span>
          </Link>
        )}
      </nav>

      {/* Пользователь и выход */}
      <div className="flex items-center gap-3">
        {user && (
          <div className="hidden lg:flex items-center gap-2">
            <span className="text-xs font-medium text-slate-700">{user.full_name}</span>
            <Badge variant={getRoleBadgeVariant(user.role)}>
              {getRoleLabel(user.role)}
            </Badge>
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
  );
};
