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
