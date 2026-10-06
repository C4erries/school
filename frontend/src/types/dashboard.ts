import { LessonFormat } from './schedule';

export interface TodayLesson {
  id: string;
  client_id: string;
  client_name: string;
  start_at: string;
  end_at: string;
  format: LessonFormat | string;
  status: string;
  location_type?: string | null;
  online_link?: string | null;
  classroom_name?: string | null;
  classroom_color?: string | null;
}

export interface FinancialSnapshot {
  month_earned: number;
  month_forecast: number;
  total_debts: number;
  active_clients_count: number;
  weekly_hours: number;
}

export interface DashboardSummary {
  today_lessons: TodayLesson[];
  financial_snapshot: FinancialSnapshot;
}

export type DashboardSummaryResponse = DashboardSummary;
