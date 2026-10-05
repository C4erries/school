import React, { useState, useEffect } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassInput } from '../../../shared/components/GlassInput';
import { GlassButton } from '../../../shared/components/GlassButton';
import { Client } from '../../../types/schedule';
import { PaymentMethod } from '../../../types/finance';
import { createPayment } from '../../../api/finance';
import { Check, Sparkles, CreditCard, Banknote, Smartphone, HelpCircle } from 'lucide-react';

interface AddPaymentModalProps {
  isOpen: boolean;
  onClose: () => void;
  clients: Client[];
  initialClientId?: string;
  onPaymentCreated: () => Promise<void>;
}

export const AddPaymentModal: React.FC<AddPaymentModalProps> = ({
  isOpen,
  onClose,
  clients,
  initialClientId,
  onPaymentCreated,
}) => {
  const [clientId, setClientId] = useState<string>('');
  const [format, setFormat] = useState<'individual' | 'pair' | 'group'>('individual');
  const [hours, setHours] = useState<string>('4');
  const [amount, setAmount] = useState<string>('');
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('transfer');
  const [paidAt, setPaidAt] = useState<string>('');
  const [notes, setNotes] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      const activeClients = clients.filter((c) => !c.is_archived);
      setClientId(initialClientId || (activeClients.length > 0 ? activeClients[0].id : ''));
      setFormat('individual');
      setHours('4');
      setPaymentMethod('transfer');
      setNotes('');
      setErrorMessage(null);
      const now = new Date();
      now.setMinutes(now.getMinutes() - now.getTimezoneOffset());
      setPaidAt(now.toISOString().slice(0, 16));
    }
  }, [isOpen, initialClientId, clients]);

  const selectedClient = clients.find((c) => c.id === clientId);
  const currentRate = selectedClient
    ? format === 'individual'
      ? selectedClient.rate_individual || selectedClient.base_rate || 0
      : format === 'pair'
      ? selectedClient.rate_pair || selectedClient.rate_individual || selectedClient.base_rate || 0
      : selectedClient.rate_group || selectedClient.rate_individual || selectedClient.base_rate || 0
    : 0;

  useEffect(() => {
    if (currentRate > 0 && hours) {
      setAmount(String(Math.round(Number(hours) * currentRate)));
    }
  }, [currentRate, hours]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!clientId) return setErrorMessage('Выберите ученика');
    const numHours = Number(hours);
    const numAmount = Number(amount);
    if (isNaN(numHours) || numHours <= 0) return setErrorMessage('Укажите количество часов');
    if (isNaN(numAmount) || numAmount < 0) return setErrorMessage('Укажите сумму оплаты');

    setIsSubmitting(true);
    setErrorMessage(null);
    try {
      await createPayment({
        client_id: clientId,
        hours: numHours,
        amount: numAmount,
        format,
        payment_method: paymentMethod,
        paid_at: paidAt ? new Date(paidAt).toISOString() : new Date().toISOString(),
        notes: notes.trim() || null,
      });
      await onPaymentCreated();
      onClose();
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Ошибка регистрации оплаты');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isOpen) return null;

  const paymentMethods: { id: PaymentMethod; label: string; icon: React.ReactNode }[] = [
    { id: 'transfer', label: 'СБП / Перевод', icon: <Smartphone className="w-3.5 h-3.5" /> },
    { id: 'card', label: 'Карта', icon: <CreditCard className="w-3.5 h-3.5" /> },
    { id: 'cash', label: 'Наличные', icon: <Banknote className="w-3.5 h-3.5" /> },
    { id: 'other', label: 'Другое', icon: <HelpCircle className="w-3.5 h-3.5" /> },
  ];

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title="Внести оплату"
      description="Фиксация поступления денег и автоматическое пополнение абонемента"
      maxWidth="md"
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        {errorMessage && (
          <div className="p-3 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-xs font-semibold text-rose-700">
            {errorMessage}
          </div>
        )}

        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
            Ученик
          </label>
          <select
            value={clientId}
            onChange={(e) => setClientId(e.target.value)}
            className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 bg-white/60 border border-black/[0.08] focus:outline-none focus:ring-2 focus:ring-indigo-400"
            required
          >
            <option value="" disabled>Выберите ученика</option>
            {clients.filter((c) => !c.is_archived).map((c) => (
              <option key={c.id} value={c.id}>
                {c.name} {c.phone ? `(${c.phone})` : ''} — баланс: {c.balances?.total_hours ?? c.balance ?? 0} ч
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
            Формат занятия
          </label>
          <div className="grid grid-cols-3 gap-2 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05]">
            {(['individual', 'pair', 'group'] as const).map((fmt) => (
              <button
                key={fmt}
                type="button"
                onClick={() => setFormat(fmt)}
                className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                  format === fmt ? 'bg-white text-indigo-700 shadow-sm' : 'text-slate-600 hover:text-slate-900'
                }`}
              >
                {fmt === 'individual' ? 'Индивидуально' : fmt === 'pair' ? 'В паре' : 'В группе'}
              </button>
            ))}
          </div>
        </div>

        <div>
          <div className="flex items-center justify-between mb-1.5 ml-1">
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500">
              Оплачиваемые часы
            </label>
            <div className="flex items-center gap-1">
              {['1', '4', '8', '12'].map((preset) => (
                <button
                  key={preset}
                  type="button"
                  onClick={() => setHours(preset)}
                  className={`px-2 py-0.5 rounded-lg text-xs font-semibold transition-all ${
                    hours === preset ? 'bg-indigo-600 text-white shadow-sm' : 'bg-black/[0.04] text-slate-600'
                  }`}
                >
                  +{preset}ч
                </button>
              ))}
            </div>
          </div>
          <GlassInput
            type="number"
            step="0.5"
            min="0.5"
            placeholder="Например, 4"
            value={hours}
            onChange={(e) => setHours(e.target.value)}
            required
          />
        </div>

        <div>
          <div className="flex items-center justify-between mb-1.5 ml-1">
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500">
              Сумма оплаты (₽)
            </label>
            {currentRate > 0 && (
              <span className="text-xs text-indigo-600 font-medium flex items-center gap-1">
                <Sparkles className="w-3 h-3" />
                Ставка: {currentRate} ₽/ч
              </span>
            )}
          </div>
          <GlassInput
            type="number"
            step="1"
            min="0"
            placeholder="Сумма в рублях"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            required
          />
        </div>

        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
            Способ оплаты
          </label>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
            {paymentMethods.map((m) => (
              <button
                key={m.id}
                type="button"
                onClick={() => setPaymentMethod(m.id)}
                className={`flex items-center justify-center gap-1.5 py-2.5 px-2 rounded-2xl text-xs font-semibold border transition-all ${
                  paymentMethod === m.id
                    ? 'bg-indigo-600 text-white border-indigo-600 shadow-md shadow-indigo-600/20'
                    : 'bg-white/50 text-slate-700 border-black/[0.08] hover:bg-white/80'
                }`}
              >
                {m.icon}
                <span>{m.label}</span>
              </button>
            ))}
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <GlassInput
            type="datetime-local"
            label="Дата и время"
            value={paidAt}
            onChange={(e) => setPaidAt(e.target.value)}
            required
          />
          <GlassInput
            label="Примечание / Чек"
            placeholder="Чек в WhatsApp, СБП..."
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
          />
        </div>

        <div className="flex justify-end gap-2 pt-2 border-t border-black/[0.04]">
          <GlassButton type="button" variant="ghost" onClick={onClose}>
            Отмена
          </GlassButton>
          <GlassButton
            type="submit"
            variant="primary"
            isLoading={isSubmitting}
            icon={<Check className="w-4 h-4" />}
          >
            Зафиксировать оплату
          </GlassButton>
        </div>
      </form>
    </GlassModal>
  );
};
