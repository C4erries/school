import {
  AnalyticsOverview,
  AnalyticsDynamicsPoint,
  AnalyticsFormatStat,
  AnalyticsClientStat,
  AnalyticsOverviewParams,
  AnalyticsDynamicsParams,
  AnalyticsFormatsParams,
  AnalyticsClientsParams,
} from '../types/analytics';

const BASE_URL = '/api/v1';

const getHeaders = () => {
  const token = localStorage.getItem('school_access_token');
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
};

export async function getAnalyticsOverview(
  params?: AnalyticsOverviewParams
): Promise<AnalyticsOverview> {
  const query = new URLSearchParams();
  if (params?.from) query.append('from', params.from);
  if (params?.to) query.append('to', params.to);

  const qs = query.toString() ? `?${query.toString()}` : '';
  const res = await fetch(`${BASE_URL}/analytics/overview${qs}`, {
    headers: getHeaders(),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить ключевые показатели');
  }

  return await res.json();
}

export async function getAnalyticsDynamics(
  params?: AnalyticsDynamicsParams
): Promise<AnalyticsDynamicsPoint[]> {
  const query = new URLSearchParams();
  if (params?.interval) query.append('interval', params.interval);
  if (params?.from) query.append('from', params.from);
  if (params?.to) query.append('to', params.to);

  const qs = query.toString() ? `?${query.toString()}` : '';
  const res = await fetch(`${BASE_URL}/analytics/dynamics${qs}`, {
    headers: getHeaders(),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить динамику показателей');
  }

  return await res.json();
}

export async function getAnalyticsFormats(
  params?: AnalyticsFormatsParams
): Promise<AnalyticsFormatStat[]> {
  const query = new URLSearchParams();
  if (params?.from) query.append('from', params.from);
  if (params?.to) query.append('to', params.to);

  const qs = query.toString() ? `?${query.toString()}` : '';
  const res = await fetch(`${BASE_URL}/analytics/formats${qs}`, {
    headers: getHeaders(),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить распределение по форматам');
  }

  return await res.json();
}

export async function getAnalyticsClients(
  params?: AnalyticsClientsParams
): Promise<AnalyticsClientStat[]> {
  const query = new URLSearchParams();
  if (params?.from) query.append('from', params.from);
  if (params?.to) query.append('to', params.to);
  if (params?.sort) query.append('sort', params.sort);
  if (params?.limit !== undefined) query.append('limit', String(params.limit));

  const qs = query.toString() ? `?${query.toString()}` : '';
  const res = await fetch(`${BASE_URL}/analytics/clients${qs}`, {
    headers: getHeaders(),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить рейтинг учеников');
  }

  return await res.json();
}

export async function getForecast(
  params?: { from?: string; to?: string }
): Promise<import('../types/analytics').AnalyticsForecast> {
  const query = new URLSearchParams();
  if (params?.from) query.append('from', params.from);
  if (params?.to) query.append('to', params.to);

  const qs = query.toString() ? `?${query.toString()}` : '';
  const res = await fetch(`${BASE_URL}/analytics/forecast${qs}`, {
    headers: getHeaders(),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить прогноз аналитики');
  }

  return await res.json();
}

export async function getTagStats(
  params?: { from?: string; to?: string }
): Promise<import('../types/analytics').TagStat[]> {
  const query = new URLSearchParams();
  if (params?.from) query.append('from', params.from);
  if (params?.to) query.append('to', params.to);

  const qs = query.toString() ? `?${query.toString()}` : '';
  const res = await fetch(`${BASE_URL}/analytics/tags${qs}`, {
    headers: getHeaders(),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить статистику по тегам');
  }

  return await res.json();
}
