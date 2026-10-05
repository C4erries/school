import React, { forwardRef } from 'react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

interface GlassInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: React.ReactNode;
  labelClassName?: string;
  error?: string;
  icon?: React.ReactNode;
}

export const GlassInput = forwardRef<HTMLInputElement, GlassInputProps>(
  ({ label, labelClassName, error, icon, className, ...props }, ref) => {
    return (
      <div className="w-full space-y-1.5 text-left">
        {label && (
          <label
            className={twMerge(
              'block text-xs font-semibold uppercase tracking-wider text-slate-500 ml-1 whitespace-nowrap',
              labelClassName
            )}
          >
            {label}
          </label>
        )}
        <div className="relative flex items-center">
          {icon && (
            <div className="absolute left-3.5 text-slate-400 pointer-events-none flex items-center justify-center">
              {icon}
            </div>
          )}
          <input
            ref={ref}
            className={twMerge(
              clsx(
                'w-full rounded-2xl px-4 py-3 text-sm text-slate-800 placeholder-slate-400',
                'liquid-glass-input transition-all duration-200',
                icon && 'pl-11',
                error && 'border-rose-400 focus:border-rose-500 focus:ring-rose-200',
                className
              )
            )}
            {...props}
          />
        </div>
        {error && (
          <p className="text-xs font-medium text-rose-500 ml-1 mt-1">
            {error}
          </p>
        )}
      </div>
    );
  }
);

GlassInput.displayName = 'GlassInput';
