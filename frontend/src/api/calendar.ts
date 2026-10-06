import { CalendarSettings, CalendarExportParams } from '../types/calendar';

const BASE_URL = '/api/v1';

const getHeaders = (contentType = true) => {
  const token = localStorage.getItem('school_access_token');
  return {
    ...(contentType ? { 'Content-Type': 'application/json' } : {}),
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
};

export async function getCalendarSettings(): Promise<CalendarSettings> {
  const res = await fetch(`${BASE_URL}/integrations/calendar/settings`, {
    headers: getHeaders(),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить настройки календаря');
  }

  return await res.json();
}

export async function rotateCalendarToken(): Promise<CalendarSettings> {
  const res = await fetch(`${BASE_URL}/integrations/calendar/rotate-token`, {
    method: 'POST',
    headers: getHeaders(),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось перевыпустить токен календаря');
  }

  return await res.json();
}

export async function downloadCalendarExport(
  params?: CalendarExportParams
): Promise<void> {
  const query = new URLSearchParams();
  if (params?.from) query.append('from', params.from);
  if (params?.to) query.append('to', params.to);
  const qs = query.toString() ? `?${query.toString()}` : '';

  const res = await fetch(`${BASE_URL}/integrations/calendar/export${qs}`, {
    headers: getHeaders(false),
  });

  if (!res.ok) {
    let errorMsg = 'Не удалось выгрузить календарь';
    try {
      const err = await res.json();
      if (err.error?.message) {
        errorMsg = err.error.message;
      }
    } catch {
      // not JSON
    }
    throw new Error(errorMsg);
  }

  const blob = await res.blob();
  const disposition = res.headers.get('Content-Disposition');
  let filename = `schedule_${new Date().toISOString().slice(0, 10)}.ics`;
  if (disposition) {
    const filenameMatch = disposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/);
    if (filenameMatch && filenameMatch[1]) {
      filename = filenameMatch[1].replace(/['"]/g, '');
    }
  }

  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.setAttribute('download', filename);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
}
