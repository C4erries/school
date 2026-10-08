import React from 'react';

interface LessonScoreChipsProps {
  score: number | null;
  onChange: (score: number | null) => void;
  disabled?: boolean;
}

const SCORE_CONFIG: Record<
  number,
  { label: string; activeClass: string; desc: string }
> = {
  1: {
    label: '1',
    desc: 'Очень сложно',
    activeClass: 'bg-rose-500 text-white shadow-rose-500/30 border-rose-400',
  },
  2: {
    label: '2',
    desc: 'С трудом',
    activeClass: 'bg-orange-500 text-white shadow-orange-500/30 border-orange-400',
  },
  3: {
    label: '3',
    desc: 'Удовлетворительно',
    activeClass: 'bg-amber-500 text-white shadow-amber-500/30 border-amber-400',
  },
  4: {
    label: '4',
    desc: 'Хорошо',
    activeClass: 'bg-indigo-600 text-white shadow-indigo-600/30 border-indigo-400',
  },
  5: {
    label: '5',
    desc: 'Отлично усвоено',
    activeClass: 'bg-emerald-500 text-white shadow-emerald-500/30 border-emerald-400',
  },
};

export const LessonScoreChips: React.FC<LessonScoreChipsProps> = ({
  score,
  onChange,
  disabled = false,
}) => {
  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between">
        <label className="block text-xs font-semibold text-slate-700">
          Оценка понимания материала
        </label>
        {score && (
          <span className="text-xs font-medium text-slate-500 animate-in fade-in">
            {SCORE_CONFIG[score]?.desc}
          </span>
        )}
      </div>

      <div className="grid grid-cols-5 gap-2">
        {[1, 2, 3, 4, 5].map((val) => {
          const isSelected = score === val;
          const conf = SCORE_CONFIG[val];
          return (
            <button
              key={val}
              type="button"
              disabled={disabled}
              onClick={() => onChange(isSelected ? null : val)}
              className={`h-11 rounded-2xl flex flex-col items-center justify-center font-bold text-sm transition-all duration-200 border shadow-xs select-none active:scale-95 disabled:opacity-50 ${
                isSelected
                  ? `${conf.activeClass} shadow-md scale-[1.02]`
                  : 'bg-white/40 hover:bg-white/70 backdrop-blur-md text-slate-700 border-white/60'
              }`}
            >
              <span>{val}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
};
