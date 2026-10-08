import React, { useMemo } from 'react';
import { Client } from '../../types/schedule';
import { JournalClientItem } from './JournalClientItem';
import { GlassInput } from '../../shared/components/GlassInput';
import { Search, Users, RefreshCw } from 'lucide-react';

interface JournalSidebarProps {
  clients: Client[];
  selectedClientId: string | null;
  onSelectClient: (client: Client) => void;
  searchQuery: string;
  onSearchChange: (query: string) => void;
  isLoading: boolean;
}

export const JournalSidebar: React.FC<JournalSidebarProps> = ({
  clients,
  selectedClientId,
  onSelectClient,
  searchQuery,
  onSearchChange,
  isLoading,
}) => {
  // Сортировка по last_lesson_at DESC, затем по имени
  const sortedClients = useMemo(() => {
    return [...clients].sort((a, b) => {
      const timeA = a.last_lesson_at ? new Date(a.last_lesson_at).getTime() : 0;
      const timeB = b.last_lesson_at ? new Date(b.last_lesson_at).getTime() : 0;
      if (timeB !== timeA) {
        return timeB - timeA;
      }
      return a.name.localeCompare(b.name, 'ru');
    });
  }, [clients]);

  const filteredClients = useMemo(() => {
    if (!searchQuery.trim()) return sortedClients;
    const q = searchQuery.toLowerCase().trim();
    return sortedClients.filter((c) =>
      c.name.toLowerCase().includes(q) || (c.phone && c.phone.includes(q))
    );
  }, [sortedClients, searchQuery]);

  return (
    <aside className="w-full lg:w-80 shrink-0 flex flex-col h-full rounded-3xl liquid-glass border border-white/60 p-4 shadow-sm overflow-hidden">
      {/* Шапка сайдбара с поиском */}
      <div className="space-y-3 pb-3 border-b border-black/[0.04]">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Users className="w-4 h-4 text-indigo-600" />
            <h2 className="text-base font-bold text-slate-900 tracking-tight">
              Ученики
            </h2>
          </div>
          <span className="text-xs font-semibold px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-700">
            {clients.length}
          </span>
        </div>

        <GlassInput
          icon={<Search className="w-4 h-4 text-slate-400" />}
          placeholder="Поиск по имени или телефону..."
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
        />
      </div>

      {/* Список учеников */}
      <div className="flex-1 overflow-y-auto space-y-2 pt-3 custom-scrollbar pr-1">
        {isLoading ? (
          <div className="py-12 flex flex-col items-center justify-center text-slate-500 space-y-2">
            <RefreshCw className="w-6 h-6 animate-spin text-indigo-600" />
            <span className="text-xs font-medium">Загрузка учеников...</span>
          </div>
        ) : filteredClients.length === 0 ? (
          <div className="py-12 text-center text-slate-500 space-y-1">
            <p className="text-xs font-medium">
              {searchQuery ? 'Никого не найдено' : 'Список учеников пуст'}
            </p>
            {searchQuery && (
              <button
                type="button"
                onClick={() => onSearchChange('')}
                className="text-xs text-indigo-600 hover:underline"
              >
                Сбросить поиск
              </button>
            )}
          </div>
        ) : (
          filteredClients.map((client) => (
            <JournalClientItem
              key={client.id}
              client={client}
              isSelected={client.id === selectedClientId}
              onSelect={onSelectClient}
            />
          ))
        )}
      </div>
    </aside>
  );
};
