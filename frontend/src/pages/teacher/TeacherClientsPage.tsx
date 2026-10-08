import React, { useEffect, useState, useCallback } from 'react';
import { UserPlus, Coins, RefreshCw, AlertCircle } from 'lucide-react';
import { Client, Tag, UserDefaultRates } from '../../types/schedule';
import { getClients, getTags, getDefaultRates, archiveClient, unarchiveClient } from '../../api/schedule';
import { GlassButton } from '../../shared/components/GlassButton';
import { GlassCard } from '../../shared/components/GlassCard';
import { ClientCard } from '../../features/clients/components/ClientCard';
import { ClientFilters } from '../../features/clients/components/ClientFilters';
import { useClientFilters } from '../../features/clients/hooks/useClientFilters';
import { CreateClientModal } from '../../features/clients/components/CreateClientModal';
import { EditClientModal } from '../../features/clients/components/EditClientModal';
import { AddSubscriptionModal } from '../../features/clients/components/AddSubscriptionModal';
import { AdjustBalanceModal } from '../../features/clients/components/AdjustBalanceModal';
import { DefaultRatesModal } from '../../features/clients/components/DefaultRatesModal';
import { ClientJournalModal } from '../../features/clients/components/ClientJournalModal';

export const TeacherClientsPage: React.FC = () => {
  const [clients, setClients] = useState<Client[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [defaultRates, setDefaultRates] = useState<UserDefaultRates | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isRatesOpen, setIsRatesOpen] = useState(false);
  const [editingClient, setEditingClient] = useState<Client | null>(null);
  const [subClient, setSubClient] = useState<Client | null>(null);
  const [adjustClient, setAdjustClient] = useState<Client | null>(null);
  const [journalClient, setJournalClient] = useState<Client | null>(null);

  const {
    searchQuery, setSearchQuery, statusTab, setStatusTab,
    balanceFilter, setBalanceFilter, selectedTagId, setSelectedTagId,
    counts, filteredClients,
  } = useClientFilters(clients);

  const loadData = useCallback(async () => {
    setIsLoading(true);
    setErrorMessage(null);
    try {
      const [clientsData, tagsData, ratesData] = await Promise.all([
        getClients(),
        getTags().catch(() => []),
        getDefaultRates().catch(() => null),
      ]);
      setClients(clientsData);
      setTags(tagsData);
      setDefaultRates(ratesData);
    } catch (e: unknown) {
      setErrorMessage(e instanceof Error ? e.message : 'Ошибка подключения к серверу');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const handleArchive = async (client: Client) => {
    if (!window.confirm(`Переместить ученика «${client.name}» в архив?`)) return;
    try {
      await archiveClient(client.id);
      await loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка архивации ученика');
    }
  };

  const handleUnarchive = async (client: Client) => {
    try {
      await unarchiveClient(client.id);
      await loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка восстановления ученика');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">CRM Репетитора</h1>
          <p className="text-sm text-slate-500 mt-1">Управление базой учеников, архив, тарифные сетки и балансы</p>
        </div>
        <div className="flex items-center gap-2.5 flex-wrap">
          <GlassButton variant="secondary" onClick={() => setIsRatesOpen(true)} icon={<Coins className="w-4 h-4 text-amber-500" />}>
            Базовый прайс
          </GlassButton>
          <GlassButton variant="primary" onClick={() => setIsCreateOpen(true)} icon={<UserPlus className="w-4 h-4" />}>
            Добавить ученика
          </GlassButton>
        </div>
      </div>

      {errorMessage && (
        <div className="p-4 rounded-3xl bg-rose-500/10 border border-rose-500/20 backdrop-blur-md flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
            <p className="text-sm font-medium text-rose-900">{errorMessage}</p>
          </div>
          <GlassButton variant="secondary" size="sm" onClick={loadData} icon={<RefreshCw className="w-3.5 h-3.5" />}>
            Повторить
          </GlassButton>
        </div>
      )}

      <ClientFilters
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        statusTab={statusTab}
        onStatusTabChange={setStatusTab}
        balanceFilter={balanceFilter}
        onBalanceFilterChange={setBalanceFilter}
        selectedTagId={selectedTagId}
        onTagChange={setSelectedTagId}
        tags={tags}
        counts={counts}
      />

      {isLoading ? (
        <div className="p-16 flex flex-col items-center justify-center liquid-glass rounded-3xl">
          <RefreshCw className="w-8 h-8 animate-spin text-indigo-600 mb-3" />
          <p className="text-sm font-medium text-slate-600">Загрузка данных CRM...</p>
        </div>
      ) : filteredClients.length === 0 ? (
        <GlassCard className="text-center p-12 flex flex-col items-center">
          <div className="w-16 h-16 rounded-3xl bg-indigo-50/70 border border-indigo-200/50 flex items-center justify-center text-indigo-500 mb-4 shadow-sm">
            <UserPlus className="w-8 h-8" />
          </div>
          <h3 className="text-lg font-bold text-slate-900 mb-2">Ученики не найдены</h3>
          <p className="text-slate-500 text-sm max-w-sm mb-6">
            {searchQuery || balanceFilter !== 'all' || selectedTagId || statusTab !== 'active'
              ? 'Попробуйте сбросить фильтры поиска или переключить статус.'
              : 'Добавьте первого ученика, настройте тарифную сетку и пополните баланс часов.'}
          </p>
          {clients.length === 0 && (
            <GlassButton variant="primary" onClick={() => setIsCreateOpen(true)} icon={<UserPlus className="w-4 h-4" />}>
              Добавить первого ученика
            </GlassButton>
          )}
        </GlassCard>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-5">
          {filteredClients.map((client) => (
            <ClientCard
              key={client.id}
              client={client}
              onEdit={setEditingClient}
              onAddSubscription={setSubClient}
              onAdjustBalance={setAdjustClient}
              onOpenJournal={setJournalClient}
              onArchive={handleArchive}
              onUnarchive={handleUnarchive}
            />
          ))}
        </div>
      )}

      <CreateClientModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        tags={tags}
        defaultRates={defaultRates}
        onCreated={loadData}
        onTagCreated={(t) => setTags((prev) => [...prev, t])}
      />
      <EditClientModal
        isOpen={Boolean(editingClient)}
        onClose={() => setEditingClient(null)}
        client={editingClient}
        tags={tags}
        onUpdated={loadData}
        onTagCreated={(t) => setTags((prev) => [...prev, t])}
      />
      <AddSubscriptionModal
        isOpen={Boolean(subClient)}
        onClose={() => setSubClient(null)}
        client={subClient}
        onAdded={loadData}
      />
      <AdjustBalanceModal
        isOpen={Boolean(adjustClient)}
        onClose={() => setAdjustClient(null)}
        client={adjustClient}
        onAdjusted={loadData}
      />
      <ClientJournalModal
        isOpen={Boolean(journalClient)}
        onClose={() => setJournalClient(null)}
        client={journalClient}
      />
      <DefaultRatesModal
        isOpen={isRatesOpen}
        onClose={() => setIsRatesOpen(false)}
        defaultRates={defaultRates}
        onSaved={setDefaultRates}
      />
    </div>
  );
};

export default TeacherClientsPage;
