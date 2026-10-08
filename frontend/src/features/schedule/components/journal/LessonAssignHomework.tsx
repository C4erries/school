import React, { useState } from 'react';
import { HomeworkAssignment, HomeworkStatus } from '../../../../types/journal';
import { Badge } from '../../../../shared/components/Badge';
import { BookOpen, Calendar, ChevronDown, ChevronUp } from 'lucide-react';

interface LessonAssignHomeworkProps {
  assignedHomeworks: HomeworkAssignment[];
  title: string;
  onTitleChange: (val: string) => void;
  description: string;
  onDescriptionChange: (val: string) => void;
  dueDate: string;
  onDueDateChange: (val: string) => void;
  disabled?: boolean;
}

export const LessonAssignHomework: React.FC<LessonAssignHomeworkProps> = ({
  assignedHomeworks,
  title,
  onTitleChange,
  description,
  onDescriptionChange,
  dueDate,
  onDueDateChange,
  disabled = false,
}) => {
  const [isOpen, setIsOpen] = useState(true);

  const renderStatusBadge = (status: HomeworkStatus) => {
    switch (status) {
      case 'completed':
        return <Badge variant="mint">Сдано</Badge>;
      case 'not_done':
        return <Badge variant="danger">Не сдано</Badge>;
      case 'assigned':
      default:
        return <Badge variant="amber">В работе</Badge>;
    }
  };

  return (
    <div className="space-y-3 p-4 rounded-2xl bg-white/30 backdrop-blur-md border border-white/50 shadow-xs">
      <div
        className="flex items-center justify-between cursor-pointer select-none"
        onClick={() => setIsOpen((prev) => !prev)}
      >
        <div className="flex items-center gap-2">
          <BookOpen className="w-4 h-4 text-indigo-600" />
          <span className="text-xs font-bold uppercase tracking-wider text-slate-700">
            Домашнее задание к следующему уроку
          </span>
          {assignedHomeworks.length > 0 && (
            <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-indigo-500/15 text-indigo-800 border border-indigo-400/30">
              {assignedHomeworks.length}
            </span>
          )}
        </div>
        <button
          type="button"
          className="text-slate-400 hover:text-slate-600 p-1"
          aria-label={isOpen ? 'Свернуть блок ДЗ' : 'Развернуть блок ДЗ'}
        >
          {isOpen ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
        </button>
      </div>

      {isOpen && (
        <div className="space-y-3 pt-1 animate-in fade-in">
          {/* Уже заданные на этом уроке ДЗ */}
          {assignedHomeworks.length > 0 && (
            <div className="space-y-1.5 pb-2 border-b border-black/[0.05]">
              <span className="text-[11px] font-semibold text-slate-500">
                Ранее задано на этом уроке:
              </span>
              <div className="space-y-1.5">
                {assignedHomeworks.map((hw) => (
                  <div
                    key={hw.id}
                    className="p-2.5 rounded-xl bg-white/50 backdrop-blur-md border border-white/60 flex items-center justify-between gap-2"
                  >
                    <div className="min-w-0">
                      <div className="text-xs font-bold text-slate-800 truncate">
                        {hw.title}
                      </div>
                      {hw.due_date && (
                        <div className="flex items-center gap-1 text-[10px] text-slate-500 font-mono">
                          <Calendar className="w-2.5 h-2.5 text-slate-400" />
                          <span>до {new Date(hw.due_date).toLocaleDateString('ru-RU')}</span>
                        </div>
                      )}
                    </div>
                    {renderStatusBadge(hw.status)}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Форма назначения нового ДЗ */}
          <div className="space-y-2">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
              <div className="sm:col-span-2 space-y-1">
                <label className="block text-[11px] font-semibold text-slate-600 ml-1">
                  Заголовок / Что задано
                </label>
                <input
                  type="text"
                  disabled={disabled}
                  placeholder="Параграф 14, № 120-128"
                  value={title}
                  onChange={(e) => onTitleChange(e.target.value)}
                  className="w-full rounded-xl px-3 py-2 text-xs text-slate-800 placeholder-slate-400 liquid-glass-input"
                />
              </div>

              <div className="space-y-1">
                <label className="block text-[11px] font-semibold text-slate-600 ml-1">
                  Срок сдачи
                </label>
                <input
                  type="date"
                  disabled={disabled}
                  value={dueDate}
                  onChange={(e) => onDueDateChange(e.target.value)}
                  className="w-full rounded-xl px-2.5 py-2 text-xs text-slate-800 liquid-glass-input font-mono"
                />
              </div>
            </div>

            <div className="space-y-1">
              <label className="block text-[11px] font-semibold text-slate-600 ml-1">
                Пояснения к заданию (опционально)
              </label>
              <textarea
                disabled={disabled}
                placeholder="Обратить внимание на формулу дискриминанта, оформить в тетради..."
                value={description}
                onChange={(e) => onDescriptionChange(e.target.value)}
                rows={2}
                className="w-full rounded-xl p-2.5 text-xs text-slate-800 placeholder-slate-400 liquid-glass-input resize-none"
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
