import React from 'react';
import type { HealthStatus } from '@/types/health';

interface HeaderProps {
  status: HealthStatus;
}

export const Header: React.FC<HeaderProps> = ({ status }) => {
  const getStatusLabel = (s: HealthStatus): string => {
    switch (s) {
      case 'healthy':
        return 'Backend Connected';
      case 'checking':
        return 'Connecting...';
      case 'unhealthy':
        return 'Disconnected';
      case 'idle':
        return 'Standby';
    }
  };

  return (
    <header className="app-header">
      <div className="header-brand">
        <span className="brand-logo" role="img" aria-label="School Platform">🎓</span>
        <div>
          <h1 className="brand-title">School Platform</h1>
          <p className="brand-subtitle">Репетиторский центр &amp; Управление обучением</p>
        </div>
      </div>
      <div className="header-status-indicator">
        <span className={`status-dot dot-${status}`} aria-hidden="true" />
        <span className="status-label">{getStatusLabel(status)}</span>
      </div>
    </header>
  );
};
