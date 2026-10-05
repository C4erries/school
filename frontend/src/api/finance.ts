import {
  FinanceSummaryResponse,
  PaymentResponse,
  CreatePaymentRequest,
  GetPaymentsParams,
  PartnerSettlementResponse,
  CreatePartnerPayoutRequest,
  PartnerPayoutResponse,
  ExportEntity,
  ExportParams,
} from '../types/finance';

const BASE_URL = '/api/v1';

const getHeaders = (contentType = true) => {
  const token = localStorage.getItem('school_access_token');
  return {
    ...(contentType ? { 'Content-Type': 'application/json' } : {}),
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
};

export async function getFinanceSummary(month?: string): Promise<FinanceSummaryResponse> {
  const query = month ? `?month=${encodeURIComponent(month)}` : '';
  const res = await fetch(`${BASE_URL}/finance/summary${query}`, {
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить финансовую сводку');
  }
  return await res.json();
}

export async function getPayments(params?: GetPaymentsParams): Promise<PaymentResponse[]> {
  const query = new URLSearchParams();
  if (params?.client_id) query.append('client_id', params.client_id);
  if (params?.from) query.append('from', params.from);
  if (params?.to) query.append('to', params.to);
  if (params?.limit !== undefined) query.append('limit', String(params.limit));
  if (params?.offset !== undefined) query.append('offset', String(params.offset));

  const qs = query.toString() ? `?${query.toString()}` : '';
  const res = await fetch(`${BASE_URL}/finance/payments${qs}`, {
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить журнал оплат');
  }
  return await res.json();
}

export async function createPayment(data: CreatePaymentRequest): Promise<PaymentResponse> {
  const res = await fetch(`${BASE_URL}/finance/payments`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось зарегистрировать оплату');
  }
  return await res.json();
}

export async function getPartnerSettlements(month?: string): Promise<PartnerSettlementResponse[]> {
  const query = month ? `?month=${encodeURIComponent(month)}` : '';
  const res = await fetch(`${BASE_URL}/finance/partner-settlements${query}`, {
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить расчеты со школами');
  }
  return await res.json();
}

export async function createPartnerPayout(
  data: CreatePartnerPayoutRequest
): Promise<PartnerPayoutResponse> {
  const res = await fetch(`${BASE_URL}/finance/partner-payouts`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось зафиксировать выплату партнеру');
  }
  return await res.json();
}

export async function downloadExport(
  entity: ExportEntity,
  params?: ExportParams
): Promise<void> {
  const query = new URLSearchParams();
  if (params?.from) query.append('from', params.from);
  if (params?.to) query.append('to', params.to);
  const qs = query.toString() ? `?${query.toString()}` : '';

  const res = await fetch(`${BASE_URL}/export/${entity}${qs}`, {
    headers: getHeaders(false),
  });

  if (!res.ok) {
    let errorMsg = 'Не удалось выгрузить данные';
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
  let filename = `${entity}_${new Date().toISOString().slice(0, 10)}.csv`;
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
