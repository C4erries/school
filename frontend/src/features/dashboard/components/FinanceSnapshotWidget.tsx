import React from 'react';
import { FinancialSnapshot } from '../../../types/dashboard';
import { GlassCard } from '../../../shared/components/GlassCard';
import { GlassButton } from '../../../shared/components/GlassButton';
import { Wallet, Sparkles, TrendingUp, AlertCircle, ArrowRight } from 'lucide-react';

interface FinanceSnapshotWidgetProps {
  financialSnapshot?: FinancialSnapshot;
  onNavigateToFinance: () => void;
}

export const FinanceSnapshotWidget: React.FC<FinanceSnapshotWidgetProps> = ({
  financialSnapshot,
  onNavigateToFinance,
}) => {
  const formatCurrency = (val?: number) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      maximumFractionDigits: 0,
    }).format(Math.round(val ?? 0));
  };

  return (
    <GlassCard className="p-5 sm:p-6 space-y-4 flex flex-col justify-between">
      <div className="flex items-center justify-between border-b border-black/[0.05] pb-3">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-2xl bg-emerald-500/15 text-emerald-600 flex items-center justify-center border border-emerald-400/20">
            <Wallet className="w-5 h-5" />
          </div>
          <div>
            <h3 className="font-bold text-base text-slate-900">Финансы месяца</h3>
            <p className="text-xs text-slate-500">Срез доходов и прогноз</p>
          </div>
        </div>
        <Sparkles className="w-4 h-4 text-purple-600" />
      </div>

      <div className="space-y-4 py-1">
        {/* Заработано факт */}
        <div className="p-3.5 rounded-2xl bg-emerald-500/10 border border-emerald-500/20 space-y-1">
          <div className="text-xs font-semibold text-emerald-800 uppercase tracking-wider">
            Факт заработано
          </div>
          <div className="text-2xl font-bold text-emerald-700">
            {formatCurrency(financialSnapshot?.month_earned)}
          </div>
          <div className="text-[11px] text-emerald-800/80">
            По проведенным занятиям текущего месяца
          </div>
        </div>

        {/* Прогноз до конца месяца */}
        <div className="p-3.5 rounded-2xl bg-indigo-500/10 border border-indigo-500/20 space-y-1">
          <div className="text-xs font-semibold text-indigo-800 uppercase tracking-wider flex items-center justify-between">
            <span>Ожидается прогноз</span>
            <TrendingUp className="w-3.5 h-3.5" />
          </div>
          <div className="text-2xl font-bold text-indigo-700">
            {formatCurrency(financialSnapshot?.month_forecast)}
          </div>
          <div className="text-[11px] text-indigo-800/80">
            Потенциал до конца месяца по запланированным урокам
          </div>
        </div>

        {/* Дебиторская задолженность */}
        <div className="p-3.5 rounded-2xl bg-rose-500/10 border border-rose-500/20 space-y-1">
          <div className="text-xs font-semibold text-rose-800 uppercase tracking-wider flex items-center justify-between">
            <span>Долги учеников</span>
            <AlertCircle className="w-3.5 h-3.5" />
          </div>
          <div className="text-2xl font-bold text-rose-600">
            {formatCurrency(financialSnapshot?.total_debts)}
          </div>
          <div className="text-[11px] text-rose-800/80">
            Сумма отрицательных балансов учеников
          </div>
        </div>
      </div>

      <GlassButton
        variant="secondary"
        className="w-full justify-between"
        onClick={onNavigateToFinance}
      >
        <span>Перейти в бухгалтерию</span>
        <ArrowRight className="w-4 h-4" />
      </GlassButton>
    </GlassCard>
  );
};
