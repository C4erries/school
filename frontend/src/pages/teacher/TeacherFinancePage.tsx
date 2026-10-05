import React, { useState, useEffect, useCallback } from 'react';
import { getFinanceSummary, getPayments, getPartnerSettlements } from '../../api/finance';
import { getClients } from '../../api/schedule';
import { FinanceSummary, Payment, PartnerSettlement } from '../../types/finance';
import { Client } from '../../types/schedule';
import { GlassButton } from '../../shared/components/GlassButton';
import { FinanceSummaryCards } from '../../features/finance/components/FinanceSummaryCards';
import { DebtsAndBalancesTable } from '../../features/finance/components/DebtsAndBalancesTable';
import { PaymentsLedgerTable } from '../../features/finance/components/PaymentsLedgerTable';
import { PartnerSettlementsCard } from '../../features/finance/components/PartnerSettlementsCard';
import { AddPaymentModal } from '../../features/finance/components/AddPaymentModal';
import { ExportDataModal } from '../../features/finance/components/ExportDataModal';
import {
  Download,
  Plus,
  RefreshCw,
  AlertCircle,
  AlertTriangle,
  Receipt,
  Building2,
} from 'lucide-react';

export const TeacherFinancePage: React.FC = () => {
  // Текущий месяц по умолчанию в формате YYYY-MM
  const getCurrentMonthString = () => {
    const now = new Date();
    const y = now.getFullYear();
    const m = String(now.getMonth() + 1).padStart(2, '0');
    return `${y}-${m}`;
  };

  const [currentMonth, setCurrentMonth] = useState<string>(getCurrentMonthString());
  const [activeTab, setActiveTab] = useState<'debts' | 'ledger' | 'partners'>('debts');

  // Данные
  const [summary, setSummary] = useState<FinanceSummary | null>(null);
  const [payments, setPayments] = useState<Payment[]>([]);
  const [settlements, setSettlements] = useState<PartnerSettlement[]>([]);
  const [clients, setClients] = useState<Client[]>([]);

  // Состояния загрузки и ошибок
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Состояния модалок
  const [isAddPaymentOpen, setIsAddPaymentOpen] = useState(false);
  const [paymentClientId, setPaymentClientId] = useState<string | undefined>(undefined);
  const [isExportOpen, setIsExportOpen] = useState(false);

  const loadData = useCallback(async () => {
    setIsLoading(true);
    setErrorMessage(null);
    try {
      const [sumData, payData, setlData, clientData] = await Promise.all([
        getFinanceSummary(currentMonth).catch(() => null),
        getPayments().catch(() => []),
        getPartnerSettlements(currentMonth).catch(() => []),
        getClients().catch(() => []),
      ]);

      setSummary(sumData);
      setPayments(payData);
      setSettlements(setlData);
      setClients(clientData);
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Ошибка загрузки финансовых данных');
    } finally {
      setIsLoading(false);
    }
  }, [currentMonth]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleAcceptPayment = (client: Client) => {
    setPaymentClientId(client.id);
    setIsAddPaymentOpen(true);
  };

  const handleOpenAddPayment = () => {
    setPaymentClientId(undefined);
    setIsAddPaymentOpen(true);
  };

  const debtorsCount = summary?.debtors_count ?? 0;
  const unpaidSchoolsCount = settlements.filter((s) => !s.is_paid).length;

  return (
    <div className="space-y-6">
      {/* Шапка страницы */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">
            Бухгалтерия и финансы
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Журнал поступивших оплат, контроль задолженностей и взаиморасчеты с партнерскими школами
          </p>
        </div>

        <div className="flex items-center gap-2.5 flex-wrap">
          <GlassButton
            variant="secondary"
            onClick={() => setIsExportOpen(true)}
            icon={<Download className="w-4 h-4 text-slate-600" />}
          >
            Экспорт данных
          </GlassButton>

          <GlassButton
            variant="primary"
            onClick={handleOpenAddPayment}
            icon={<Plus className="w-4 h-4" />}
          >
            Внести оплату
          </GlassButton>
        </div>
      </div>

      {/* Ошибка подключения / загрузки */}
      {errorMessage && (
        <div className="p-4 rounded-3xl bg-rose-500/10 border border-rose-500/20 backdrop-blur-md flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
            <p className="text-sm font-medium text-rose-900">{errorMessage}</p>
          </div>
          <GlassButton
            variant="secondary"
            size="sm"
            onClick={loadData}
            icon={<RefreshCw className="w-3.5 h-3.5" />}
          >
            Повторить
          </GlassButton>
        </div>
      )}

      {/* Сводные карточки метрик с селектором месяца */}
      <FinanceSummaryCards
        summary={summary}
        currentMonth={currentMonth}
        onMonthChange={setCurrentMonth}
        isLoading={isLoading}
      />

      {/* Переключатель вкладок раздела */}
      <div className="flex items-center gap-1.5 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05] overflow-x-auto">
        <button
          type="button"
          onClick={() => setActiveTab('debts')}
          className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-semibold transition-all whitespace-nowrap ${
            activeTab === 'debts'
              ? 'bg-white text-indigo-600 shadow-sm'
              : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
          }`}
        >
          <AlertTriangle className="w-3.5 h-3.5 text-rose-500" />
          <span>Долги и балансы</span>
          {debtorsCount > 0 && (
            <span className="px-1.5 py-0.5 rounded-full text-[10px] bg-rose-500/15 text-rose-800">
              {debtorsCount}
            </span>
          )}
        </button>

        <button
          type="button"
          onClick={() => setActiveTab('ledger')}
          className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-semibold transition-all whitespace-nowrap ${
            activeTab === 'ledger'
              ? 'bg-white text-indigo-600 shadow-sm'
              : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
          }`}
        >
          <Receipt className="w-3.5 h-3.5 text-indigo-500" />
          <span>Журнал оплат</span>
          <span className="px-1.5 py-0.5 rounded-full text-[10px] bg-indigo-500/15 text-indigo-800">
            {payments.length}
          </span>
        </button>

        <button
          type="button"
          onClick={() => setActiveTab('partners')}
          className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-semibold transition-all whitespace-nowrap ${
            activeTab === 'partners'
              ? 'bg-white text-indigo-600 shadow-sm'
              : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
          }`}
        >
          <Building2 className="w-3.5 h-3.5 text-amber-500" />
          <span>Школы и партнеры</span>
          {unpaidSchoolsCount > 0 && (
            <span className="px-1.5 py-0.5 rounded-full text-[10px] bg-amber-500/15 text-amber-800">
              {unpaidSchoolsCount}
            </span>
          )}
        </button>
      </div>

      {/* Содержимое активной вкладки */}
      {activeTab === 'debts' && (
        <DebtsAndBalancesTable
          clients={clients}
          onAcceptPayment={handleAcceptPayment}
          isLoading={isLoading}
        />
      )}

      {activeTab === 'ledger' && (
        <PaymentsLedgerTable
          payments={payments}
          clients={clients}
          onOpenAddPayment={handleOpenAddPayment}
          isLoading={isLoading}
        />
      )}

      {activeTab === 'partners' && (
        <PartnerSettlementsCard
          settlements={settlements}
          month={currentMonth}
          onPayoutSuccess={loadData}
          isLoading={isLoading}
        />
      )}

      {/* Модалка внесения оплаты */}
      <AddPaymentModal
        isOpen={isAddPaymentOpen}
        onClose={() => setIsAddPaymentOpen(false)}
        clients={clients}
        initialClientId={paymentClientId}
        onPaymentCreated={loadData}
      />

      {/* Модалка экспорта */}
      <ExportDataModal
        isOpen={isExportOpen}
        onClose={() => setIsExportOpen(false)}
      />
    </div>
  );
};
