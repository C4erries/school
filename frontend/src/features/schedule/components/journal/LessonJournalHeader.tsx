import React from 'react';
import { GlassButton } from '../../../../shared/components/GlassButton';
import { BookOpen, Calendar, Clock, ExternalLink, Sparkles } from 'lucide-react';

interface LessonJournalHeaderProps {
  clientDisplayName: string;
  lessonDateStr: string;
  lessonTimeStr: string;
  restoredDraftBadge: boolean;
  onOpenInFullJournal: () => void;
}

export const LessonJournalHeader: React.FC<LessonJournalHeaderProps> = ({
  clientDisplayName,
  lessonDateStr,
  lessonTimeStr,
  restoredDraftBadge,
  onOpenInFullJournal,
}) => {
  return (
    <div className="flex items-center justify-between gap-3 p-3 rounded-2xl bg-white/35 backdrop-blur-md border border-white/60 text-xs text-slate-600 flex-wrap">
      <div className="flex items-center gap-2 flex-wrap">
        <div className="flex items-center gap-1.5 font-semibold text-slate-800">
          <BookOpen className="w-4 h-4 text-indigo-600" />
          <span>{clientDisplayName}</span>
        </div>
        <span className="text-slate-300">•</span>
        <div className="flex items-center gap-1 font-mono text-slate-600">
          <Clock className="w-3.5 h-3.5 text-slate-400" />
          <span>{lessonTimeStr}</span>
        </div>
        <span className="text-slate-300">•</span>
        <div className="flex items-center gap-1 text-slate-600">
          <Calendar className="w-3.5 h-3.5 text-slate-400" />
          <span>{lessonDateStr}</span>
        </div>
      </div>

      <div className="flex items-center gap-2">
        {restoredDraftBadge && (
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-xl text-[11px] font-semibold bg-amber-500/15 text-amber-900 border border-amber-400/30 animate-in fade-in">
            <Sparkles className="w-3 h-3 text-amber-600" />
            Восстановлен черновик
          </span>
        )}
        <GlassButton
          type="button"
          variant="secondary"
          size="sm"
          onClick={onOpenInFullJournal}
          icon={<ExternalLink className="w-3.5 h-3.5 text-indigo-600" />}
          className="text-xs"
        >
          Открыть в журнале ↗
        </GlassButton>
      </div>
    </div>
  );
};
