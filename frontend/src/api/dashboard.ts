import { DashboardSummary } from '../types/dashboard';

const BASE_URL = '/api/v1';

const getHeaders = () => {
  const token = localStorage.getItem('school_access_token');
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
};

export async function getDashboardSummary(): Promise<DashboardSummary> {
  const res = await fetch(`${BASE_URL}/dashboard/summary`, {
    headers: getHeaders(),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить сводку дашборда');
  }

  return await res.json();
}
