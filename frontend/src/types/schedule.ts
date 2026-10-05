import { User } from './auth';

export type LessonStatus =
  | 'scheduled'
  | 'pending_confirmation'
  | 'confirmed'
  | 'declined'
  | 'completed'
  | 'cancelled'
  | 'cancelled_by_teacher'
  | 'cancelled_by_student'
  | 'no_show';

export type LessonFormat = 'online' | 'offline';

export interface Classroom {
  id: string;
  name: string;
  capacity: number;
  color: string;
  description?: string;
  created_at?: string;
}

export interface TeacherStudent {
  id: string;
  teacher_id: string;
  student_id: string;
  teacher?: User;
  student?: User;
  teacher_name?: string;
  student_name?: string;
  created_at: string;
}

export interface Client {
  id: string;
  teacher_id?: string;
  user_id?: string;
  name: string;
  phone?: string | null;
  base_rate: number;
  school_percent_tag?: number;
  tag?: string | null;
  balance: number;
  created_at: string;
}

export interface ClientSubscription {
  id: string;
  client_id: string;
  type?: 'lessons' | 'hours' | string;
  balance?: number;
  amount?: number;
  created_at: string;
}

export interface Lesson {
  id: string;
  teacher_id: string;
  client_id?: string;
  student_id?: string;
  classroom_id?: string | null;
  title: string;
  format: LessonFormat;
  status: LessonStatus;
  start_time: string; // ISO 8601
  end_time: string;   // ISO 8601
  location_or_url?: string | null;
  notes?: string | null;
  online_link?: string | null;
  comment?: string | null;
  cancel_reason?: string | null;
  decline_reason?: string | null;
  teacher?: User;
  student?: User;
  client?: Client;
  classroom?: Classroom;
  teacher_name?: string;
  student_name?: string;
  client_name?: string;
  classroom_name?: string | null;
  classroom_color?: string | null;
  created_at?: string;
  updated_at?: string;
}

export interface CreateClassroomRequest {
  name: string;
  capacity: number;
  color: string;
  description?: string;
}

export interface AssignStudentRequest {
  teacher_id: string;
  student_id: string;
}

export interface CreateClientRequest {
  name: string;
  phone?: string | null;
  base_rate: number;
  school_percent_tag?: number;
  tag?: string;
}

export interface AddSubscriptionRequest {
  client_id: string;
  type?: 'lessons' | 'hours';
  balance?: number;
  amount?: number;
}

export interface CreateSubscriptionRequest {
  type: 'lessons' | 'hours';
  balance: number;
}

export interface CreateLessonRequest {
  client_id?: string;
  student_id?: string;
  classroom_id?: string | null;
  title: string;
  format: LessonFormat;
  start_time: string;
  end_time: string;
  online_link?: string;
  comment?: string;
}

export interface DeclineLessonRequest {
  reason: string;
}

export interface CancelLessonRequest {
  reason?: string;
}

export interface DashboardMetrics {
  gross_potential_revenue: number;
  net_income: number;
  average_rate: number;
}

export interface FinancialDashboardStats {
  gross_revenue: number;
  net_income: number;
  average_hourly_rate: number;
  gross_potential_revenue?: number;
  average_rate?: number;
}
