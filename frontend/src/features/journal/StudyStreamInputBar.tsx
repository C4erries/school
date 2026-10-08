import React, { useState } from 'react';
import { GlassButton } from '../../shared/components/GlassButton';
import { Send, FileEdit, Loader2 } from 'lucide-react';

interface StudyStreamInputBarProps {
  onSendNote: (content: string) => Promise<void>;
  onOpenReportModal?: () => void;
  disabled?: boolean;
}

export const StudyStreamInputBar: React.FC<StudyStreamInputBarProps> = ({
  onSendNote,
  onOpenReportModal,
  disabled = false,
}) => {
  const [content, setContent] = useState('');
  const [isSending, setIsSending] = useState(false);

  const handleSubmit = async (e?: React.FormEvent) => {
    e?.preventDefault();
    const trimmed = content.trim();
    if (!trimmed || isSending || disabled) return;

    setIsSending(true);
    try {
      await onSendNote(trimmed);
      setContent('');
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка отправки заметки');
    } finally {
      setIsSending(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    // Enter отправляет заметку, Shift+Enter переносит строку
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  };

  return (
    <div className="p-3 sm:p-4 rounded-2xl bg-white/40 border border-white/60 backdrop-blur-md shadow-md space-y-2.5">
      <form onSubmit={handleSubmit} className="flex items-end gap-2.5">
        <div className="flex-1 min-w-0 relative">
          <textarea
            rows={2}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={disabled || isSending}
            placeholder="Быстрая заметка по ученику... (Enter для отправки, Shift+Enter перенос строки)"
            className="w-full rounded-2xl p-3 pr-4 text-xs sm:text-sm text-slate-800 placeholder-slate-400 liquid-glass-input resize-none focus:outline-none"
          />
        </div>

        <div className="flex items-center gap-2 shrink-0">
          {onOpenReportModal && (
            <GlassButton
              type="button"
              variant="secondary"
              onClick={onOpenReportModal}
              disabled={disabled || isSending}
              icon={<FileEdit className="w-4 h-4 text-indigo-600" />}
              title="Заполнить отчет по уроку"
              className="hidden sm:inline-flex"
            >
              Отчет по уроку
            </GlassButton>
          )}

          <GlassButton
            type="submit"
            variant="primary"
            disabled={!content.trim() || disabled || isSending}
            icon={
              isSending ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Send className="w-4 h-4" />
              )
            }
          >
            <span className="hidden sm:inline">Отправить</span>
          </GlassButton>
        </div>
      </form>
    </div>
  );
};
