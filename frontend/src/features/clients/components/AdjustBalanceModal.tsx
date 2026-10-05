import React, { useState, useEffect } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassInput } from '../../../shared/components/GlassInput';
import { GlassButton } from '../../../shared/components/GlassButton';
import { Client, SubscriptionFormat } from '../../../types/schedule';
import { adjustClientBalance } from '../../../api/schedule';
import { Scale, ArrowRight, AlertCircle } from 'lucide-react';

interface AdjustBalanceModalProps {
  isOpen: boolean;
  onClose: () => void;
  client: Client | null;
  onAdjusted: () => Promise<void>;
}

export const AdjustBalanceModal: React.FC<AdjustBalanceModalProps> = ({
  isOpen,
  onClose,
  client,
  onAdjusted,
}) => {
  const [format, setFormat] = useState<SubscriptionFormat>('individual');
  const [deltaHours, setDeltaHours] = useState('');
  const [reason, setReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      setFormat('individual');
      setDeltaHours('');
      setReason('');
      setError(null);
    }
  }, [isOpen]);

  if (!client) return null;

  const currentHours =
    format === 'individual'
      ? client.balances?.individual_hours ?? client.balance ?? 0
      : format === 'pair'
      ? client.balances?.pair_hours ?? 0
      : client.balances?.group_hours ?? 0;

  const numericDelta = parseFloat(deltaHours.replace(',', '.')) || 0;
  const newHours = currentHours + numericDelta;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmedReason = reason.trim();
    if (!trimmedReason) {
      setError('Укажите обязательную причину корректировки для аудит-лога');
      return;
    }
    if (numericDelta === 0) {
      setError('Изменение часов не может быть нулевым');
      return;
    }

    setIsSubmitting(true);
    setError(null);
    try {
      await adjustClientBalance(client.id, {
        format,
        delta_hours: numericDelta,
        reason: trimmedReason,
      });
      await onAdjusted();
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Не удалось скорректировать баланс');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title="Ручная корректировка баланса"
      description={`Ручное начисление или списание часов для ${client.name} с фиксацией в аудит-логе.`}
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-800 text-xs flex items-center gap-2">
            <AlertCircle className="w-4 h-4 shrink-0 text-rose-600" />
            <span>{error}</span>
          </div>
        )}

        {/* Выбор формата */}
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-2 ml-1">
            Формат занятия
          </label>
          <div className="grid grid-cols-3 gap-2 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05]">
            <button
              type="button"
              onClick={() => setFormat('individual')}
              className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                format === 'individual'
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Индив.
            </button>
            <button
              type="button"
              onClick={() => setFormat('pair')}
              className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                format === 'pair'
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              В паре
            </button>
            <button
              type="button"
              onClick={() => setFormat('group')}
              className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                format === 'group'
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Группа
            </button>
          </div>
        </div>

        {/* Поле изменения часов */}
        <div>
          <GlassInput
            label={
              <span>
                Изменение часов (<span className="text-indigo-600 font-mono">+1.5</span> или{' '}
                <span className="text-rose-600 font-mono">-1.0</span>) *
              </span>
            }
            type="text"
            inputMode="decimal"
            placeholder="например: +1.5, 1.5 или -1.0"
            value={deltaHours}
            onChange={(e) => setDeltaHours(e.target.value)}
            required
            icon={<Scale className="w-4 h-4" />}
          />

          {/* Быстрые пресеты */}
          <div className="flex items-center gap-1.5 pt-1.5 flex-wrap">
            <span className="text-[11px] text-slate-400 font-medium">Быстро:</span>
            {[
              { val: '+1', label: '+1 ч' },
              { val: '+1.5', label: '+1.5 ч' },
              { val: '+2', label: '+2 ч' },
              { val: '-1', label: '-1 ч' },
              { val: '-1.5', label: '-1.5 ч' },
            ].map((p) => {
              const isSelected =
                deltaHours === p.val ||
                (p.val.startsWith('+') && deltaHours === p.val.slice(1));
              return (
                <button
                  key={p.val}
                  type="button"
                  onClick={() => setDeltaHours(p.val)}
                  className={`px-2 py-0.5 rounded-lg text-xs font-mono font-medium transition-all ${
                    isSelected
                      ? 'bg-indigo-600 text-white shadow-sm'
                      : 'bg-white/40 text-slate-700 hover:bg-white/80 border border-white/60'
                  }`}
                >
                  {p.label}
                </button>
              );
            })}
          </div>
        </div>

        {/* Превью пересчета */}
        <div className="p-3.5 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 flex items-center justify-between text-xs shadow-sm">
          <div>
            <span className="text-slate-500 block text-[10px] uppercase font-bold tracking-wider">
              Текущий баланс
            </span>
            <span className="font-bold text-slate-800 text-sm">{currentHours} ч</span>
          </div>

          <div className="flex items-center gap-1.5 text-slate-400">
            <span
              className={`font-mono font-bold text-xs ${
                numericDelta > 0
                  ? 'text-emerald-600'
                  : numericDelta < 0
                  ? 'text-rose-600'
                  : 'text-slate-400'
              }`}
            >
              {numericDelta > 0 ? `+${numericDelta}` : numericDelta} ч
            </span>
            <ArrowRight className="w-4 h-4" />
          </div>

          <div className="text-right">
            <span className="text-slate-500 block text-[10px] uppercase font-bold tracking-wider">
              Итоговый баланс
            </span>
            <span
              className={`font-bold text-sm ${
                newHours < 0 ? 'text-rose-600' : 'text-slate-900'
              }`}
            >
              {newHours} ч
            </span>
          </div>
        </div>

        {/* Обязательная причина / комментарий */}
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-700 mb-1.5 ml-1">
            Причина корректировки * <span className="text-slate-400 font-normal normal-case">(аудит-лог)</span>
          </label>
          <textarea
            rows={2}
            placeholder="например: Компенсация за сбой связи / откат ошибочного списания"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            required
            className="w-full rounded-2xl px-4 py-2.5 text-sm text-slate-800 bg-white/70 border border-slate-200 focus:outline-none focus:border-indigo-500 placeholder:text-slate-400"
          />
        </div>

        <div className="pt-3 flex justify-end gap-3 border-t border-black/[0.05]">
          <GlassButton
            type="button"
            variant="secondary"
            onClick={onClose}
          >
            Отмена
          </GlassButton>
          <GlassButton
            type="submit"
            variant="primary"
            isLoading={isSubmitting}
          >
            Применить корректировку
          </GlassButton>
        </div>
      </form>
    </GlassModal>
  );
};
