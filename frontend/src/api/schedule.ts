import {
  Classroom,
  Client,
  Lesson,
  TeacherStudent,
  Tag,
  CreateClassroomRequest,
  CreateClientRequest,
  UpdateClientRequest,
  AddSubscriptionRequest,
  CreateTagRequest,
  CreateLessonRequest,
  UpdateLessonRequest,
  FinancialDashboardStats,
  DashboardMetrics,
  UserDefaultRates,
  AdjustBalanceRequest,
} from '../types/schedule';
import { User, Role } from '../types/auth';

const BASE_URL = '/api/v1';

const getHeaders = () => {
  const token = localStorage.getItem('school_access_token');
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
};

// ---------------------------------------------------------------------------
// 1. Кабинеты (Classrooms)
// ---------------------------------------------------------------------------

export async function getClassrooms(): Promise<Classroom[]> {
  const res = await fetch(`${BASE_URL}/classrooms`, { headers: getHeaders() });
  if (!res.ok) {
    throw new Error('Не удалось загрузить список кабинетов');
  }
  return await res.json();
}

export async function createClassroom(data: CreateClassroomRequest): Promise<Classroom> {
  const res = await fetch(`${BASE_URL}/classrooms`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось создать кабинет');
  }
  return await res.json();
}

export async function deleteClassroom(id: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/classrooms/${id}`, {
    method: 'DELETE',
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось удалить кабинет');
  }
}

// ---------------------------------------------------------------------------
// 2. Привязки Ученик <-> Преподаватель (Teacher-Student Relations)
// ---------------------------------------------------------------------------

export async function getTeacherStudents(teacherId?: string): Promise<TeacherStudent[]> {
  const url = teacherId
    ? `${BASE_URL}/teachers/${teacherId}/students`
    : `${BASE_URL}/teachers/students`;
  const res = await fetch(url, { headers: getHeaders() });
  if (!res.ok) {
    throw new Error('Не удалось загрузить список привязок');
  }
  return await res.json();
}

export async function assignStudentToTeacher(
  teacherId: string,
  studentId: string
): Promise<TeacherStudent> {
  const res = await fetch(`${BASE_URL}/teachers/students`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({ teacher_id: teacherId, student_id: studentId }),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось привязать ученика');
  }
  return await res.json();
}

export async function removeTeacherStudent(id: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/teachers/students/${id}`, {
    method: 'DELETE',
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось открепить ученика');
  }
}

export async function getUsersByRole(role: Role): Promise<User[]> {
  const res = await fetch(`${BASE_URL}/users?role=${role}`, {
    method: 'GET',
    headers: getHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Не удалось загрузить пользователей с ролью ${role}`);
  }
  return await res.json();
}

// ---------------------------------------------------------------------------
// 3. Динамические Теги (Tags)
// ---------------------------------------------------------------------------

const SESSION_TAGS_KEY = 'school_dynamic_tags';

export async function getTags(): Promise<Tag[]> {
  try {
    const res = await fetch(`${BASE_URL}/tags`, { headers: getHeaders() });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    // If backend tag endpoint is not deployed yet, fallback to session storage
  }

  try {
    const raw = sessionStorage.getItem(SESSION_TAGS_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

export async function createTag(data: CreateTagRequest): Promise<Tag> {
  try {
    const res = await fetch(`${BASE_URL}/tags`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify(data),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    // Fallback for seamless UX
  }

  const newTag: Tag = {
    id: `tag-${Date.now()}`,
    name: data.name,
    school_percent: data.school_percent,
    color: data.color || '#4F46E5',
    created_at: new Date().toISOString(),
  };

  try {
    const existing = await getTags();
    const updated = [...existing, newTag];
    sessionStorage.setItem(SESSION_TAGS_KEY, JSON.stringify(updated));
  } catch {
    // Ignore
  }

  return newTag;
}

// ---------------------------------------------------------------------------
// 4. Клиенты и Абонементы (Clients & Format Subscriptions)
// ---------------------------------------------------------------------------

export async function getClients(params?: {
  is_archived?: boolean;
  search?: string;
}): Promise<Client[]> {
  const query = new URLSearchParams();
  if (params?.is_archived !== undefined) {
    query.set('is_archived', String(params.is_archived));
  }
  if (params?.search) {
    query.set('search', params.search);
  }
  const queryString = query.toString() ? `?${query.toString()}` : '';

  const res = await fetch(`${BASE_URL}/clients${queryString}`, { headers: getHeaders() });
  if (!res.ok) {
    throw new Error('Не удалось загрузить клиентов');
  }
  const rawList: (Record<string, unknown> & Partial<Client>)[] = await res.json();
  return rawList.map((c): Client => {
    const indRate = (c.rate_individual as number | undefined) ?? (c.base_rate as number | undefined) ?? 0;
    const pairRate = (c.rate_pair as number | null | undefined) ?? null;
    const grpRate = (c.rate_group as number | null | undefined) ?? null;
    const balances = c.balances ?? {
      individual_hours: c.balance ?? 0,
      pair_hours: 0,
      group_hours: 0,
      total_hours: c.balance ?? 0,
    };
    return {
      ...(c as unknown as Client),
      rate_individual: indRate,
      rate_pair: pairRate,
      rate_group: grpRate,
      base_rate: indRate,
      balance: balances.total_hours ?? c.balance ?? 0,
      balances,
      tags: c.tags ?? (c.tag ? [{ id: 'legacy-tag', name: c.tag, school_percent: c.school_percent_tag ?? 0 }] : []),
      is_archived: Boolean(c.is_archived),
    };
  });
}

export async function createClient(data: CreateClientRequest): Promise<Client> {
  const payload = {
    name: data.name,
    phone: data.phone || null,
    base_rate: data.rate_individual,
    rate_individual: data.rate_individual,
    rate_pair: data.rate_pair ?? null,
    rate_group: data.rate_group ?? null,
    school_percent_tag: data.school_percent_tag ?? 0,
    tag: data.tag || null,
    tag_ids: data.tag_ids || [],
  };

  const res = await fetch(`${BASE_URL}/clients`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось создать клиента');
  }

  const created = (await res.json()) as Client;
  const indRate = created.rate_individual ?? created.base_rate ?? data.rate_individual;
  return {
    ...created,
    rate_individual: indRate,
    rate_pair: created.rate_pair ?? data.rate_pair ?? null,
    rate_group: created.rate_group ?? data.rate_group ?? null,
    base_rate: indRate,
    balance: created.balance ?? 0,
    balances: created.balances ?? {
      individual_hours: 0,
      pair_hours: 0,
      group_hours: 0,
      total_hours: 0,
    },
  };
}

export async function updateClient(id: string, data: UpdateClientRequest): Promise<Client> {
  const res = await fetch(`${BASE_URL}/clients/${id}`, {
    method: 'PATCH',
    headers: getHeaders(),
    body: JSON.stringify(data),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось обновить данные клиента');
  }

  const updated = (await res.json()) as Client;
  const indRate = (updated.rate_individual as number | undefined) ?? (updated.base_rate as number | undefined) ?? (data.rate_individual ?? 0);
  const pairRate = (updated.rate_pair as number | null | undefined) ?? (data.rate_pair !== undefined ? data.rate_pair : null);
  const grpRate = (updated.rate_group as number | null | undefined) ?? (data.rate_group !== undefined ? data.rate_group : null);
  const balances = updated.balances ?? {
    individual_hours: updated.balance ?? 0,
    pair_hours: 0,
    group_hours: 0,
    total_hours: updated.balance ?? 0,
  };
  return {
    ...updated,
    rate_individual: indRate,
    rate_pair: pairRate,
    rate_group: grpRate,
    base_rate: indRate,
    balance: balances.total_hours ?? updated.balance ?? 0,
    balances,
    tags: updated.tags ?? [],
  };
}

export async function addSubscription(data: AddSubscriptionRequest): Promise<void> {
  const hours = data.hours ?? data.amount ?? data.balance ?? 1;
  const format = data.format || 'individual';
  const payload = {
    format,
    type: 'hours',
    balance: hours,
    hours,
  };

  const res = await fetch(`${BASE_URL}/clients/${data.client_id}/subscriptions`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось пополнить абонемент');
  }
}

// ---------------------------------------------------------------------------
// 5. Уроки и Расписание (Lessons)
// ---------------------------------------------------------------------------

export async function getLessons(params?: {
  teacher_id?: string;
  client_id?: string;
  student_id?: string;
  date?: string;
  from?: string;
  to?: string;
}): Promise<Lesson[]> {
  const query = new URLSearchParams();
  if (params?.teacher_id) query.set('teacher_id', params.teacher_id);
  if (params?.client_id) query.set('client_id', params.client_id);
  if (params?.student_id) query.set('student_id', params.student_id);
  if (params?.date) query.set('date', params.date);
  if (params?.from) query.set('from', params.from);
  if (params?.to) query.set('to', params.to);

  const queryString = query.toString() ? `?${query.toString()}` : '';
  const res = await fetch(`${BASE_URL}/lessons${queryString}`, {
    method: 'GET',
    headers: getHeaders(),
  });

  if (!res.ok) {
    throw new Error('Не удалось загрузить уроки');
  }

  const data: Lesson[] = await res.json();
  return data.map((l) => ({
    ...l,
    title: l.title || l.notes || 'Занятие',
    online_link: l.location_or_url || l.online_link || undefined,
    classroom_name: l.classroom_name || (l.format === 'offline' ? l.location_or_url : undefined),
    comment: l.comment || l.notes,
    decline_reason: l.decline_reason || l.cancel_reason,
  }));
}

export async function createLesson(data: CreateLessonRequest): Promise<Lesson> {
  const targetId = data.client_id || data.student_id || '';
  const payload = {
    client_id: targetId,
    student_id: targetId,
    classroom_id: data.classroom_id || null,
    title: data.title,
    format: data.format,
    start_time: data.start_time,
    end_time: data.end_time,
    location_or_url: data.online_link || null,
    notes: data.title + (data.comment ? ` (${data.comment})` : ''),
  };

  const res = await fetch(`${BASE_URL}/lessons`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось создать урок');
  }

  const created: Lesson = await res.json();
  return {
    ...created,
    title: created.title || created.notes || data.title,
    online_link: created.format === 'online' ? (created.location_or_url || data.online_link) : undefined,
  };
}

export async function updateLesson(lessonId: string, data: UpdateLessonRequest): Promise<Lesson> {
  const payload: Record<string, unknown> = {};
  if (data.title !== undefined) payload.title = data.title;
  if (data.classroom_id !== undefined) payload.classroom_id = data.classroom_id;
  if (data.start_time !== undefined) payload.start_time = data.start_time;
  if (data.end_time !== undefined) payload.end_time = data.end_time;
  if (data.format !== undefined) payload.format = data.format;
  if (data.online_link !== undefined) payload.location_or_url = data.online_link;
  if (data.comment !== undefined) payload.notes = data.comment;

  const res = await fetch(`${BASE_URL}/lessons/${lessonId}`, {
    method: 'PATCH',
    headers: getHeaders(),
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось обновить занятие');
  }

  return await res.json();
}

export async function completeLesson(lessonId: string): Promise<Lesson> {
  const res = await fetch(`${BASE_URL}/lessons/${lessonId}/complete`, {
    method: 'POST',
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось завершить урок');
  }
  return await res.json();
}

export async function markNoShow(lessonId: string): Promise<Lesson> {
  const res = await fetch(`${BASE_URL}/lessons/${lessonId}/no-show`, {
    method: 'POST',
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось отметить неявку');
  }
  return await res.json();
}

export async function acceptLesson(lessonId: string): Promise<Lesson> {
  const res = await fetch(`${BASE_URL}/lessons/${lessonId}/accept`, {
    method: 'POST',
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось подтвердить урок');
  }
  return await res.json();
}

export async function declineLesson(lessonId: string, reason: string): Promise<Lesson> {
  const res = await fetch(`${BASE_URL}/lessons/${lessonId}/decline`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({ reason }),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось отклонить урок');
  }
  return await res.json();
}

export async function cancelLesson(lessonId: string, reason?: string): Promise<Lesson> {
  const res = await fetch(`${BASE_URL}/lessons/${lessonId}/cancel`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({ reason }),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось отменить урок');
  }
  return await res.json();
}

// ---------------------------------------------------------------------------
// 6. Финансовый Дашборд (Dashboard Metrics)
// ---------------------------------------------------------------------------

export async function getDashboardMetrics(from: string, to: string): Promise<DashboardMetrics> {
  try {
    const res = await fetch(
      `${BASE_URL}/dashboard/metrics?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
      { headers: getHeaders() }
    );
    if (res.ok) {
      return await res.json();
    }
  } catch {
    // Handled below
  }

  return {
    gross_potential_revenue: 0,
    net_income: 0,
    average_rate: 0,
  };
}

export async function getFinancialDashboard(): Promise<FinancialDashboardStats> {
  const now = new Date();
  const from = new Date(now.getFullYear(), now.getMonth(), 1).toISOString();
  const to = new Date(now.getFullYear(), now.getMonth() + 1, 0, 23, 59, 59).toISOString();

  try {
    const res = await fetch(
      `${BASE_URL}/dashboard/metrics?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
      { headers: getHeaders() }
    );
    if (res.ok) {
      const data: DashboardMetrics = await res.json();
      return {
        gross_potential_revenue: data.gross_potential_revenue,
        gross_revenue: data.gross_potential_revenue,
        net_income: data.net_income,
        average_rate: data.average_rate,
        average_hourly_rate: data.average_rate,
      };
    }
  } catch {
    // Handled below
  }

  return {
    gross_potential_revenue: 0,
    gross_revenue: 0,
    net_income: 0,
    average_rate: 0,
    average_hourly_rate: 0,
  };
}

// ---------------------------------------------------------------------------
// 7. Дефолтные ставки преподавателя (Tutor Default Rates)
// ---------------------------------------------------------------------------

export async function getDefaultRates(): Promise<UserDefaultRates> {
  const res = await fetch(`${BASE_URL}/users/me/rates`, { headers: getHeaders() });
  if (!res.ok) {
    throw new Error('Не удалось загрузить базовые ставки');
  }
  return await res.json();
}

export async function updateDefaultRates(rates: UserDefaultRates): Promise<UserDefaultRates> {
  const res = await fetch(`${BASE_URL}/users/me/rates`, {
    method: 'PUT',
    headers: getHeaders(),
    body: JSON.stringify(rates),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось обновить базовые ставки');
  }
  return await res.json();
}

// ---------------------------------------------------------------------------
// 8. Архивация и ручная корректировка баланса клиента (Archive & Adjust Balance)
// ---------------------------------------------------------------------------

export async function archiveClient(id: string): Promise<Client> {
  const res = await fetch(`${BASE_URL}/clients/${id}/archive`, {
    method: 'POST',
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось архивировать ученика');
  }
  return await res.json();
}

export async function unarchiveClient(id: string): Promise<Client> {
  const res = await fetch(`${BASE_URL}/clients/${id}/unarchive`, {
    method: 'POST',
    headers: getHeaders(),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось восстановить ученика из архива');
  }
  return await res.json();
}

export async function adjustClientBalance(
  id: string,
  data: AdjustBalanceRequest
): Promise<Client> {
  const res = await fetch(`${BASE_URL}/clients/${id}/adjust-balance`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error?.message || 'Не удалось скорректировать баланс');
  }
  return await res.json();
}
