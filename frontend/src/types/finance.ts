export type PaymentMethod = 'transfer' | 'cash' | 'card' | 'other';

export interface FinanceSummary {
  month: string;
  total_payments: number;
  total_earned: number;
  total_debts: number;
  total_commissions: number;
  active_subscriptions_count: number;
  debtors_count: number;
}

export type FinanceSummaryResponse = FinanceSummary;

export interface CreatePaymentRequest {
  client_id: string;
  amount: number;
  hours: number;
  format: 'individual' | 'pair' | 'group';
  payment_method?: PaymentMethod;
  paid_at?: string | null;
  notes?: string | null;
}

export interface Payment {
  id: string;
  teacher_id: string;
  client_id: string;
  client_name?: string | null;
  amount: number;
  hours: number;
  format: 'individual' | 'pair' | 'group';
  payment_method: PaymentMethod;
  paid_at: string;
  notes?: string | null;
  created_at: string;
}

export type PaymentResponse = Payment;

export interface GetPaymentsParams {
  client_id?: string;
  from?: string;
  to?: string;
  limit?: number;
  offset?: number;
}

export interface PartnerSettlement {
  tag_id: string;
  tag_name: string;
  tag_color?: string | null;
  school_percent: number;
  period_month: string;
  lessons_count: number;
  gross_amount: number;
  commission_amount: number;
  is_paid: boolean;
  paid_at?: string | null;
  payout_id?: string | null;
}

export type PartnerSettlementResponse = PartnerSettlement;

export interface CreatePartnerPayoutRequest {
  tag_id: string;
  period_month: string;
  gross_amount: number;
  commission_amount: number;
  paid_at?: string | null;
  notes?: string | null;
}

export interface PartnerPayout {
  id: string;
  teacher_id: string;
  tag_id: string;
  period_month: string;
  gross_amount: number;
  commission_amount: number;
  paid_at: string;
  notes?: string | null;
  created_at: string;
}

export type PartnerPayoutResponse = PartnerPayout;

export type ExportEntity = 'clients' | 'lessons' | 'payments';

export interface ExportParams {
  from?: string;
  to?: string;
}
