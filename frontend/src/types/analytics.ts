import { LessonFormat } from './schedule';

export interface AnalyticsOverview {
  from?: string;
  to?: string;
  total_lessons: number;
  completed_lessons: number;
  cancelled_lessons: number;
  completion_rate: number;
  completed_hours: number;
  gross_revenue: number;
  net_income: number;
  effective_hourly_rate: number;
}

export type AnalyticsOverviewResponse = AnalyticsOverview;

export interface AnalyticsDynamicsPoint {
  label: string;
  from: string;
  to: string;
  completed_hours: number;
  net_income: number;
  completed_count: number;
  cancelled_count: number;
}

export interface AnalyticsFormatStat {
  format: LessonFormat | string;
  completed_hours: number;
  net_income: number;
  lessons_count: number;
  hours_share_percent: number;
  revenue_share_percent: number;
}

export interface AnalyticsClientStat {
  client_id: string;
  client_name: string;
  completed_hours: number;
  net_income: number;
  completed_count: number;
  cancelled_count: number;
  attendance_rate: number;
}

export interface FormatForecast {
  format: LessonFormat | string;
  hours: number;
  revenue: number;
}

export interface AnalyticsForecast {
  from: string;
  to: string;
  scheduled_lessons: number;
  scheduled_hours: number;
  gross_potential_revenue: number;
  partner_commission_expected: number;
  net_potential_income: number;
  by_format: FormatForecast[];
}

export type AnalyticsForecastResponse = AnalyticsForecast;

export interface TagStat {
  tag_id: string;
  tag_name: string;
  tag_color?: string | null;
  students_count: number;
  completed_hours: number;
  gross_revenue: number;
  net_income: number;
}

export type TagStatResponse = TagStat;

export interface AnalyticsOverviewParams {
  from?: string;
  to?: string;
}

export interface AnalyticsDynamicsParams {
  interval?: 'week' | 'month';
  from?: string;
  to?: string;
}

export interface AnalyticsFormatsParams {
  from?: string;
  to?: string;
}

export interface AnalyticsClientsParams {
  from?: string;
  to?: string;
  sort?: 'hours' | 'revenue' | 'cancellations';
  limit?: number;
}

export interface AnalyticsForecastParams {
  from?: string;
  to?: string;
}

export interface AnalyticsTagsParams {
  from?: string;
  to?: string;
}
