import React, { useState, useMemo } from 'react';
import { Payment, PaymentMethod } from '../../../types/finance';
import { Client } from '../../../types/schedule';
import { GlassCard } from '../../../shared/components/GlassCard';
import { GlassButton } from '../../../shared/components/GlassButton';
import { GlassInput } from '../../../shared/components/GlassInput';
import { Badge } from '../../../shared/components/Badge';
import {
  Plus,
  Search,
  CreditCard,
  Banknote,
  Smartphone,
  HelpCircle,
  Receipt,
  Clock,
  User,
} from 'lucide-react';

interface PaymentsLedgerTableProps {
  payments: Payment[];
  clients: Client[];
  onOpenAddPayment: () => void;
  isLoading?: boolean;
}

export const PaymentsLedgerTable: React.FC<PaymentsLedgerTableProps> = ({
  payments,
  clients,
  onOpenAddPayment,
  isLoading = false,
}) => {
  const [search, setSearch] = useState('');
  const [selectedClientId, setSelectedClientId] = useState<string>('all');
  const [selectedMethod, setSelectedMethod] = useState<string>('all');

  const filteredPayments = useMemo(() => {
    return payments.filter((p) => {
      if (selectedClientId !== 'all' && p.client_id !== selectedClientId) {
        return false;
      }
      if (selectedMethod !== 'all' && p.payment_method !== selectedMethod) {
        return false;
      }
      if (search.trim()) {
        const q = search.toLowerCase();
        const clientName = p.client_name?.toLowerCase() || '';
        const notes = p.notes?.toLowerCase() || '';
        if (!clientName.includes(q) && !notes.includes(q)) {
          return false;
        }
      }
      return true;
    });
  }, [payments, selectedClientId, selectedMethod, search]);

  const totalFilteredAmount = useMemo(() => {
    return filteredPayments.reduce((acc, p) => acc + (p.amount || 0), 0);
  }, [filteredPayments]);

  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      maximumFractionDigits: 0,
    }).format(val);
  };

  const formatDate = (dateStr: string) => {
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch {
      return dateStr;
    }
  };

  const getFormatBadge = (fmt: string) => {
    switch (fmt) {
      case 'individual':
        return <Badge variant="indigo">Индивидуально</Badge>;
      case 'pair':
        return <Badge variant="amber">В паре</Badge>;
      case 'group':
        return <Badge variant="mint">В группе</Badge>;
      default:
        return <Badge variant="neutral">{fmt}</Badge>;
    }
  };

  const getMethodBadge = (method: PaymentMethod) => {
    switch (method) {
      case 'transfer':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-xl text-xs font-semibold bg-blue-500/10 text-blue-700 border border-blue-400/20">
            <Smartphone className="w-3 h-3 text-blue-500" />
            СБП / Перевод
          </span>
        );
      case 'cash':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-xl text-xs font-semibold bg-emerald-500/10 text-emerald-700 border border-emerald-400/20">
            <Banknote className="w-3 h-3 text-emerald-500" />
            Наличные
          </span>
        );
      case 'card':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-xl text-xs font-semibold bg-purple-500/10 text-purple-700 border border-purple-400/20">
            <CreditCard className="w-3 h-3 text-purple-500" />
            Карта
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-xl text-xs font-semibold bg-slate-500/10 text-slate-700 border border-slate-300/40">
            <HelpCircle className="w-3 h-3 text-slate-500" />
            Другое
          </span>
        );
    }
  };

  return (
    <div className="space-y-4">
      {/* Верхняя панель: Заголовок, Сумма и кнопка создания платежа */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-2xl bg-indigo-500/15 text-indigo-600 flex items-center justify-center border border-indigo-400/20">
            <Receipt className="w-5 h-5" />
          </div>
          <div>
            <h2 className="font-bold text-lg text-slate-900">Журнал поступивших платежей</h2>
            <p className="text-xs text-slate-500">
              Найдено: {filteredPayments.length} платежей на сумму{' '}
              <strong className="text-emerald-600 font-bold">{formatCurrency(totalFilteredAmount)}</strong>
            </p>
          </div>
        </div>

        <GlassButton
          variant="primary"
          onClick={onOpenAddPayment}
          icon={<Plus className="w-4 h-4" />}
        >
          Внести оплату
        </GlassButton>
      </div>

      {/* Панель фильтров */}
      <GlassCard className="p-4">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {/* Поиск */}
          <GlassInput
            icon={<Search className="w-4 h-4" />}
            placeholder="Поиск по ученику или примечанию..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />

          {/* Фильтр по ученику */}
          <div>
            <select
              value={selectedClientId}
              onChange={(e) => setSelectedClientId(e.target.value)}
              className="w-full rounded-2xl px-3 py-2.5 text-xs text-slate-800 bg-white/50 border border-black/[0.08] focus:outline-none focus:ring-1 focus:ring-indigo-400"
            >
              <option value="all">Все ученики ({clients.length})</option>
              {clients.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>

          {/* Фильтр по способу оплаты */}
          <div>
            <select
              value={selectedMethod}
              onChange={(e) => setSelectedMethod(e.target.value)}
              className="w-full rounded-2xl px-3 py-2.5 text-xs text-slate-800 bg-white/50 border border-black/[0.08] focus:outline-none focus:ring-1 focus:ring-indigo-400"
            >
              <option value="all">Все способы оплаты</option>
              <option value="transfer">СБП / Перевод на карту</option>
              <option value="cash">Наличные</option>
              <option value="card">Банковская карта</option>
              <option value="other">Другое</option>
            </select>
          </div>
        </div>
      </GlassCard>

      {/* Таблица / Список */}
      {isLoading ? (
        <GlassCard className="p-8 text-center text-slate-400">
          <div className="w-6 h-6 border-2 border-indigo-600 border-t-transparent rounded-full animate-spin mx-auto mb-2" />
          <span>Загрузка платежей...</span>
        </GlassCard>
      ) : filteredPayments.length === 0 ? (
        <GlassCard className="p-10 text-center text-slate-500 space-y-2">
          <Receipt className="w-8 h-8 text-slate-300 mx-auto" />
          <p className="font-semibold text-sm">Платежи не найдены</p>
          <p className="text-xs text-slate-400">
            Зафиксируйте первый платеж кнопкой «Внести оплату»
          </p>
        </GlassCard>
      ) : (
        <div className="overflow-x-auto rounded-3xl border border-black/[0.05] liquid-glass">
          <table className="w-full text-left border-collapse text-xs">
            <thead>
              <tr className="border-b border-black/[0.06] bg-black/[0.02] text-slate-500 font-semibold uppercase tracking-wider">
                <th className="py-3 px-4">Дата</th>
                <th className="py-3 px-4">Ученик</th>
                <th className="py-3 px-4">Сумма</th>
                <th className="py-3 px-4">Часы</th>
                <th className="py-3 px-4">Формат</th>
                <th className="py-3 px-4">Способ</th>
                <th className="py-3 px-4">Примечание</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-black/[0.04]">
              {filteredPayments.map((p) => (
                <tr key={p.id} className="hover:bg-white/40 transition-colors">
                  <td className="py-3 px-4 whitespace-nowrap text-slate-600 font-medium">
                    <div className="flex items-center gap-1.5">
                      <Clock className="w-3.5 h-3.5 text-slate-400" />
                      <span>{formatDate(p.paid_at)}</span>
                    </div>
                  </td>
                  <td className="py-3 px-4 whitespace-nowrap font-bold text-slate-800">
                    <div className="flex items-center gap-1.5">
                      <User className="w-3.5 h-3.5 text-slate-400" />
                      <span>{p.client_name || 'Ученик'}</span>
                    </div>
                  </td>
                  <td className="py-3 px-4 whitespace-nowrap font-bold text-emerald-600 text-sm">
                    +{formatCurrency(p.amount)}
                  </td>
                  <td className="py-3 px-4 whitespace-nowrap font-semibold text-slate-700">
                    +{p.hours} ч
                  </td>
                  <td className="py-3 px-4 whitespace-nowrap">
                    {getFormatBadge(p.format)}
                  </td>
                  <td className="py-3 px-4 whitespace-nowrap">
                    {getMethodBadge(p.payment_method)}
                  </td>
                  <td className="py-3 px-4 text-slate-500 max-w-xs truncate">
                    {p.notes || '—'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
