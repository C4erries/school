import React from 'react';
import { Client } from '../../../types/schedule';
import { GlassCard } from '../../../shared/components/GlassCard';
import { GlassButton } from '../../../shared/components/GlassButton';
import { Badge } from '../../../shared/components/Badge';
import {
  Phone,
  Coins,
  Clock,
  AlertCircle,
  Pencil,
  Plus,
  SlidersHorizontal,
  Archive,
  RotateCcw,
  BookOpen,
} from 'lucide-react';

interface ClientCardProps {
  client: Client;
  onEdit: (client: Client) => void;
  onAddSubscription: (client: Client) => void;
  onAdjustBalance: (client: Client) => void;
  onOpenJournal: (client: Client) => void;
  onArchive: (client: Client) => void;
  onUnarchive: (client: Client) => void;
}

export const ClientCard: React.FC<ClientCardProps> = ({
  client,
  onEdit,
  onAddSubscription,
  onAdjustBalance,
  onOpenJournal,
  onArchive,
  onUnarchive,
}) => {
  const indBal = client.balances?.individual_hours ?? client.balance ?? 0;
  const pairBal = client.balances?.pair_hours ?? 0;
  const grpBal = client.balances?.group_hours ?? 0;
  const isArchived = Boolean(client.is_archived);

  const renderBalanceBadge = (hours: number, label: string) => {
    if (hours > 0) {
      return (
        <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-xl text-xs font-bold bg-emerald-500/15 text-emerald-800 border border-emerald-400/30 shadow-sm">
          <Clock className="w-3.5 h-3.5 text-emerald-600" />
          {hours} ч {label}
        </span>
      );
    }
    if (hours < 0) {
      return (
        <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-xl text-xs font-bold bg-rose-500/15 text-rose-800 border border-rose-400/30 shadow-sm">
          <AlertCircle className="w-3.5 h-3.5 text-rose-600" />
          {hours} ч (долг)
        </span>
      );
    }
    return (
      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-xl text-xs font-semibold bg-white/40 text-slate-500 border border-white/60">
        0 ч {label}
      </span>
    );
  };

  return (
    <GlassCard
      className={`flex flex-col p-5 sm:p-6 transition-all duration-200 hover:shadow-lg hover:border-white/90 ${
        isArchived ? 'opacity-75 bg-slate-100/40' : ''
      }`}
    >
      {/* Шапка карточки */}
      <div className="flex justify-between items-start mb-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2 flex-wrap">
            <h3 className="text-lg font-bold text-slate-900 tracking-tight">
              {client.name}
            </h3>
            {isArchived && (
              <Badge variant="neutral" className="text-[10px] font-bold">
                В архиве
              </Badge>
            )}
            <button
              type="button"
              onClick={() => onEdit(client)}
              className="p-1.5 rounded-xl hover:bg-white/40 text-slate-400 hover:text-slate-700 transition-colors"
              title="Редактировать ученика"
            >
              <Pencil className="w-3.5 h-3.5" />
            </button>
          </div>
          {client.phone && (
            <div className="flex items-center text-xs text-slate-500 gap-1.5 font-medium">
              <Phone className="w-3.5 h-3.5 text-slate-400" />
              {client.phone}
            </div>
          )}
        </div>

        {/* Теги ученика */}
        <div className="flex flex-wrap gap-1 justify-end max-w-[150px]">
          {client.tags && client.tags.length > 0 ? (
            client.tags.map((t) => (
              <Badge
                key={t.id}
                variant="amber"
                className="text-[10px] font-bold tracking-tight"
              >
                {t.name} {t.school_percent > 0 ? `(${t.school_percent}%)` : ''}
              </Badge>
            ))
          ) : client.tag ? (
            <Badge variant="amber" className="text-[10px] font-bold tracking-tight">
              {client.tag}
            </Badge>
          ) : null}
        </div>
      </div>

      {/* Тарифная сетка ставок */}
      <div className="space-y-2 mb-4">
        <div className="text-[11px] font-bold uppercase tracking-wider text-slate-500 flex items-center gap-1.5 ml-1">
          <Coins className="w-3.5 h-3.5 text-indigo-500" />
          Тарифная сетка
        </div>

        <div className="grid grid-cols-3 gap-2">
          <div className="p-2.5 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 shadow-sm text-center">
            <span className="block text-[10px] font-semibold text-slate-500 uppercase">
              Индив.
            </span>
            <span className="text-xs sm:text-sm font-bold text-slate-900">
              {client.rate_individual} ₽/ч
            </span>
          </div>

          <div className="p-2.5 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 shadow-sm text-center">
            <span className="block text-[10px] font-semibold text-slate-500 uppercase">
              Пара
            </span>
            <span className="text-xs sm:text-sm font-bold text-slate-900">
              {client.rate_pair ? `${client.rate_pair} ₽/ч` : '—'}
            </span>
          </div>

          <div className="p-2.5 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 shadow-sm text-center">
            <span className="block text-[10px] font-semibold text-slate-500 uppercase">
              Группа
            </span>
            <span className="text-xs sm:text-sm font-bold text-slate-900">
              {client.rate_group ? `${client.rate_group} ₽/ч` : '—'}
            </span>
          </div>
        </div>
      </div>

      {/* Баланс часов по форматам */}
      <div className="p-3.5 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 shadow-sm mb-5 space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-500">
            Остаток абонемента
          </span>
          <span className="text-xs font-semibold text-slate-600">
            Всего: <strong className="text-slate-900">{client.balance ?? indBal} ч</strong>
          </span>
        </div>

        <div className="flex flex-wrap items-center gap-2 pt-1">
          {renderBalanceBadge(indBal, 'индив.')}
          {(client.rate_pair || pairBal !== 0) && renderBalanceBadge(pairBal, 'пара')}
          {(client.rate_group || grpBal !== 0) && renderBalanceBadge(grpBal, 'группа')}
        </div>
      </div>

      {/* Кнопки действий */}
      <div className="mt-auto space-y-2">
        <GlassButton
          variant="secondary"
          size="sm"
          onClick={() => onOpenJournal(client)}
          icon={<BookOpen className="w-3.5 h-3.5 text-indigo-600" />}
          className="w-full"
        >
          Дневник & ДЗ
        </GlassButton>

        <div className="grid grid-cols-2 gap-2">
          <GlassButton
            variant="secondary"
            size="sm"
            onClick={() => onAddSubscription(client)}
            icon={<Plus className="w-3.5 h-3.5 text-indigo-600" />}
          >
            Пополнить
          </GlassButton>

          <GlassButton
            variant="secondary"
            size="sm"
            onClick={() => onAdjustBalance(client)}
            icon={<SlidersHorizontal className="w-3.5 h-3.5 text-slate-600" />}
          >
            Баланс
          </GlassButton>
        </div>

        <div className="pt-1 flex justify-end">
          {isArchived ? (
            <button
              type="button"
              onClick={() => onUnarchive(client)}
              className="text-xs font-semibold text-emerald-600 hover:text-emerald-700 flex items-center gap-1.5 py-1 px-2 rounded-xl hover:bg-emerald-500/10 transition-colors"
            >
              <RotateCcw className="w-3.5 h-3.5" />
              Восстановить из архива
            </button>
          ) : (
            <button
              type="button"
              onClick={() => onArchive(client)}
              className="text-xs font-semibold text-slate-400 hover:text-slate-600 flex items-center gap-1.5 py-1 px-2 rounded-xl hover:bg-black/5 transition-colors"
            >
              <Archive className="w-3.5 h-3.5" />
              В архив
            </button>
          )}
        </div>
      </div>
    </GlassCard>
  );
};
