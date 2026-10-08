import {
  LessonJournal,
  HomeworkAssignment,
  LessonJournalBundle,
  UpsertLessonJournalInput,
  CreateHomeworkInput,
  UpdateHomeworkStatusInput,
  HomeworkStatus,
  StudyStream,
  ClientNote,
} from '../types/journal';

const BASE_URL = '/api/v1';

const getHeaders = () => {
  const token = localStorage.getItem('school_access_token');
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
};

/**
 * Получить бандл журнала и домашних заданий по уроку
 */
export async function getLessonJournalBundle(lessonId: string): Promise<LessonJournalBundle> {
  const res = await fetch(`${BASE_URL}/schedule/lessons/${lessonId}/journal`, {
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить журнал занятия');
  }
  return await res.json();
}

/**
 * Создать или обновить отчет по уроку
 */
export async function upsertLessonJournal(
  lessonId: string,
  data: UpsertLessonJournalInput
): Promise<LessonJournal> {
  const res = await fetch(`${BASE_URL}/schedule/lessons/${lessonId}/journal`, {
    method: 'PUT',
    headers: getHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось сохранить отчет по уроку');
  }
  return await res.json();
}

/**
 * Получить список всех отчетов по урокам ученика (хронологический таймлайн)
 */
export async function getClientJournals(clientId: string): Promise<LessonJournal[]> {
  const res = await fetch(`${BASE_URL}/crm/clients/${clientId}/journal`, {
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить историю занятий ученика');
  }
  return await res.json();
}

/**
 * Список всех домашних заданий ученика с опциональным фильтром по статусу
 */
export async function listClientHomework(
  clientId: string,
  status?: HomeworkStatus
): Promise<HomeworkAssignment[]> {
  const url = status
    ? `${BASE_URL}/crm/clients/${clientId}/homework?status=${encodeURIComponent(status)}`
    : `${BASE_URL}/crm/clients/${clientId}/homework`;

  const res = await fetch(url, {
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить домашние задания ученика');
  }
  return await res.json();
}

/**
 * Выдать домашнее задание ученику
 */
export async function createHomework(
  clientId: string,
  data: CreateHomeworkInput
): Promise<HomeworkAssignment> {
  const res = await fetch(`${BASE_URL}/crm/clients/${clientId}/homework`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось создать домашнее задание');
  }
  return await res.json();
}

/**
 * Обновить статус домашнего задания и рецензию
 */
export async function updateHomeworkStatus(
  homeworkId: string,
  data: UpdateHomeworkStatusInput
): Promise<HomeworkAssignment> {
  const res = await fetch(`${BASE_URL}/homework/${homeworkId}`, {
    method: 'PATCH',
    headers: getHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось обновить статус домашнего задания');
  }
  return await res.json();
}

/**
 * Удалить домашнее задание
 */
export async function deleteHomework(homeworkId: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/homework/${homeworkId}`, {
    method: 'DELETE',
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось удалить домашнее задание');
  }
}

/**
 * Получить единый хронологический поток обучения ученика (Study Stream)
 */
export async function getClientStudyStream(clientId: string): Promise<StudyStream> {
  const res = await fetch(`${BASE_URL}/crm/clients/${clientId}/stream`, {
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось загрузить поток обучения ученика');
  }
  return await res.json();
}

/**
 * Создать свободную быструю заметку по ученику
 */
export async function createClientNote(
  clientId: string,
  content: string
): Promise<ClientNote> {
  const res = await fetch(`${BASE_URL}/crm/clients/${clientId}/notes`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({ content }),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось сохранить заметку');
  }
  return await res.json();
}

/**
 * Удалить свободную заметку по ученику
 */
export async function deleteClientNote(noteId: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/crm/clients/notes/${noteId}`, {
    method: 'DELETE',
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось удалить заметку');
  }
}

