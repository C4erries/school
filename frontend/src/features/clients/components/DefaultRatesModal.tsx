import React, { useState, useEffect } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassInput } from '../../../shared/components/GlassInput';
import { GlassButton } from '../../../shared/components/GlassButton';
import { UserDefaultRates } from '../../../types/schedule';
import { updateDefaultRates } from '../../../api/schedule';
import { Coins, Sparkles, AlertCircle } from 'lucide-react';

interface DefaultRatesModalProps {
  isOpen: boolean;
  onClose: () => void;
  defaultRates?: UserDefaultRates | null;
  onSaved: (rates: UserDefaultRates) => void;
}

export const DefaultRatesModal: React.FC<DefaultRatesModalProps> = ({
  isOpen,
  onClose,
  defaultRates,
  onSaved,
}) => {
  const [indivRate, setIndivRate] = useState('1500');
  const [pairRate, setPairRate] = useState('1000');
  const [groupRate, setGroupRate] = useState('700');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      setError(null);
      if (defaultRates) {
        setIndivRate(String(defaultRates.rate_individual || 1500));
        setPairRate(String(defaultRates.rate_pair || 1000));
        setGroupRate(String(defaultRates.rate_group || 700));
      } else {
        setIndivRate('1500');
        setPairRate('1000');
        setGroupRate('700');
      }
    }
  }, [isOpen, defaultRates]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const ind = parseFloat(indivRate);
    const pair = parseFloat(pairRate);
    const grp = parseFloat(groupRate);

    if (isNaN(ind) || ind <= 0) {
      setError('Индивидуальная ставка должна быть больше нуля');
      return;
    }

    setIsSubmitting(true);
    setError(null);
    try {
      const payload: UserDefaultRates = {
        rate_individual: ind,
        rate_pair: isNaN(pair) ? 1000 : pair,
        rate_group: isNaN(grp) ? 700 : grp,
      };
      const updated = await updateDefaultRates(payload);
      onSaved(updated);
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Не удалось сохранить ставки');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title="Базовый прайс репетитора"
      description="Персональные ставки по умолчанию, которые будут автоматически подставляться при добавлении новых учеников."
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-800 text-xs flex items-center gap-2">
            <AlertCircle className="w-4 h-4 shrink-0 text-rose-600" />
            <span>{error}</span>
          </div>
        )}

        <div className="p-3.5 rounded-2xl bg-indigo-50/50 border border-indigo-100 flex items-start gap-3">
          <Sparkles className="w-4 h-4 text-indigo-600 shrink-0 mt-0.5" />
          <p className="text-xs text-indigo-950 leading-relaxed">
            Вы можете изменить ставки для конкретного ученика в любое время в его карточке.
          </p>
        </div>

        <div className="space-y-3">
          <GlassInput
            label={
              <span>
                Индивидуальное занятие (₽/час) <span className="text-rose-500">*</span>
              </span>
            }
            type="number"
            min="100"
            step="50"
            placeholder="1500"
            value={indivRate}
            onChange={(e) => setIndivRate(e.target.value)}
            required
            icon={<Coins className="w-4 h-4 text-indigo-500" />}
          />

          <GlassInput
            label="Занятие в паре (₽/час с каждого)"
            type="number"
            min="100"
            step="50"
            placeholder="1000"
            value={pairRate}
            onChange={(e) => setPairRate(e.target.value)}
            required
            icon={<Coins className="w-4 h-4 text-purple-500" />}
          />

          <GlassInput
            label="Занятие в группе (₽/час с каждого)"
            type="number"
            min="100"
            step="50"
            placeholder="700"
            value={groupRate}
            onChange={(e) => setGroupRate(e.target.value)}
            required
            icon={<Coins className="w-4 h-4 text-amber-500" />}
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
            Сохранить базовый прайс
          </GlassButton>
        </div>
      </form>
    </GlassModal>
  );
};
