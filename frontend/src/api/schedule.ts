import {
  Classroom,
  Client,
  Lesson,
  TeacherStudent,
  CreateClassroomRequest,
  CreateClientRequest,
  AddSubscriptionRequest,
  CreateLessonRequest,
  FinancialDashboardStats,
  DashboardMetrics,
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

const STORAGE_KEY_CLASSROOMS = 'school_mock_classrooms';
const STORAGE_KEY_RELATIONS = 'school_mock_relations';
const STORAGE_KEY_CLIENTS = 'school_mock_clients';
const STORAGE_KEY_LESSONS = 'school_mock_lessons';

const defaultClassrooms: Classroom[] = [
  {
    id: 'room-101',
    name: 'Кабинет 101 — Квант',
    capacity: 4,
    color: '#4F46E5',
    description: 'Интерактивная доска, микроскоп',
    created_at: new Date().toISOString(),
  },
  {
    id: 'room-102',
    name: 'Кабинет 102 — Тесла',
    capacity: 6,
    color: '#10B981',
    description: 'Компьютерный класс с проектором',
    created_at: new Date().toISOString(),
  },
  {
    id: 'room-103',
    name: 'Кабинет 103 — Архимед',
    capacity: 2,
    color: '#F59E0B',
    description: 'Индивидуальные занятия и репетиторство',
    created_at: new Date().toISOString(),
  },
];

const mockUsers: User[] = [
  {
    id: 'teacher-1',
    email: 'teacher@school.ru',
    full_name: 'Александр Верников',
    role: 'teacher',
    created_at: new Date().toISOString(),
  },
  {
    id: 'teacher-2',
    email: 'elena@school.ru',
    full_name: 'Елена Соколова',
    role: 'teacher',
    created_at: new Date().toISOString(),
  },
  {
    id: 'student-1',
    email: 'ivan@school.ru',
    full_name: 'Иван Смирнов',
    role: 'student',
    created_at: new Date().toISOString(),
  },
  {
    id: 'student-2',
    email: 'anna@school.ru',
    full_name: 'Анна Кузнецова',
    role: 'student',
    created_at: new Date().toISOString(),
  },
  {
    id: 'student-3',
    email: 'dmitry@school.ru',
    full_name: 'Дмитрий Попов',
    role: 'student',
    created_at: new Date().toISOString(),
  },
];

const defaultRelations: TeacherStudent[] = [
  {
    id: 'ts-1',
    teacher_id: 'teacher-1',
    student_id: 'student-1',
    teacher_name: 'Александр Верников',
    student_name: 'Иван Смирнов',
    created_at: new Date().toISOString(),
  },
  {
    id: 'ts-2',
    teacher_id: 'teacher-1',
    student_id: 'student-2',
    teacher_name: 'Александр Верников',
    student_name: 'Анна Кузнецова',
    created_at: new Date().toISOString(),
  },
  {
    id: 'ts-3',
    teacher_id: 'teacher-2',
    student_id: 'student-3',
    teacher_name: 'Елена Соколова',
    student_name: 'Дмитрий Попов',
    created_at: new Date().toISOString(),
  },
];

const defaultClients: Client[] = [
  {
    id: 'client-1',
    teacher_id: 'teacher-1',
    user_id: 'teacher-1',
    name: 'Иван Смирнов',
    phone: '+7 999 123 45 67',
    base_rate: 1500,
    school_percent_tag: 5,
    tag: '5% комиссия школе',
    balance: 5,
    created_at: new Date().toISOString(),
  },
  {
    id: 'client-2',
    teacher_id: 'teacher-1',
    user_id: 'teacher-1',
    name: 'Анна Кузнецова',
    phone: '+7 999 765 43 21',
    base_rate: 1800,
    school_percent_tag: 10,
    tag: '10% школе',
    balance: 2,
    created_at: new Date().toISOString(),
  },
];

const getStored = <T>(key: string, defaults: T): T => {
  try {
    const raw = localStorage.getItem(key);
    if (!raw) {
      localStorage.setItem(key, JSON.stringify(defaults));
      return defaults;
    }
    return JSON.parse(raw);
  } catch {
    return defaults;
  }
};

const setStored = <T>(key: string, data: T): void => {
  try {
    localStorage.setItem(key, JSON.stringify(data));
  } catch (e) {
    console.error('Failed to save to localStorage', e);
  }
};

const initMockLessons = (): Lesson[] => {
  const today = new Date();
  const y = today.getFullYear();
  const m = String(today.getMonth() + 1).padStart(2, '0');
  const d = String(today.getDate()).padStart(2, '0');
  const dateStr = `${y}-${m}-${d}`;

  return [
    {
      id: 'lesson-1',
      teacher_id: 'teacher-1',
      client_id: 'client-1',
      student_id: 'student-1',
      classroom_id: 'room-101',
      title: 'Математический анализ: Пределы и производные',
      format: 'offline',
      status: 'scheduled',
      start_time: `${dateStr}T14:00:00`,
      end_time: `${dateStr}T15:30:00`,
      client_name: 'Иван Смирнов',
      student_name: 'Иван Смирнов',
      teacher_name: 'Александр Верников',
      classroom_name: 'Кабинет 101 — Квант',
      classroom_color: '#4F46E5',
      created_at: new Date().toISOString(),
    },
    {
      id: 'lesson-2',
      teacher_id: 'teacher-1',
      client_id: 'client-2',
      student_id: 'student-2',
      classroom_id: null,
      title: 'Подготовка к ЕГЭ: Геометрия (часть 2)',
      format: 'online',
      status: 'scheduled',
      start_time: `${dateStr}T15:00:00`,
      end_time: `${dateStr}T16:00:00`,
      online_link: 'https://telemost.yandex.ru/j/school-math-102',
      client_name: 'Анна Кузнецова',
      student_name: 'Анна Кузнецова',
      teacher_name: 'Александр Верников',
      classroom_color: '#10B981',
      created_at: new Date().toISOString(),
    },
  ];
};

// ---------------------------------------------------------------------------
// 1. Кабинеты (Classrooms)
// ---------------------------------------------------------------------------

export async function getClassrooms(): Promise<Classroom[]> {
  try {
    const res = await fetch(`${BASE_URL}/classrooms`, { headers: getHeaders() });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    void 0;
  }
  return getStored<Classroom[]>(STORAGE_KEY_CLASSROOMS, defaultClassrooms);
}

export async function createClassroom(data: CreateClassroomRequest): Promise<Classroom> {
  try {
    const res = await fetch(`${BASE_URL}/classrooms`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify(data),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    void 0;
  }

  const existing = getStored<Classroom[]>(STORAGE_KEY_CLASSROOMS, defaultClassrooms);
  const newRoom: Classroom = {
    id: `room-${Date.now()}`,
    name: data.name,
    capacity: data.capacity,
    color: data.color,
    description: data.description,
    created_at: new Date().toISOString(),
  };
  setStored(STORAGE_KEY_CLASSROOMS, [newRoom, ...existing]);
  return newRoom;
}

export async function deleteClassroom(id: string): Promise<void> {
  try {
    const res = await fetch(`${BASE_URL}/classrooms/${id}`, {
      method: 'DELETE',
      headers: getHeaders(),
    });
    if (res.ok) return;
  } catch {
    void 0;
  }

  const existing = getStored<Classroom[]>(STORAGE_KEY_CLASSROOMS, defaultClassrooms);
  setStored(STORAGE_KEY_CLASSROOMS, existing.filter((r) => r.id !== id));
}

// ---------------------------------------------------------------------------
// 2. Привязки Ученик <-> Преподаватель (Teacher-Student Relations)
// ---------------------------------------------------------------------------

export async function getTeacherStudents(teacherId?: string): Promise<TeacherStudent[]> {
  try {
    const url = teacherId
      ? `${BASE_URL}/teachers/${teacherId}/students`
      : `${BASE_URL}/teachers/students`;
    const res = await fetch(url, { headers: getHeaders() });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    void 0;
  }

  const all = getStored<TeacherStudent[]>(STORAGE_KEY_RELATIONS, defaultRelations);
  if (teacherId) {
    return all.filter((r) => r.teacher_id === teacherId);
  }
  return all;
}

export async function assignStudentToTeacher(teacherId: string, studentId: string): Promise<TeacherStudent> {
  try {
    const res = await fetch(`${BASE_URL}/teachers/students`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ teacher_id: teacherId, student_id: studentId }),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    void 0;
  }

  const all = getStored<TeacherStudent[]>(STORAGE_KEY_RELATIONS, defaultRelations);
  const existing = all.find((r) => r.teacher_id === teacherId && r.student_id === studentId);
  if (existing) {
    return existing;
  }

  const teacher = mockUsers.find((u) => u.id === teacherId);
  const student = mockUsers.find((u) => u.id === studentId);

  const newRelation: TeacherStudent = {
    id: `ts-${Date.now()}`,
    teacher_id: teacherId,
    student_id: studentId,
    teacher_name: teacher?.full_name || 'Преподаватель',
    student_name: student?.full_name || 'Ученик',
    created_at: new Date().toISOString(),
  };

  const updated = [newRelation, ...all];
  setStored(STORAGE_KEY_RELATIONS, updated);
  return newRelation;
}

export async function removeTeacherStudent(id: string): Promise<void> {
  try {
    const res = await fetch(`${BASE_URL}/teachers/students/${id}`, {
      method: 'DELETE',
      headers: getHeaders(),
    });
    if (res.ok) return;
  } catch {
    void 0;
  }

  const all = getStored<TeacherStudent[]>(STORAGE_KEY_RELATIONS, defaultRelations);
  setStored(
    STORAGE_KEY_RELATIONS,
    all.filter((r) => r.id !== id)
  );
}

export async function getUsersByRole(role: Role): Promise<User[]> {
  try {
    const res = await fetch(`${BASE_URL}/users?role=${role}`, {
      method: 'GET',
      headers: getHeaders(),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    void 0;
  }

  return mockUsers.filter((u) => u.role === role);
}

// ---------------------------------------------------------------------------
// 3. Клиенты и Абонементы (Clients & Subscriptions)
// ---------------------------------------------------------------------------

export async function getClients(): Promise<Client[]> {
  try {
    const res = await fetch(`${BASE_URL}/clients`, { headers: getHeaders() });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    void 0;
  }
  return getStored<Client[]>(STORAGE_KEY_CLIENTS, defaultClients);
}

export async function createClient(data: CreateClientRequest): Promise<Client> {
  const percentMatch = data.tag ? data.tag.match(/(\d+)/) : null;
  const parsedPercent = percentMatch ? parseInt(percentMatch[1], 10) : 0;
  const payload = {
    name: data.name,
    phone: data.phone || null,
    base_rate: data.base_rate,
    school_percent_tag: data.school_percent_tag ?? parsedPercent,
  };

  try {
    const res = await fetch(`${BASE_URL}/clients`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify(payload),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    void 0;
  }

  const existing = getStored<Client[]>(STORAGE_KEY_CLIENTS, defaultClients);
  const newClient: Client = {
    id: `client-${Date.now()}`,
    teacher_id: 'teacher-1',
    user_id: 'teacher-1',
    name: data.name,
    phone: data.phone || null,
    base_rate: data.base_rate,
    school_percent_tag: data.school_percent_tag ?? parsedPercent,
    tag: data.tag || (parsedPercent > 0 ? `${parsedPercent}% школе` : null),
    balance: 0,
    created_at: new Date().toISOString(),
  };
  setStored(STORAGE_KEY_CLIENTS, [newClient, ...existing]);
  return newClient;
}

export async function addSubscription(data: AddSubscriptionRequest): Promise<void> {
  const amount = data.amount ?? data.balance ?? 1;
  const payload = {
    type: data.type || 'lessons',
    balance: amount,
  };

  try {
    const res = await fetch(`${BASE_URL}/clients/${data.client_id}/subscriptions`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify(payload),
    });
    if (res.ok) return;
  } catch {
    void 0;
  }

  const existing = getStored<Client[]>(STORAGE_KEY_CLIENTS, defaultClients);
  setStored(
    STORAGE_KEY_CLIENTS,
    existing.map((c) =>
      c.id === data.client_id ? { ...c, balance: c.balance + amount } : c
    )
  );
}

// ---------------------------------------------------------------------------
// 4. Уроки и Расписание (Lessons)
// ---------------------------------------------------------------------------

export async function getLessons(params?: {
  teacher_id?: string;
  client_id?: string;
  student_id?: string;
  date?: string;
}): Promise<Lesson[]> {
  try {
    const query = new URLSearchParams();
    if (params?.teacher_id) query.set('teacher_id', params.teacher_id);
    if (params?.client_id) query.set('client_id', params.client_id);
    if (params?.student_id) query.set('student_id', params.student_id);
    if (params?.date) query.set('date', params.date);

    const queryString = query.toString() ? `?${query.toString()}` : '';
    const res = await fetch(`${BASE_URL}/lessons${queryString}`, {
      method: 'GET',
      headers: getHeaders(),
    });
    if (res.ok) {
      const data: Lesson[] = await res.json();
      return data.map((l) => ({
        ...l,
        title: l.title || l.notes || 'Занятие',
        online_link: l.format === 'online' ? (l.location_or_url || l.online_link) : undefined,
        classroom_name: l.classroom_name || (l.format === 'offline' ? l.location_or_url : undefined),
        comment: l.comment || l.notes,
        decline_reason: l.decline_reason || l.cancel_reason,
      }));
    }
  } catch {
    void 0;
  }

  let lessons = getStored<Lesson[]>(STORAGE_KEY_LESSONS, initMockLessons());
  if (params?.teacher_id) {
    lessons = lessons.filter((l) => l.teacher_id === params.teacher_id);
  }
  if (params?.client_id) {
    lessons = lessons.filter((l) => l.client_id === params.client_id);
  }
  if (params?.student_id) {
    lessons = lessons.filter((l) => l.student_id === params.student_id || l.client_id === params.student_id);
  }
  return lessons;
}

export async function createLesson(data: CreateLessonRequest): Promise<Lesson> {
  const classrooms = getStored<Classroom[]>(STORAGE_KEY_CLASSROOMS, defaultClassrooms);
  const selectedRoom = data.classroom_id
    ? classrooms.find((c) => c.id === data.classroom_id)
    : undefined;

  const targetId = data.client_id || data.student_id || '';
  const payload = {
    ...data,
    client_id: targetId,
    student_id: targetId,
    location_or_url: data.online_link || (selectedRoom ? selectedRoom.name : undefined),
    notes: data.title + (data.comment ? ` (${data.comment})` : ''),
  };

  try {
    const res = await fetch(`${BASE_URL}/lessons`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify(payload),
    });
    if (res.ok) {
      const created: Lesson = await res.json();
      return {
        ...created,
        title: created.title || created.notes || data.title,
        online_link: created.format === 'online' ? (created.location_or_url || data.online_link) : undefined,
        classroom_name: selectedRoom?.name,
      };
    }
  } catch {
    void 0;
  }

  const clients = getStored<Client[]>(STORAGE_KEY_CLIENTS, defaultClients);
  const client = clients.find((c) => c.id === targetId);
  const student = mockUsers.find((u) => u.id === targetId);
  const displayName = client?.name || student?.full_name || 'Ученик';

  const newLesson: Lesson = {
    id: `lesson-${Date.now()}`,
    teacher_id: 'teacher-1',
    client_id: targetId,
    student_id: targetId,
    classroom_id: data.classroom_id || null,
    title: data.title,
    format: data.format,
    status: 'scheduled',
    start_time: data.start_time,
    end_time: data.end_time,
    online_link: data.online_link || null,
    comment: data.comment || null,
    client_name: displayName,
    student_name: displayName,
    teacher_name: 'Александр Верников',
    classroom_name: selectedRoom?.name,
    classroom_color: selectedRoom?.color || (data.format === 'online' ? '#10B981' : '#4F46E5'),
    created_at: new Date().toISOString(),
  };

  const storedLessons = getStored<Lesson[]>(STORAGE_KEY_LESSONS, initMockLessons());
  setStored(STORAGE_KEY_LESSONS, [newLesson, ...storedLessons]);
  return newLesson;
}

export async function completeLesson(lessonId: string): Promise<Lesson> {
  try {
    const res = await fetch(`${BASE_URL}/lessons/${lessonId}/complete`, {
      method: 'POST',
      headers: getHeaders(),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    void 0;
  }

  const storedLessons = getStored<Lesson[]>(STORAGE_KEY_LESSONS, initMockLessons());
  let completedLesson: Lesson | undefined;

  const updatedLessons = storedLessons.map((l) => {
    if (l.id === lessonId) {
      completedLesson = { ...l, status: 'completed' as const };
      return completedLesson;
    }
    return l;
  });

  setStored(STORAGE_KEY_LESSONS, updatedLessons);

  if (completedLesson) {
    const clients = getStored<Client[]>(STORAGE_KEY_CLIENTS, defaultClients);
    const targetClientId = completedLesson.client_id || completedLesson.student_id;
    setStored(
      STORAGE_KEY_CLIENTS,
      clients.map((c) =>
        c.id === targetClientId ? { ...c, balance: Math.max(0, c.balance - 1) } : c
      )
    );
  }

  if (!completedLesson) throw new Error('Lesson not found');
  return completedLesson;
}

export async function markNoShow(lessonId: string): Promise<Lesson> {
  try {
    const res = await fetch(`${BASE_URL}/lessons/${lessonId}/no-show`, {
      method: 'POST',
      headers: getHeaders(),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    void 0;
  }

  const storedLessons = getStored<Lesson[]>(STORAGE_KEY_LESSONS, initMockLessons());
  let updatedLesson: Lesson | undefined;
  setStored(
    STORAGE_KEY_LESSONS,
    storedLessons.map((l) => {
      if (l.id === lessonId) {
        updatedLesson = { ...l, status: 'no_show' as const };
        return updatedLesson;
      }
      return l;
    })
  );
  if (!updatedLesson) throw new Error('Lesson not found');
  return updatedLesson;
}

export async function acceptLesson(lessonId: string): Promise<Lesson> {
  try {
    const res = await fetch(`${BASE_URL}/lessons/${lessonId}/accept`, {
      method: 'POST',
      headers: getHeaders(),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    void 0;
  }

  const storedLessons = getStored<Lesson[]>(STORAGE_KEY_LESSONS, initMockLessons());
  const updated = storedLessons.map((l) =>
    l.id === lessonId ? { ...l, status: 'confirmed' as const } : l
  );
  setStored(STORAGE_KEY_LESSONS, updated);
  const found = updated.find((l) => l.id === lessonId);
  if (!found) throw new Error('Lesson not found');
  return found;
}

export async function declineLesson(lessonId: string, reason: string): Promise<Lesson> {
  try {
    const res = await fetch(`${BASE_URL}/lessons/${lessonId}/decline`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ reason }),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    void 0;
  }

  const storedLessons = getStored<Lesson[]>(STORAGE_KEY_LESSONS, initMockLessons());
  const updated = storedLessons.map((l) =>
    l.id === lessonId
      ? { ...l, status: 'declined' as const, decline_reason: reason }
      : l
  );
  setStored(STORAGE_KEY_LESSONS, updated);
  const found = updated.find((l) => l.id === lessonId);
  if (!found) throw new Error('Lesson not found');
  return found;
}

export async function cancelLesson(lessonId: string, reason?: string): Promise<Lesson> {
  try {
    const res = await fetch(`${BASE_URL}/lessons/${lessonId}/cancel`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ reason }),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    void 0;
  }

  const storedLessons = getStored<Lesson[]>(STORAGE_KEY_LESSONS, initMockLessons());
  const updated = storedLessons.map((l) =>
    l.id === lessonId
      ? { ...l, status: 'cancelled' as const, cancel_reason: reason }
      : l
  );
  setStored(STORAGE_KEY_LESSONS, updated);
  const found = updated.find((l) => l.id === lessonId);
  if (!found) throw new Error('Lesson not found');
  return found;
}

// ---------------------------------------------------------------------------
// 5. Финансовый Дашборд (Dashboard Metrics)
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
    void 0;
  }

  const fallback = await getFinancialDashboard();
  return {
    gross_potential_revenue: fallback.gross_revenue,
    net_income: fallback.net_income,
    average_rate: fallback.average_hourly_rate,
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
    void 0;
  }

  // Local calculation fallback
  const lessons = getStored<Lesson[]>(STORAGE_KEY_LESSONS, initMockLessons());
  const clients = getStored<Client[]>(STORAGE_KEY_CLIENTS, defaultClients);

  let gross = 0;
  let net = 0;
  let hours = 0;

  lessons.forEach((l) => {
    if (l.status === 'cancelled') return;
    const client = clients.find((c) => c.id === (l.client_id || l.student_id));
    const rate = client ? client.base_rate : 1500;

    const start = new Date(l.start_time).getTime();
    const end = new Date(l.end_time).getTime();
    const durationHours = Math.max(0.5, (end - start) / (1000 * 60 * 60));

    hours += durationHours;
    const lessonGross = durationHours * rate;
    gross += lessonGross;

    const commissionPercent = client?.school_percent_tag ?? (client?.tag?.match(/(\d+)%/) ? parseInt(client.tag.match(/(\d+)%/)![1], 10) : 0);
    const commission = lessonGross * (commissionPercent / 100);
    net += lessonGross - commission;
  });

  const avgRate = hours > 0 ? Math.round(gross / hours) : 1500;
  return {
    gross_potential_revenue: Math.round(gross),
    gross_revenue: Math.round(gross),
    net_income: Math.round(net),
    average_rate: avgRate,
    average_hourly_rate: avgRate,
  };
}
