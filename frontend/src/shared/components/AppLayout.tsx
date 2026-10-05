import React from 'react';
import { Outlet } from 'react-router-dom';
import { LiquidBackground } from './LiquidBackground';
import { AppNavbar } from './AppNavbar';

/**
 * Сквозной AppLayout согласно ADR-006.
 * Обеспечивает единую визуальную среду Apple Liquid Glass
 * для всех защищенных страниц приложения.
 */
export const AppLayout: React.FC = () => {
  return (
    <div className="relative min-h-screen w-full overflow-x-hidden font-sans">
      <LiquidBackground />
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-6">
        <AppNavbar />
      </div>
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <Outlet />
      </main>
    </div>
  );
};

export default AppLayout;
