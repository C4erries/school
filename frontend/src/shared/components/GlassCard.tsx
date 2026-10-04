import React from 'react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

interface GlassCardProps extends React.HTMLAttributes<HTMLDivElement> {
  interactive?: boolean;
  elevation?: 'low' | 'medium' | 'high';
  children: React.ReactNode;
}

export const GlassCard: React.FC<GlassCardProps> = ({
  interactive = false,
  className,
  children,
  ...props
}) => {
  return (
    <div
      className={twMerge(
        clsx(
          'rounded-3xl p-6 sm:p-8',
          interactive ? 'liquid-glass-interactive cursor-pointer' : 'liquid-glass',
          className
        )
      )}
      {...props}
    >
      {children}
    </div>
  );
};
