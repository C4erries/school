import React, { useState } from 'react';
import { PartnerSettlement } from '../../../types/finance';
import { GlassCard } from '../../../shared/components/GlassCard';
import { GlassButton } from '../../../shared/components/GlassButton';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassInput } from '../../../shared/components/GlassInput';
import { Badge } from '../../../shared/components/Badge';
import { createPartnerPayout } from '../../../api/finance';
import {
  Building2,
  CheckCircle2,
  Clock,
  Send,
  BookOpen,
  PieChart,
} from 'lucide-react';

interface PartnerSettlementsCardProps {
  settlements: PartnerSettlement[];
  month: string;
  onPayoutSuccess: () => Promise<void>;
  isLoading?: boolean;
}

export const PartnerSettlementsCard: React.FC<PartnerSettlementsCardProps> = ({
  settlements,
  onPayoutSuccess,
  isLoading = false,
}) => {
  const [payoutTarget, setPayoutTarget] = useState<PartnerSettlement | null>(null);
  const [notes, setNotes] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      maximumFractionDigits: 0,
    }).format(val);
  };

  const handleOpenPayout = (s: PartnerSettlement) => {
    setPayoutTarget(s);
    setNotes('');
  };

  const handleConfirmPayout = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!payoutTarget) return;

    setIsSubmitting(true);
    try {
      await createPartnerPayout({
        tag_id: payoutTarget.tag_id,
        period_month: payoutTarget.period_month,
        gross_amount: payoutTarget.gross_amount,
        commission_amount: payoutTarget.commission_amount,
        paid_at: new Date().toISOString(),
        notes: notes.trim() || null,
      });
      await onPayoutSuccess();
      setPayoutTarget(null);
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка подтверждения выплаты');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Согласно ADR-009 и ТЗ спринта 2.2.3: исключаем теги с комиссией <= 0 (чисто информационные)
  const activeSettlements = settlements.filter((s) => (s.school_percent || 0) > 0);

  const totalCommissions = activeSettlements.reduce((acc, s) => acc + (s.commission_amount || 0), 0);
  const totalPaid = activeSettlements
    .filter((s) => s.is_paid)
    .reduce((acc, s) => acc + (s.commission_amount || 0), 0);
  const totalUnpaid = totalCommissions - totalPaid;

  return (
    <div className="space-y-4">
      {/* Шапка раздела */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-2xl bg-amber-500/15 text-amber-600 flex items-center justify-center border border-amber-400/20">
            <Building2 className="w-5 h-5" />
          </div>
          <div>
            <h2 className="font-bold text-lg text-slate-900">Взаиморасчеты с партнерскими школами</h2>
            <p className="text-xs text-slate-500">
              Комиссионные отчисления по партнерским тегам за расчетный месяц
            </p>
          </div>
        </div>

        <div className="flex items-center gap-3 text-xs">
          <span className="text-slate-500">
            К выплате:{' '}
            <strong className="text-amber-600 font-bold">
              {formatCurrency(totalUnpaid)}
            </strong>
          </span>
          <span className="text-slate-300">|</span>
          <span className="text-slate-500">
            Выплачено:{' '}
            <strong className="text-emerald-600 font-bold">
              {formatCurrency(totalPaid)}
            </strong>
          </span>
        </div>
      </div>

      {/* Список партнерских школ */}
      {isLoading ? (
        <GlassCard className="p-8 text-center text-slate-400">
          <div className="w-6 h-6 border-2 border-indigo-600 border-t-transparent rounded-full animate-spin mx-auto mb-2" />
          <span>Загрузка взаиморасчетов...</span>
        </GlassCard>
      ) : activeSettlements.length === 0 ? (
        <GlassCard className="p-10 text-center text-slate-500 space-y-2">
          <PieChart className="w-8 h-8 text-slate-300 mx-auto" />
          <p className="font-semibold text-sm">Нет данных по партнерским школам</p>
          <p className="text-xs text-slate-400">
            Привяжите учеников к тегам школ с процентом комиссии, чтобы отслеживать взаиморасчеты
          </p>
        </GlassCard>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {activeSettlements.map((s) => {
            const netIncome = (s.gross_amount || 0) - (s.commission_amount || 0);

            return (
              <GlassCard
                key={s.tag_id}
                className="p-5 flex flex-col justify-between space-y-4 hover:shadow-md transition-shadow relative overflow-hidden"
              >
                <div>
                  {/* Верхняя строка: Название школы и статус */}
                  <div className="flex items-start justify-between gap-2 mb-3">
                    <div className="flex items-center gap-2">
                      <div className="w-8 h-8 rounded-xl bg-indigo-50 text-indigo-700 flex items-center justify-center font-bold text-xs border border-indigo-100">
                        {s.tag_name.charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <h3 className="font-bold text-base text-slate-900 leading-snug">
                          {s.tag_name}
                        </h3>
                        <span className="text-xs text-slate-500 font-medium">
                          Комиссия школы: {s.school_percent}%
                        </span>
                      </div>
                    </div>

                    {s.is_paid ? (
                      <Badge variant="mint" className="gap-1">
                        <CheckCircle2 className="w-3 h-3 text-emerald-600" />
                        Оплачено
                      </Badge>
                    ) : (
                      <Badge variant="amber" className="gap-1">
                        <Clock className="w-3 h-3 text-amber-600" />
                        К выплате
                      </Badge>
                    )}
                  </div>

                  {/* Сводка финансовых показателей */}
                  <div className="grid grid-cols-3 gap-2 p-3 rounded-2xl bg-black/[0.03] border border-black/[0.04] text-center my-2">
                    <div>
                      <div className="text-[10px] text-slate-500 font-semibold uppercase">
                        Уроков
                      </div>
                      <div className="text-sm font-bold text-slate-800 flex items-center justify-center gap-1 mt-0.5">
                        <BookOpen className="w-3 h-3 text-slate-400" />
                        {s.lessons_count}
                      </div>
                      <div className="text-[10px] text-slate-400">проведено</div>
                    </div>

                    <div>
                      <div className="text-[10px] text-slate-500 font-semibold uppercase">
                        Выручка
                      </div>
                      <div className="text-sm font-bold text-slate-800 mt-0.5">
                        {formatCurrency(s.gross_amount)}
                      </div>
                      <div className="text-[10px] text-slate-400">Gross</div>
                    </div>

                    <div>
                      <div className="text-[10px] text-slate-500 font-semibold uppercase">
                        Комиссия
                      </div>
                      <div className="text-sm font-bold text-amber-600 mt-0.5">
                        {formatCurrency(s.commission_amount)}
                      </div>
                      <div className="text-[10px] text-slate-400">Школе</div>
                    </div>
                  </div>

                  <div className="flex items-center justify-between text-xs px-1 text-slate-600">
                    <span>Чистый доход преподавателя:</span>
                    <strong className="text-emerald-700 font-bold">
                      {formatCurrency(netIncome)}
                    </strong>
                  </div>
                </div>

                {/* Действие: Отметить выплату */}
                <div className="pt-3 border-t border-black/[0.05] flex items-center justify-between">
                  {s.is_paid ? (
                    <span className="text-xs text-emerald-700 font-medium flex items-center gap-1">
                      <CheckCircle2 className="w-3.5 h-3.5" />
                      Выплата зафиксирована
                    </span>
                  ) : (
                    <span className="text-xs text-amber-700 font-medium">
                      Требуется перечислить комиссию
                    </span>
                  )}

                  {!s.is_paid && (
                    <GlassButton
                      variant="primary"
                      size="sm"
                      onClick={() => handleOpenPayout(s)}
                      icon={<Send className="w-3.5 h-3.5" />}
                    >
                      Отметить выплату
                    </GlassButton>
                  )}
                </div>
              </GlassCard>
            );
          })}
        </div>
      )}

      {/* Модалка фиксации выплаты */}
      {payoutTarget && (
        <GlassModal
          isOpen={Boolean(payoutTarget)}
          onClose={() => setPayoutTarget(null)}
          title={`Выплата комиссии школе «${payoutTarget.tag_name}»`}
          description={`Фиксация перечисления ${formatCurrency(payoutTarget.commission_amount)} за период ${payoutTarget.period_month}`}
        >
          <form onSubmit={handleConfirmPayout} className="space-y-4">
            <div className="p-3 rounded-2xl bg-black/[0.03] border border-black/[0.05] space-y-1.5 text-xs">
              <div className="flex justify-between text-slate-600">
                <span>Проведено уроков:</span>
                <strong className="text-slate-800">{payoutTarget.lessons_count}</strong>
              </div>
              <div className="flex justify-between text-slate-600">
                <span>Общий оборот (Gross):</span>
                <strong className="text-slate-800">{formatCurrency(payoutTarget.gross_amount)}</strong>
              </div>
              <div className="flex justify-between text-slate-800 font-bold pt-1 border-t border-black/[0.05]">
                <span>Сумма выплаты ({payoutTarget.school_percent}%):</span>
                <strong className="text-amber-600">{formatCurrency(payoutTarget.commission_amount)}</strong>
              </div>
            </div>

            <GlassInput
              label="Номер платежки / комментарий"
              placeholder="Например, чек СБП от 05.10 или номер транзакции..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
            />

            <div className="flex justify-end gap-2 pt-2">
              <GlassButton
                type="button"
                variant="ghost"
                onClick={() => setPayoutTarget(null)}
              >
                Отмена
              </GlassButton>
              <GlassButton
                type="submit"
                variant="primary"
                isLoading={isSubmitting}
                icon={<CheckCircle2 className="w-4 h-4" />}
              >
                Подтвердить выплату
              </GlassButton>
            </div>
          </form>
        </GlassModal>
      )}
    </div>
  );
};
