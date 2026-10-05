import React, { useState, useMemo } from 'react';
import { Client } from '../../../types/schedule';
import { GlassCard } from '../../../shared/components/GlassCard';
import { GlassButton } from '../../../shared/components/GlassButton';
import { GlassInput } from '../../../shared/components/GlassInput';
import { Badge } from '../../../shared/components/Badge';
import { Search, CreditCard, AlertTriangle, Phone, Clock, Sparkles, Users } from 'lucide-react';

interface DebtsAndBalancesTableProps {
  clients: Client[];
  onAcceptPayment: (client: Client) => void;
  isLoading?: boolean;
}

export const DebtsAndBalancesTable: React.FC<DebtsAndBalancesTableProps> = ({
  clients,
  onAcceptPayment,
  isLoading = false,
}) => {
  const [filterMode, setFilterMode] = useState<'debtors' | 'low' | 'all'>('debtors');
  const [search, setSearch] = useState('');

  const calculateDebtAmount = (client: Client): number => {
    let debt = 0;
    const indivRate = client.rate_individual || client.base_rate || 0;
    const pairRate = client.rate_pair || indivRate;
    const groupRate = client.rate_group || indivRate;

    if (client.balances) {
      if (client.balances.individual_hours < 0) debt += Math.abs(client.balances.individual_hours) * indivRate;
      if (client.balances.pair_hours < 0) debt += Math.abs(client.balances.pair_hours) * pairRate;
      if (client.balances.group_hours < 0) debt += Math.abs(client.balances.group_hours) * groupRate;
    } else if (client.balance < 0) {
      debt = Math.abs(client.balance) * indivRate;
    }
    return debt;
  };

  const getFormatRate = (client: Client, format: 'individual' | 'pair' | 'group') => {
    if (format === 'individual') return client.rate_individual || client.base_rate || 0;
    if (format === 'pair') return client.rate_pair || client.rate_individual || client.base_rate || 0;
    return client.rate_group || client.rate_individual || client.base_rate || 0;
  };

  const isClientDebtor = (c: Client) => {
    const total = c.balances?.total_hours ?? c.balance ?? 0;
    return total < 0 || (c.balances?.individual_hours ?? 0) < 0 || (c.balances?.pair_hours ?? 0) < 0 || (c.balances?.group_hours ?? 0) < 0;
  };

  const activeClients = useMemo(() => clients.filter((c) => !c.is_archived), [clients]);
  const debtorCount = useMemo(() => activeClients.filter(isClientDebtor).length, [activeClients]);
  const lowBalanceCount = useMemo(() => activeClients.filter((c) => {
    const total = c.balances?.total_hours ?? c.balance ?? 0;
    return total >= 0 && total <= 1;
  }).length, [activeClients]);

  const filteredClients = useMemo(() => {
    return activeClients.filter((c) => {
      if (search.trim()) {
        const q = search.toLowerCase();
        if (!c.name.toLowerCase().includes(q) && !(c.phone?.toLowerCase().includes(q))) return false;
      }
      const total = c.balances?.total_hours ?? c.balance ?? 0;
      if (filterMode === 'debtors') return isClientDebtor(c);
      if (filterMode === 'low') return total >= 0 && total <= 1;
      return true;
    });
  }, [activeClients, search, filterMode]);

  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(val);
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
        <div className="flex items-center gap-1.5 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05] overflow-x-auto">
          <button
            type="button"
            onClick={() => setFilterMode('debtors')}
            className={`flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all whitespace-nowrap ${
              filterMode === 'debtors' ? 'bg-white text-rose-700 shadow-sm' : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <AlertTriangle className="w-3.5 h-3.5 text-rose-500" />
            <span>Должники</span>
            <span className="ml-1 px-1.5 py-0.5 rounded-full text-[10px] bg-rose-500/15 text-rose-800">{debtorCount}</span>
          </button>
          <button
            type="button"
            onClick={() => setFilterMode('low')}
            className={`flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all whitespace-nowrap ${
              filterMode === 'low' ? 'bg-white text-amber-700 shadow-sm' : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <Clock className="w-3.5 h-3.5 text-amber-500" />
            <span>Мало часов (≤1 ч)</span>
            <span className="ml-1 px-1.5 py-0.5 rounded-full text-[10px] bg-amber-500/15 text-amber-800">{lowBalanceCount}</span>
          </button>
          <button
            type="button"
            onClick={() => setFilterMode('all')}
            className={`flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all whitespace-nowrap ${
              filterMode === 'all' ? 'bg-white text-indigo-700 shadow-sm' : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <Users className="w-3.5 h-3.5 text-slate-500" />
            <span>Все ученики</span>
            <span className="ml-1 px-1.5 py-0.5 rounded-full text-[10px] bg-slate-500/15 text-slate-800">{activeClients.length}</span>
          </button>
        </div>
        <div className="w-full sm:w-64">
          <GlassInput
            icon={<Search className="w-4 h-4" />}
            placeholder="Поиск по имени или телефону..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
      </div>

      {isLoading ? (
        <GlassCard className="p-8 text-center text-slate-400">
          <div className="w-6 h-6 border-2 border-indigo-600 border-t-transparent rounded-full animate-spin mx-auto mb-2" />
          <span>Загрузка балансов...</span>
        </GlassCard>
      ) : filteredClients.length === 0 ? (
        <GlassCard className="p-10 text-center text-slate-500 space-y-2">
          <Sparkles className="w-8 h-8 text-slate-300 mx-auto" />
          <p className="font-semibold text-sm">
            {filterMode === 'debtors' ? 'Отлично! Учеников с задолженностями нет' : 'Ученики не найдены'}
          </p>
        </GlassCard>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredClients.map((client) => {
            const totalHours = client.balances?.total_hours ?? client.balance ?? 0;
            const debt = calculateDebtAmount(client);
            const isNegative = totalHours < 0 || debt > 0;

            return (
              <GlassCard key={client.id} className="p-5 flex flex-col justify-between space-y-4 hover:shadow-md transition-shadow">
                <div>
                  <div className="flex items-start justify-between gap-2 mb-2">
                    <div className="min-w-0">
                      <h3 className="font-bold text-base text-slate-900 truncate">{client.name}</h3>
                      {client.phone && (
                        <p className="text-xs text-slate-500 flex items-center gap-1 mt-0.5">
                          <Phone className="w-3 h-3 text-slate-400 shrink-0" />
                          <span>{client.phone}</span>
                        </p>
                      )}
                    </div>
                    {isNegative ? (
                      <Badge variant="danger" className="shrink-0">Долг {formatCurrency(debt)}</Badge>
                    ) : totalHours <= 1 ? (
                      <Badge variant="amber" className="shrink-0">Осталось {totalHours} ч</Badge>
                    ) : (
                      <Badge variant="mint" className="shrink-0">Баланс {totalHours} ч</Badge>
                    )}
                  </div>

                  {client.tags && client.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1 mb-2">
                      {client.tags.map((t) => (
                        <span key={t.id} className="px-2 py-0.5 rounded-lg text-[10px] font-semibold bg-indigo-50 text-indigo-700 border border-indigo-200">
                          {t.name} ({t.school_percent}%)
                        </span>
                      ))}
                    </div>
                  )}

                  <div className="grid grid-cols-3 gap-2 p-2.5 rounded-2xl bg-black/[0.03] border border-black/[0.04] text-center my-3">
                    <div>
                      <div className="text-[10px] text-slate-500 font-semibold uppercase">Индивид.</div>
                      <div className={`text-xs font-bold ${(client.balances?.individual_hours ?? 0) < 0 ? 'text-rose-600' : 'text-slate-800'}`}>
                        {client.balances?.individual_hours ?? client.balance ?? 0} ч
                      </div>
                      <div className="text-[10px] text-slate-400">{getFormatRate(client, 'individual')} ₽</div>
                    </div>
                    <div>
                      <div className="text-[10px] text-slate-500 font-semibold uppercase">В паре</div>
                      <div className={`text-xs font-bold ${(client.balances?.pair_hours ?? 0) < 0 ? 'text-rose-600' : 'text-slate-800'}`}>
                        {client.balances?.pair_hours ?? 0} ч
                      </div>
                      <div className="text-[10px] text-slate-400">{getFormatRate(client, 'pair')} ₽</div>
                    </div>
                    <div>
                      <div className="text-[10px] text-slate-500 font-semibold uppercase">В группе</div>
                      <div className={`text-xs font-bold ${(client.balances?.group_hours ?? 0) < 0 ? 'text-rose-600' : 'text-slate-800'}`}>
                        {client.balances?.group_hours ?? 0} ч
                      </div>
                      <div className="text-[10px] text-slate-400">{getFormatRate(client, 'group')} ₽</div>
                    </div>
                  </div>
                </div>

                <div className="pt-2 border-t border-black/[0.05] flex items-center justify-between gap-2">
                  <span className="text-xs text-slate-500">
                    {isNegative ? 'Требуется пополнение' : 'Пополнить абонемент'}
                  </span>
                  <GlassButton
                    variant={isNegative ? 'coral' : 'secondary'}
                    size="sm"
                    onClick={() => onAcceptPayment(client)}
                    icon={<CreditCard className="w-3.5 h-3.5" />}
                  >
                    Принять оплату
                  </GlassButton>
                </div>
              </GlassCard>
            );
          })}
        </div>
      )}
    </div>
  );
};
