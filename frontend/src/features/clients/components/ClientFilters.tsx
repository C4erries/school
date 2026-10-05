import React from 'react';
import { Tag } from '../../../types/schedule';
import { Search, X, Filter } from 'lucide-react';

export type StatusTab = 'active' | 'archived' | 'all';
export type BalanceFilter = 'all' | 'positive' | 'zero' | 'debt';

interface ClientFiltersProps {
  searchQuery: string;
  onSearchChange: (query: string) => void;
  statusTab: StatusTab;
  onStatusTabChange: (tab: StatusTab) => void;
  balanceFilter: BalanceFilter;
  onBalanceFilterChange: (filter: BalanceFilter) => void;
  selectedTagId: string;
  onTagChange: (tagId: string) => void;
  tags: Tag[];
  counts: {
    active: number;
    archived: number;
    all: number;
  };
}

export const ClientFilters: React.FC<ClientFiltersProps> = ({
  searchQuery,
  onSearchChange,
  statusTab,
  onStatusTabChange,
  balanceFilter,
  onBalanceFilterChange,
  selectedTagId,
  onTagChange,
  tags,
  counts,
}) => {
  return (
    <div className="space-y-3 p-4 rounded-3xl bg-white/30 backdrop-blur-md border border-white/40 shadow-sm">
      <div className="flex flex-col md:flex-row items-stretch md:items-center justify-between gap-3">
        {/* Поиск по имени или телефону */}
        <div className="relative flex-1">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-3.5 pointer-events-none" />
          <input
            type="text"
            placeholder="Поиск по имени или телефону..."
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            className="w-full pl-10 pr-9 py-2.5 rounded-2xl bg-white/60 border border-white/80 text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:border-indigo-500 shadow-sm"
          />
          {searchQuery && (
            <button
              type="button"
              onClick={() => onSearchChange('')}
              className="absolute right-3 top-3 p-0.5 text-slate-400 hover:text-slate-600 rounded-full hover:bg-black/5"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>

        {/* Табы статуса: Активные | В архиве | Все */}
        <div className="flex items-center gap-1 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05] shrink-0">
          <button
            type="button"
            onClick={() => onStatusTabChange('active')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              statusTab === 'active'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <span>Активные</span>
            <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-indigo-50 text-indigo-600">
              {counts.active}
            </span>
          </button>

          <button
            type="button"
            onClick={() => onStatusTabChange('archived')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              statusTab === 'archived'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <span>Архив</span>
            <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-slate-200 text-slate-700">
              {counts.archived}
            </span>
          </button>

          <button
            type="button"
            onClick={() => onStatusTabChange('all')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              statusTab === 'all'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <span>Все</span>
            <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-slate-200 text-slate-700">
              {counts.all}
            </span>
          </button>
        </div>
      </div>

      {/* Фильтры второго уровня: по балансу и по тегам */}
      <div className="flex flex-wrap items-center gap-3 pt-1 text-xs">
        <div className="flex items-center gap-1.5 text-slate-500 font-semibold">
          <Filter className="w-3.5 h-3.5 text-slate-400" />
          <span>Баланс:</span>
        </div>

        <div className="flex items-center gap-1.5 flex-wrap">
          {(
            [
              { id: 'all', label: 'Все' },
              { id: 'positive', label: 'С абонементом (> 0 ч)' },
              { id: 'zero', label: 'Нулевой (0 ч)' },
              { id: 'debt', label: 'Должники (< 0 ч)' },
            ] as const
          ).map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => onBalanceFilterChange(item.id)}
              className={`px-2.5 py-1 rounded-xl font-medium transition-all ${
                balanceFilter === item.id
                  ? 'bg-indigo-600 text-white shadow-sm'
                  : 'bg-white/50 text-slate-600 hover:bg-white/80 border border-white/60'
              }`}
            >
              {item.label}
            </button>
          ))}
        </div>

        {tags.length > 0 && (
          <div className="flex items-center gap-2 ml-auto">
            <span className="text-slate-500 font-semibold">Тег:</span>
            <select
              value={selectedTagId}
              onChange={(e) => onTagChange(e.target.value)}
              className="rounded-xl px-2.5 py-1 bg-white/60 border border-white/80 text-xs font-medium text-slate-800 focus:outline-none"
            >
              <option value="">Все теги</option>
              {tags.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name}
                </option>
              ))}
            </select>
          </div>
        )}
      </div>
    </div>
  );
};
