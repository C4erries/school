import React, { useEffect } from 'react';
import { X } from 'lucide-react';

interface GlassModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children: React.ReactNode;
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl';
}

export const GlassModal: React.FC<GlassModalProps> = ({
  isOpen,
  onClose,
  title,
  description,
  children,
  maxWidth = 'md',
}) => {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };
    if (isOpen) {
      document.body.style.overflow = 'hidden';
      window.addEventListener('keydown', handleKeyDown);
    }
    return () => {
      document.body.style.overflow = '';
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const maxWidthClass = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
    xl: 'max-w-xl',
  }[maxWidth];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 overflow-y-auto">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-slate-900/40 backdrop-blur-md transition-opacity duration-300"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Modal Dialog Card */}
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
        className={`relative w-full ${maxWidthClass} max-h-[90vh] flex flex-col rounded-3xl liquid-glass shadow-2xl border border-white/70 z-10 transition-all duration-300 transform scale-100 animate-in fade-in zoom-in-95 overflow-hidden my-auto`}
      >
        <div className="p-5 sm:p-6 pb-3.5 shrink-0 flex items-start justify-between gap-4 border-b border-black/[0.04]">
          <div className="space-y-1 min-w-0 pr-2">
            <h2 id="modal-title" className="text-lg sm:text-xl font-bold tracking-tight text-slate-900 truncate">
              {title}
            </h2>
            {description && (
              <p className="text-xs text-slate-500 leading-relaxed">
                {description}
              </p>
            )}
          </div>
          <button
            type="button"
            onClick={onClose}
            className="w-8 h-8 rounded-full flex items-center justify-center text-slate-400 hover:text-slate-700 hover:bg-black/5 transition-colors focus:outline-none shrink-0"
            aria-label="Закрыть"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-5 sm:p-6 pt-4 overflow-y-auto custom-scrollbar flex-1">
          {children}
        </div>
      </div>
    </div>
  );
};
