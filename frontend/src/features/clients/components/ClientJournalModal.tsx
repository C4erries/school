import React, { useState } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { Client } from '../../../types/schedule';
import { ClientLessonsTimeline } from './journal/ClientLessonsTimeline';
import { ClientHomeworkList } from './journal/ClientHomeworkList';
import { BookOpen, CheckSquare } from 'lucide-react';

interface ClientJournalModalProps {
  isOpen: boolean;
  onClose: () => void;
  client: Client | null;
}

type ActiveTab = 'lessons' | 'homework';

export const ClientJournalModal: React.FC<ClientJournalModalProps> = ({
  isOpen,
  onClose,
  client,
}) => {
  const [activeTab, setActiveTab] = useState<ActiveTab>('lessons');

  if (!client) return null;

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title={`Дневник & ДЗ: ${client.name}`}
      description="Хронология пройденных тем, оценки усвоения и архив домашних заданий"
      maxWidth="lg"
    >
      <div className="space-y-4">
        {/* Переключатель вкладок (Segmented Control) */}
        <div className="p-1 rounded-2xl bg-black/[0.04] backdrop-blur-md border border-white/50 grid grid-cols-2 gap-1.5 shadow-2xs">
          <button
            type="button"
            onClick={() => setActiveTab('lessons')}
            className={`py-2 px-3 rounded-xl text-xs sm:text-sm font-bold flex items-center justify-center gap-2 transition-all ${
              activeTab === 'lessons'
                ? 'bg-white/85 text-indigo-950 shadow-sm border border-white/80'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/30'
            }`}
          >
            <BookOpen className="w-4 h-4 text-indigo-600" />
            <span>История уроков</span>
          </button>

          <button
            type="button"
            onClick={() => setActiveTab('homework')}
            className={`py-2 px-3 rounded-xl text-xs sm:text-sm font-bold flex items-center justify-center gap-2 transition-all ${
              activeTab === 'homework'
                ? 'bg-white/85 text-indigo-950 shadow-sm border border-white/80'
                : 'text-slate-600 hover:text-slate-900 hover:bg-white/30'
            }`}
          >
            <CheckSquare className="w-4 h-4 text-emerald-600" />
            <span>Домашние задания</span>
          </button>
        </div>

        {/* Контент активной вкладки */}
        <div className="pt-1">
          {activeTab === 'lessons' ? (
            <ClientLessonsTimeline clientId={client.id} />
          ) : (
            <ClientHomeworkList clientId={client.id} />
          )}
        </div>
      </div>
    </GlassModal>
  );
};
