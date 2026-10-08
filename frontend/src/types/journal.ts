export type HomeworkStatus = 'assigned' | 'completed' | 'not_done';

export interface LessonJournal {
  id: string;
  lesson_id: string;
  client_id: string;
  teacher_id: string;
  topic: string;
  notes: string | null;
  performance_score: number | null;
  created_at: string;
  updated_at: string;
}

export interface HomeworkAssignment {
  id: string;
  client_id: string;
  teacher_id: string;
  assigned_lesson_id: string | null;
  title: string;
  description: string | null;
  due_date: string | null;
  status: HomeworkStatus;
  review_notes: string | null;
  created_at: string;
  updated_at: string;
}

export interface LessonJournalBundle {
  journal: LessonJournal | null;
  assigned_homeworks: HomeworkAssignment[];
  due_homeworks: HomeworkAssignment[];
}

export interface UpsertLessonJournalInput {
  topic: string;
  notes?: string | null;
  performance_score?: number | null;
}

export interface CreateHomeworkInput {
  title: string;
  description?: string | null;
  due_date?: string | null;
  assigned_lesson_id?: string | null;
}

export interface UpdateHomeworkStatusInput {
  status: HomeworkStatus;
  review_notes?: string | null;
}

export interface ClientNote {
  id: string;
  client_id: string;
  teacher_id: string;
  content: string;
  created_at: string;
  updated_at: string;
}

export interface UpcomingLessonInfo {
  lesson_id: string;
  start_time: string;
  end_time: string;
  format: string;
  title: string;
  topic?: string | null;
}

export interface StudyStreamItem {
  id: string;
  type: 'lesson_report' | 'note';
  timestamp: string;
  lesson_report?: LessonJournalBundle | null;
  note?: ClientNote | null;
}

export interface StudyStream {
  client_id: string;
  upcoming_lesson?: UpcomingLessonInfo | null;
  items: StudyStreamItem[];
}
