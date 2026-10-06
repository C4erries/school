import React from 'react';
import { PlusCircle, Wallet, BarChart3, ChevronRight } from 'lucide-react';

interface DashboardQuickActionsProps {
  onNavigateToSchedule: () => void;
  onNavigateToFinance: () => void;
  onNavigateToAnalytics: () => void;
}

export const DashboardQuickActions: React.FC<DashboardQuickActionsProps> = ({
  onNavigateToSchedule,
  onNavigateToFinance,
  onNavigateToAnalytics,
}) => {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
      <button
        type="button"
        onClick={onNavigateToSchedule}
        className="p-4 rounded-2xl liquid-glass-interactive flex items-center justify-between text-left group transition-all"
      >
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-indigo-500/15 text-indigo-600 flex items-center justify-center border border-indigo-400/20 group-hover:scale-110 transition-transform">
            <PlusCircle className="w-5 h-5" />
          </div>
          <div>
            <div className="font-bold text-sm text-slate-900">Запланировать урок</div>
            <div className="text-xs text-slate-500">В сетку расписания</div>
          </div>
        </div>
        <ChevronRight className="w-4 h-4 text-slate-400 group-hover:translate-x-0.5 transition-transform" />
      </button>

      <button
        type="button"
        onClick={onNavigateToFinance}
        className="p-4 rounded-2xl liquid-glass-interactive flex items-center justify-between text-left group transition-all"
      >
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-emerald-500/15 text-emerald-600 flex items-center justify-center border border-emerald-400/20 group-hover:scale-110 transition-transform">
            <Wallet className="w-5 h-5" />
          </div>
          <div>
            <div className="font-bold text-sm text-slate-900">Внести оплату</div>
            <div className="text-xs text-slate-500">Журнал и балансы</div>
          </div>
        </div>
        <ChevronRight className="w-4 h-4 text-slate-400 group-hover:translate-x-0.5 transition-transform" />
      </button>

      <button
        type="button"
        onClick={onNavigateToAnalytics}
        className="p-4 rounded-2xl liquid-glass-interactive flex items-center justify-between text-left group transition-all"
      >
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-purple-500/15 text-purple-600 flex items-center justify-center border border-purple-400/20 group-hover:scale-110 transition-transform">
            <BarChart3 className="w-5 h-5" />
          </div>
          <div>
            <div className="font-bold text-sm text-slate-900">Смотреть статистику</div>
            <div className="text-xs text-slate-500">Факт и прогноз дохода</div>
          </div>
        </div>
        <ChevronRight className="w-4 h-4 text-slate-400 group-hover:translate-x-0.5 transition-transform" />
      </button>
    </div>
  );
};
