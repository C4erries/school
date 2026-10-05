import React from 'react';

/**
 * Атмосферный фон Liquid Glass с мягкими оптическими градиентными сферами.
 * Создаёт световую среду, сквозь которую стеклянные карточки
 * преломляют и рассеивают свет, создавая эффект настоящего жидкого стекла.
 */
export const LiquidBackground: React.FC = () => {
  return (
    <div
      aria-hidden="true"
      className="fixed inset-0 pointer-events-none overflow-hidden -z-10 select-none bg-gradient-to-br from-[#F5F6FA] via-[#ECEFF8] to-[#F1F3FA]"
    >
      {/* Мягкая радиальная сетка световых аур */}
      <div
        className="absolute inset-0 opacity-70"
        style={{
          backgroundImage: `
            radial-gradient(at 15% 15%, rgba(99, 102, 241, 0.22) 0px, transparent 45%),
            radial-gradient(at 85% 10%, rgba(244, 63, 94, 0.18) 0px, transparent 40%),
            radial-gradient(at 50% 50%, rgba(139, 92, 246, 0.15) 0px, transparent 50%),
            radial-gradient(at 10% 85%, rgba(16, 185, 129, 0.20) 0px, transparent 45%),
            radial-gradient(at 85% 85%, rgba(6, 182, 212, 0.18) 0px, transparent 45%)
          `,
        }}
      />

      {/* Крупные плавающие градиентные сферы с размытием */}
      <div className="absolute -top-24 -left-24 w-[600px] h-[600px] rounded-full bg-gradient-to-br from-indigo-500/25 to-purple-600/25 blur-[100px] animate-pulse duration-[8000ms]" />
      <div className="absolute top-1/4 -right-32 w-[550px] h-[550px] rounded-full bg-gradient-to-br from-rose-500/20 to-amber-400/20 blur-[110px]" />
      <div className="absolute -bottom-32 left-1/4 w-[650px] h-[650px] rounded-full bg-gradient-to-tr from-emerald-400/22 to-teal-500/20 blur-[120px]" />
      <div className="absolute bottom-1/3 right-1/4 w-[450px] h-[450px] rounded-full bg-gradient-to-tr from-cyan-400/20 to-blue-500/20 blur-[90px]" />
    </div>
  );
};
