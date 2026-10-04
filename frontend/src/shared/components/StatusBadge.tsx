import React from 'react';
import type { HealthStatus } from '@/types/health';

interface StatusBadgeProps {
  status: HealthStatus;
}

export const StatusBadge: React.FC<StatusBadgeProps> = ({ status }) => {
  const badgeMap: Record<HealthStatus, { text: string; className: string }> = {
    idle: { text: 'Standby', className: 'badge-idle' },
    checking: { text: 'Checking...', className: 'badge-checking' },
    healthy: { text: 'Healthy (200 OK)', className: 'badge-healthy' },
    unhealthy: { text: 'Connection Error', className: 'badge-unhealthy' },
  };

  const current = badgeMap[status];

  return <span className={`status-badge ${current.className}`}>{current.text}</span>;
};
