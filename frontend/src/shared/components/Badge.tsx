import React from 'react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

interface BadgeProps {
  variant?: 'mint' | 'indigo' | 'amber' | 'coral' | 'neutral';
  children: React.ReactNode;
  className?: string;
}

export const Badge: React.FC<BadgeProps> = ({
  variant = 'indigo',
  children,
  className,
}) => {
  const variantStyles = {
    mint: 'bg-emerald-500/15 text-emerald-800 border-emerald-400/30 shadow-sm shadow-emerald-500/5',
    indigo: 'bg-indigo-500/15 text-indigo-800 border-indigo-400/30 shadow-sm shadow-indigo-500/5',
    amber: 'bg-amber-500/15 text-amber-800 border-amber-400/30 shadow-sm shadow-amber-500/5',
    coral: 'bg-rose-500/15 text-rose-800 border-rose-400/30 shadow-sm shadow-rose-500/5',
    neutral: 'bg-slate-500/15 text-slate-800 border-slate-300/40 shadow-sm',
  };

  return (
    <span
      className={twMerge(
        clsx(
          'inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold border backdrop-blur-md transition-all',
          variantStyles[variant],
          className
        )
      )}
    >
      {children}
    </span>
  );
};
