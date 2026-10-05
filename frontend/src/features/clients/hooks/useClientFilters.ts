import { useState, useMemo } from 'react';
import { Client } from '../../../types/schedule';
import { StatusTab, BalanceFilter } from '../components/ClientFilters';

export function useClientFilters(clients: Client[]) {
  const [searchQuery, setSearchQuery] = useState('');
  const [statusTab, setStatusTab] = useState<StatusTab>('active');
  const [balanceFilter, setBalanceFilter] = useState<BalanceFilter>('all');
  const [selectedTagId, setSelectedTagId] = useState('');

  const counts = useMemo(() => ({
    active: clients.filter((c) => !c.is_archived).length,
    archived: clients.filter((c) => Boolean(c.is_archived)).length,
    all: clients.length,
  }), [clients]);

  const filteredClients = useMemo(() => {
    return clients.filter((c) => {
      if (statusTab === 'active' && c.is_archived) return false;
      if (statusTab === 'archived' && !c.is_archived) return false;

      if (searchQuery.trim()) {
        const query = searchQuery.toLowerCase().trim();
        const nameMatch = c.name.toLowerCase().includes(query);
        const phoneMatch = c.phone?.replace(/\D/g, '').includes(query.replace(/\D/g, ''));
        if (!nameMatch && !phoneMatch) return false;
      }

      const totalHours = c.balances?.total_hours ?? c.balance ?? 0;
      if (balanceFilter === 'positive' && totalHours <= 0) return false;
      if (balanceFilter === 'zero' && totalHours !== 0) return false;
      if (balanceFilter === 'debt' && totalHours >= 0) return false;

      if (selectedTagId) {
        const hasTag =
          c.tags?.some((t) => t.id === selectedTagId) ||
          c.tag_ids?.includes(selectedTagId);
        if (!hasTag) return false;
      }

      return true;
    });
  }, [clients, statusTab, searchQuery, balanceFilter, selectedTagId]);

  return {
    searchQuery,
    setSearchQuery,
    statusTab,
    setStatusTab,
    balanceFilter,
    setBalanceFilter,
    selectedTagId,
    setSelectedTagId,
    counts,
    filteredClients,
  };
}
