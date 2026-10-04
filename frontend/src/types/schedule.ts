import { User } from './auth';

export type LessonStatus =
  | 'pending_confirmation'
  | 'confirmed'
  | 'declined'
  | 'completed'
  | 'no_show'
  | 'cancelled'
  | 'cancelled_by_teacher'
  | 'cancelled_by_student';

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

export interface Lesson {
  id: string;
  teacher_id: string;
  student_id: string;
  classroom_id?: string | null;
  title: string;
  format: LessonFormat;
  status: LessonStatus;
  start_time: string; // ISO 8601
  end_time: string;   // ISO 8601
  location_or_url?: string | null;
  notes?: string | null;
  cancel_reason?: string | null;
  online_link?: string | null;
  decline_reason?: string | null;
  comment?: string | null;
  teacher?: User;
  student?: User;
  classroom?: Classroom;
  teacher_name?: string;
  student_name?: string;
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

export interface CreateLessonRequest {
  student_id: string;
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
