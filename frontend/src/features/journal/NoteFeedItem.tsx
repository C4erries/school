import React, { useState } from 'react';
import { ClientNote } from '../../types/journal';
import { MessageSquare, Trash2, Clock, Loader2 } from 'lucide-react';

interface NoteFeedItemProps {
  note: ClientNote;
  onDeleteNote: (noteId: string) => Promise<void>;
}

export const NoteFeedItem: React.FC<NoteFeedItemProps> = ({
  note,
  onDeleteNote,
}) => {
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDelete = async () => {
    if (!window.confirm('Удалить эту заметку?')) return;
    setIsDeleting(true);
    try {
      await onDeleteNote(note.id);
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка удаления');
      setIsDeleting(false);
    }
  };

  const timeStr = new Date(note.created_at).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  });

  return (
    <div className="p-4 rounded-2xl bg-amber-500/10 border border-amber-300/30 backdrop-blur-md shadow-sm transition-all hover:bg-amber-500/15 group">
      <div className="flex items-center justify-between gap-3 mb-2">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 rounded-lg bg-amber-500/20 text-amber-800 flex items-center justify-center shrink-0">
            <MessageSquare className="w-3.5 h-3.5" />
          </div>
          <span className="text-xs font-bold text-amber-950">
            Заметка преподавателя
          </span>
        </div>

        <div className="flex items-center gap-2">
          <div className="flex items-center gap-1 text-[11px] text-amber-800/70">
            <Clock className="w-3 h-3" />
            <span>{timeStr}</span>
          </div>
          <button
            type="button"
            disabled={isDeleting}
            onClick={handleDelete}
            title="Удалить заметку"
            className="p-1 rounded-lg text-amber-700/50 hover:text-rose-600 hover:bg-white/40 transition-colors opacity-0 group-hover:opacity-100 focus:opacity-100"
          >
            {isDeleting ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <Trash2 className="w-3.5 h-3.5" />
            )}
          </button>
        </div>
      </div>

      <p className="text-xs sm:text-sm text-slate-800 whitespace-pre-wrap leading-relaxed pl-8">
        {note.content}
      </p>
    </div>
  );
};
