import React, { useState, useEffect } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassInput } from '../../../shared/components/GlassInput';
import { GlassButton } from '../../../shared/components/GlassButton';
import { Client, SubscriptionFormat } from '../../../types/schedule';
import { addSubscription } from '../../../api/schedule';
import { Clock, Sparkles } from 'lucide-react';

interface AddSubscriptionModalProps {
  isOpen: boolean;
  onClose: () => void;
  client: Client | null;
  onAdded: () => Promise<void>;
}

export const AddSubscriptionModal: React.FC<AddSubscriptionModalProps> = ({
  isOpen,
  onClose,
  client,
  onAdded,
}) => {
  const [subFormat, setSubFormat] = useState<SubscriptionFormat>('individual');
  const [subHours, setSubHours] = useState('8');
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setSubFormat('individual');
      setSubHours('8');
    }
  }, [isOpen]);

  const handleAddSubscription = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!client || !subHours) return;

    setIsSubmitting(true);
    try {
      await addSubscription({
        client_id: client.id,
        format: subFormat,
        hours: Number(subHours),
      });
      await onAdded();
      onClose();
    } catch (error: unknown) {
      alert(error instanceof Error ? error.message : 'Ошибка пополнения абонемента');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!client) return null;

  const currentRate =
    subFormat === 'individual'
      ? client.rate_individual
      : subFormat === 'pair'
      ? client.rate_pair || client.rate_individual
      : client.rate_group || client.rate_individual;

  const estimatedTotal = Number(subHours || 0) * currentRate;

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title="Пополнить абонемент"
      description={`Начисление оплаченных часов для ${client.name}`}
    >
      <form onSubmit={handleAddSubscription} className="space-y-4">
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-2 ml-1">
            Формат занятия
          </label>
          <div className="grid grid-cols-3 gap-2 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05]">
            <button
              type="button"
              onClick={() => setSubFormat('individual')}
              className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                subFormat === 'individual'
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Индив.
            </button>
            <button
              type="button"
              onClick={() => setSubFormat('pair')}
              className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                subFormat === 'pair'
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              В паре
            </button>
            <button
              type="button"
              onClick={() => setSubFormat('group')}
              className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                subFormat === 'group'
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Группа
            </button>
          </div>
        </div>

        <GlassInput
          label="Количество часов"
          type="number"
          step="0.5"
          min="0.5"
          value={subHours}
          onChange={(e) => setSubHours(e.target.value)}
          required
          icon={<Clock className="w-4 h-4" />}
        />

        {/* Быстрые пресеты часов */}
        <div className="flex items-center gap-2 pt-1">
          <span className="text-xs text-slate-400 font-medium">Пресеты:</span>
          {['4', '8', '12', '16'].map((preset) => (
            <button
              key={preset}
              type="button"
              onClick={() => setSubHours(preset)}
              className={`px-2.5 py-1 rounded-lg text-xs font-semibold transition-all ${
                subHours === preset
                  ? 'bg-indigo-600 text-white shadow-sm'
                  : 'bg-white/40 text-slate-700 hover:bg-white/70 border border-white/60'
              }`}
            >
              {preset} ч
            </button>
          ))}
        </div>

        {/* Расчет суммы к оплате */}
        <div className="p-3.5 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 flex items-center justify-between text-xs shadow-sm">
          <span className="text-slate-500 font-medium flex items-center gap-1.5">
            <Sparkles className="w-3.5 h-3.5 text-indigo-500" />
            Ориентировочная сумма:
          </span>
          <span className="font-bold text-slate-900 text-sm">
            {estimatedTotal.toLocaleString('ru-RU')} ₽
          </span>
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
            variant="mint"
            isLoading={isSubmitting}
          >
            Начислить часы
          </GlassButton>
        </div>
      </form>
    </GlassModal>
  );
};
