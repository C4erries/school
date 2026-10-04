import React, { useState, useEffect, useCallback } from 'react';
import { checkBackendHealth } from '@/api/health';
import type { HealthState } from '@/types/health';
import { Header } from '@/shared/components/Header';
import { StatusBadge } from '@/shared/components/StatusBadge';
import './App.css';

export const App: React.FC = () => {
  const [healthState, setHealthState] = useState<HealthState>({
    status: 'checking',
    data: null,
    error: null,
    lastCheckedAt: null,
    latencyMs: null,
  });
  const [autoPoll, setAutoPoll] = useState(true);

  const performHealthCheck = useCallback(async () => {
    setHealthState((prev) => ({ ...prev, status: 'checking', error: null }));
    try {
      const { data, latencyMs } = await checkBackendHealth();
      setHealthState({
        status: 'healthy',
        data,
        error: null,
        lastCheckedAt: new Date(),
        latencyMs,
      });
    } catch (err) {
      setHealthState((prev) => ({
        status: 'unhealthy',
        data: null,
        error: err instanceof Error ? err.message : 'Unknown connection error',
        lastCheckedAt: new Date(),
        latencyMs: null,
      }));
    }
  }, []);

  useEffect(() => {
    performHealthCheck();
  }, [performHealthCheck]);

  useEffect(() => {
    if (!autoPoll) return;
    const timer = setInterval(() => {
      performHealthCheck();
    }, 5000);
    return () => clearInterval(timer);
  }, [autoPoll, performHealthCheck]);

  return (
    <div className="app-container">
      <Header status={healthState.status} />

      <main className="app-main">
        <section className="card monitor-card">
          <div className="card-header">
            <div>
              <h2 className="card-title">Backend Connectivity Monitor</h2>
              <p className="card-description">
                Проверка соединения с Go Backend API через Nginx reverse proxy
              </p>
            </div>
            <StatusBadge status={healthState.status} />
          </div>

          <div className="metrics-grid">
            <div className="metric-box">
              <span className="metric-label">Target Endpoint</span>
              <span className="metric-value code">/api/v1/health</span>
            </div>

            <div className="metric-box">
              <span className="metric-label">Service</span>
              <span className="metric-value">
                {healthState.data?.service || (healthState.status === 'unhealthy' ? 'Offline' : '—')}
              </span>
            </div>

            <div className="metric-box">
              <span className="metric-label">Version</span>
              <span className="metric-value">
                {healthState.data?.version || (healthState.status === 'unhealthy' ? '—' : '—')}
              </span>
            </div>

            <div className="metric-box">
              <span className="metric-label">Response Time</span>
              <span className="metric-value">
                {healthState.latencyMs !== null ? `${healthState.latencyMs} ms` : '—'}
              </span>
            </div>
          </div>

          {healthState.error && (
            <div className="error-alert">
              <strong>Connection Error:</strong> {healthState.error}
              <p className="error-hint">
                Убедитесь, что контейнеры backend и nginx запущены (`docker compose ps`).
              </p>
            </div>
          )}

          <div className="card-footer">
            <div className="timestamp-info">
              {healthState.lastCheckedAt && (
                <span>Последняя проверка: {healthState.lastCheckedAt.toLocaleTimeString()}</span>
              )}
            </div>

            <div className="actions">
              <label className="toggle-label">
                <input
                  type="checkbox"
                  checked={autoPoll}
                  onChange={(e) => setAutoPoll(e.target.checked)}
                />
                Автопроверка (5с)
              </label>

              <button
                type="button"
                className="btn-refresh"
                onClick={performHealthCheck}
                disabled={healthState.status === 'checking'}
              >
                {healthState.status === 'checking' ? 'Проверка...' : 'Обновить'}
              </button>
            </div>
          </div>
        </section>

        <section className="card architecture-card">
          <h3 className="section-title">Схема взаимодействия (Phase 0)</h3>
          <div className="pipeline-flow">
            <div className="flow-step">
              <span className="step-title">Browser</span>
              <span className="step-detail">http://localhost/</span>
            </div>
            <div className="flow-arrow" aria-hidden="true">➔</div>
            <div className="flow-step">
              <span className="step-title">Nginx (:80)</span>
              <span className="step-detail">Reverse Proxy &amp; WebSocket</span>
            </div>
            <div className="flow-arrow" aria-hidden="true">➔</div>
            <div className="flow-step active">
              <span className="step-title">Backend (:8080)</span>
              <span className="step-detail">Go net/http (/api/v1/health)</span>
            </div>
          </div>
        </section>
      </main>

      <footer className="app-footer">
        <p>School Platform • Phase 0 (Foundation) • Go DDD + React 18 + Vite + Docker</p>
      </footer>
    </div>
  );
};

export default App;
