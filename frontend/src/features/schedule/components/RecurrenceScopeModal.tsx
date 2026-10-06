import React from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassButton } from '../../../shared/components/GlassButton';
import { RecurrenceScope } from '../../../types/schedule';
import { Calendar, CalendarRange, RotateCw, AlertCircle } from 'lucide-react';

interface RecurrenceScopeModalProps {
  isOpen: boolean;
  onClose: () => void;
  actionType: 'edit' | 'cancel';
  onConfirm: (scope: RecurrenceScope) => void;
  lessonTitle?: string;
}

export const RecurrenceScopeModal: React.FC<RecurrenceScopeModalProps> = ({
  isOpen,
  onClose,
  actionType,
  onConfirm,
  lessonTitle,
}) => {
  const [selectedScope, setSelectedScope] = React.useState<RecurrenceScope>('this_only');

  React.useEffect(() => {
    if (isOpen) {
      setSelectedScope('this_only');
    }
  }, [isOpen]);

  const isCancel = actionType === 'cancel';
  const titleText = isCancel ? 'Отмена регулярного занятия' : 'Редактирование регулярного занятия';
  const descText = isCancel
    ? 'Этот урок является частью повторяющейся серии. Выберите область действия отмены:'
    : 'Этот урок является частью повторяющейся серии. Выберите область применения изменений:';

  const options: Array<{
    scope: RecurrenceScope;
    title: string;
    description: string;
    icon: React.ReactNode;
  }> = [
    {
      scope: 'this_only',
      title: 'Только этот урок',
      description: isCancel
        ? 'Отменить только занятие в выбранный день. Остальные уроки серии останутся в расписании.'
        : 'Изменения применятся исключительно к этому выбранному дню.',
      icon: <Calendar className="w-5 h-5 text-indigo-600" />,
    },
    {
      scope: 'this_and_following',
      title: 'Этот и все последующие',
      description: isCancel
        ? 'Завершить серию этим днем: текущий и все будущие уроки будут отменены.'
        : 'Применить новые параметры к текущему и всем будущим урокам серии.',
      icon: <CalendarRange className="w-5 h-5 text-indigo-600" />,
    },
    {
      scope: 'all_in_series',
      title: 'Все уроки серии',
      description: isCancel
        ? 'Отменить всю регулярную цепочку занятий целиком (все уроки серии).'
        : 'Обновить параметры для всей серии целиком (включая все уроки).',
      icon: <RotateCw className="w-5 h-5 text-indigo-600" />,
    },
  ];

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title={titleText}
      description={lessonTitle ? `Серия: «${lessonTitle}»` : undefined}
      maxWidth="md"
    >
      <div className="space-y-4">
        <div className="flex items-center gap-2 p-3 rounded-2xl bg-indigo-50/40 border border-indigo-200/50 text-indigo-900 text-xs">
          <AlertCircle className="w-4 h-4 text-indigo-600 shrink-0" />
          <span>{descText}</span>
        </div>

        <div className="space-y-2.5">
          {options.map((opt) => {
            const isSelected = selectedScope === opt.scope;
            return (
              <label
                key={opt.scope}
                onClick={() => setSelectedScope(opt.scope)}
                className={`flex items-start gap-3 p-3.5 rounded-2xl border transition-all cursor-pointer ${
                  isSelected
                    ? 'bg-indigo-600/10 border-indigo-500/50 shadow-sm'
                    : 'bg-white/40 border-white/60 hover:bg-white/60'
                }`}
              >
                <input
                  type="radio"
                  name="recurrence_scope"
                  value={opt.scope}
                  checked={isSelected}
                  onChange={() => setSelectedScope(opt.scope)}
                  className="mt-1 h-4 w-4 text-indigo-600 border-slate-300 focus:ring-indigo-500"
                />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-sm text-slate-900">{opt.title}</span>
                  </div>
                  <p className="text-xs text-slate-500 mt-0.5 leading-relaxed">
                    {opt.description}
                  </p>
                </div>
                <div className="shrink-0 p-1.5 rounded-xl bg-white/60">
                  {opt.icon}
                </div>
              </label>
            );
          })}
        </div>

        <div className="pt-3 flex justify-end gap-2.5 border-t border-black/[0.05]">
          <GlassButton
            type="button"
            variant="secondary"
            onClick={onClose}
          >
            Назад
          </GlassButton>
          <GlassButton
            type="button"
            variant={isCancel ? 'coral' : 'primary'}
            onClick={() => {
              onConfirm(selectedScope);
              onClose();
            }}
          >
            {isCancel ? 'Подтвердить отмену' : 'Продолжить'}
          </GlassButton>
        </div>
      </div>
    </GlassModal>
  );
};

