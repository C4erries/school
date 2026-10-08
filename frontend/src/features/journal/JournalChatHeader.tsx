import React from 'react';
import { Client } from '../../types/schedule';
import { GlassButton } from '../../shared/components/GlassButton';
import { ArrowLeft, ExternalLink, Phone } from 'lucide-react';

interface JournalChatHeaderProps {
  client: Client;
  onBackMobile: () => void;
  onOpenCRM: (clientId: string) => void;
}

export const JournalChatHeader: React.FC<JournalChatHeaderProps> = ({
  client,
  onBackMobile,
  onOpenCRM,
}) => {
  const indBal = client.balances?.individual_hours ?? client.balance ?? 0;

  return (
    <div className="p-3.5 sm:p-4 rounded-2xl bg-white/40 border border-white/60 backdrop-blur-md shadow-sm flex items-center justify-between gap-3">
      <div className="flex items-center gap-2 sm:gap-3 min-w-0">
        {/* Кнопка «Назад к списку» на мобилках */}
        <button
          type="button"
          onClick={onBackMobile}
          className="lg:hidden p-2 rounded-xl text-slate-500 hover:text-slate-800 hover:bg-white/40 transition-colors shrink-0"
          title="К списку учеников"
        >
          <ArrowLeft className="w-5 h-5" />
        </button>

        <div className="min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <h3 className="text-base sm:text-lg font-bold text-slate-900 tracking-tight truncate">
              {client.name}
            </h3>
            {indBal < 0 ? (
              <span className="px-2 py-0.5 rounded-lg text-xs font-bold bg-rose-500/15 text-rose-800 border border-rose-400/30">
                Долг: {indBal} ч
              </span>
            ) : (
              <span className="px-2 py-0.5 rounded-lg text-xs font-bold bg-emerald-500/15 text-emerald-800 border border-emerald-400/30">
                Остаток: {indBal} ч
              </span>
            )}
          </div>

          {client.phone && (
            <div className="flex items-center text-xs text-slate-500 gap-1.5 mt-0.5 font-medium">
              <Phone className="w-3 h-3 text-slate-400" />
              <span>{client.phone}</span>
            </div>
          )}
        </div>
      </div>

      <div className="flex items-center gap-2 shrink-0">
        <GlassButton
          variant="secondary"
          size="sm"
          onClick={() => onOpenCRM(client.id)}
          icon={<ExternalLink className="w-3.5 h-3.5 text-indigo-600" />}
          className="text-xs"
        >
          <span className="hidden sm:inline">Профиль в CRM</span>
        </GlassButton>
      </div>
    </div>
  );
};
