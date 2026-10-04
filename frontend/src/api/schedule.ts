import {
  Classroom,
  TeacherStudent,
  Lesson,
  CreateClassroomRequest,
  CreateLessonRequest,
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
// Fallback local storage for mock data when backend routes are not implemented yet
// ---------------------------------------------------------------------------
const STORAGE_KEY_CLASSROOMS = 'school_mock_classrooms';
const STORAGE_KEY_RELATIONS = 'school_mock_teacher_students';
const STORAGE_KEY_LESSONS = 'school_mock_lessons';

const defaultClassrooms: Classroom[] = [
  {
    id: 'room-101',
    name: 'Кабинет 101 — Квант',
    capacity: 4,
    color: '#4F46E5', // Apple Indigo
    description: 'Интерактивная доска, микроскоп, рабочие места для точных наук',
    created_at: new Date().toISOString(),
  },
  {
    id: 'room-102',
    name: 'Кабинет 102 — Эврика',
    capacity: 2,
    color: '#10B981', // Apple Mint
    description: 'Звукоизоляция, индивидуальные парты для репетиторства 1-на-1',
    created_at: new Date().toISOString(),
  },
  {
    id: 'room-201',
    name: 'Кабинет 201 — Логос',
    capacity: 6,
    color: '#F59E0B', // Apple Amber
    description: 'Гуманитарный кабинет, круглый стол, проектор',
    created_at: new Date().toISOString(),
  },
  {
    id: 'room-301',
    name: 'Аудитория 301 — Сириус',
    capacity: 12,
    color: '#8B5CF6', // Apple Purple
    description: 'Большой лекторий с системой видеоконференций',
    created_at: new Date().toISOString(),
  },
];

const mockUsers: User[] = [
  {
    id: 'teacher-1',
    email: 'math.teacher@school.ru',
    full_name: 'Александр Верников',
    role: 'teacher',
    created_at: '2026-09-01T10:00:00Z',
  },
  {
    id: 'teacher-2',
    email: 'physics.teacher@school.ru',
    full_name: 'Елена Соколова',
    role: 'teacher',
    created_at: '2026-09-01T10:00:00Z',
  },
  {
    id: 'student-1',
    email: 'student.ivan@school.ru',
    full_name: 'Иван Смирнов',
    role: 'student',
    created_at: '2026-09-05T12:00:00Z',
  },
  {
    id: 'student-2',
    email: 'student.anna@school.ru',
    full_name: 'Анна Кузнецова',
    role: 'student',
    created_at: '2026-09-06T14:00:00Z',
  },
  {
    id: 'student-3',
    email: 'student.dmitry@school.ru',
    full_name: 'Дмитрий Васильев',
    role: 'student',
    created_at: '2026-09-07T11:00:00Z',
  },
];

const defaultRelations: TeacherStudent[] = [
  {
    id: 'ts-1',
    teacher_id: 'teacher-1',
    student_id: 'student-1',
    teacher_name: 'Александр Верников',
    student_name: 'Иван Смирнов',
    created_at: '2026-09-10T09:00:00Z',
  },
  {
    id: 'ts-2',
    teacher_id: 'teacher-1',
    student_id: 'student-2',
    teacher_name: 'Александр Верников',
    student_name: 'Анна Кузнецова',
    created_at: '2026-09-11T10:00:00Z',
  },
  {
    id: 'ts-3',
    teacher_id: 'teacher-2',
    student_id: 'student-3',
    teacher_name: 'Елена Соколова',
    student_name: 'Дмитрий Васильев',
    created_at: '2026-09-12T11:00:00Z',
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

// Генерация стартовых уроков на сегодня/завтра для наглядной демонстрации
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
      student_id: 'student-1',
      classroom_id: 'room-101',
      title: 'Математический анализ: Пределы и производные',
      format: 'offline',
      status: 'pending_confirmation',
      start_time: `${dateStr}T14:00:00`,
      end_time: `${dateStr}T15:30:00`,
      teacher_name: 'Александр Верников',
      student_name: 'Иван Смирнов',
      classroom_name: 'Кабинет 101 — Квант',
      classroom_color: '#4F46E5',
      comment: 'Подготовить домашний конспект по правилу Лопиталя',
      created_at: new Date().toISOString(),
    },
    // Накладывающийся урок 1 (демонстрация Apple Calendar cascade)
    {
      id: 'lesson-2',
      teacher_id: 'teacher-1',
      student_id: 'student-2',
      classroom_id: null,
      title: 'ЕГЭ Профиль: Стереометрия и векторы',
      format: 'online',
      status: 'confirmed',
      start_time: `${dateStr}T14:30:00`,
      end_time: `${dateStr}T16:00:00`,
      online_link: 'https://telemost.yandex.ru/j/school-session-math',
      teacher_name: 'Александр Верников',
      student_name: 'Анна Кузнецова',
      classroom_color: '#10B981',
      comment: 'Онлайн-разбор задач №14 из открытого банка ФИПИ',
      created_at: new Date().toISOString(),
    },
    // Урок позже сегодня
    {
      id: 'lesson-3',
      teacher_id: 'teacher-1',
      student_id: 'student-3',
      classroom_id: 'room-201',
      title: 'Олимпиадная математика',
      format: 'offline',
      status: 'completed',
      start_time: `${dateStr}T11:00:00`,
      end_time: `${dateStr}T12:30:00`,
      teacher_name: 'Александр Верников',
      student_name: 'Дмитрий Васильев',
      classroom_name: 'Кабинет 201 — Логос',
      classroom_color: '#F59E0B',
      created_at: new Date().toISOString(),
    },
    // Предстоящий подтвержденный оффлайн-урок
    {
      id: 'lesson-4',
      teacher_id: 'teacher-1',
      student_id: 'student-1',
      classroom_id: 'room-102',
      title: 'Геометрия: Теорема Менелая и Чевы',
      format: 'offline',
      status: 'confirmed',
      start_time: `${dateStr}T17:00:00`,
      end_time: `${dateStr}T18:00:00`,
      teacher_name: 'Александр Верников',
      student_name: 'Иван Смирнов',
      classroom_name: 'Кабинет 102 — Эврика',
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
    const res = await fetch(`${BASE_URL}/classrooms`, {
      method: 'GET',
      headers: getHeaders(),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    // API server not reachable or endpoint not implemented
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
    // fallback
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
  const updated = [newRoom, ...existing];
  setStored(STORAGE_KEY_CLASSROOMS, updated);
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
    // fallback
  }

  const existing = getStored<Classroom[]>(STORAGE_KEY_CLASSROOMS, defaultClassrooms);
  setStored(
    STORAGE_KEY_CLASSROOMS,
    existing.filter((r) => r.id !== id)
  );
}

// ---------------------------------------------------------------------------
// 2. Привязка преподавателей и учеников (TeacherStudent)
// ---------------------------------------------------------------------------
export async function getTeacherStudents(teacherId?: string): Promise<TeacherStudent[]> {
  try {
    const query = teacherId ? `?teacher_id=${encodeURIComponent(teacherId)}` : '';
    const res = await fetch(`${BASE_URL}/teacher-students${query}`, {
      method: 'GET',
      headers: getHeaders(),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    // fallback
  }

  const all = getStored<TeacherStudent[]>(STORAGE_KEY_RELATIONS, defaultRelations);
  if (teacherId) {
    return all.filter((r) => r.teacher_id === teacherId);
  }
  return all;
}

export async function assignStudentToTeacher(
  teacherId: string,
  studentId: string
): Promise<TeacherStudent> {
  try {
    const res = await fetch(`${BASE_URL}/teacher-students`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ teacher_id: teacherId, student_id: studentId }),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch {
    // fallback
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
    await fetch(`${BASE_URL}/teacher-students/${id}`, {
      method: 'DELETE',
      headers: getHeaders(),
    });
  } catch {
    // fallback
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
    // fallback
  }

  // Include any registered users in mock list
  return mockUsers.filter((u) => u.role === role);
}

// ---------------------------------------------------------------------------
// 3. Уроки и Расписание (Lessons)
// ---------------------------------------------------------------------------
export async function getLessons(params?: {
  teacher_id?: string;
  student_id?: string;
  date?: string;
}): Promise<Lesson[]> {
  try {
    const query = new URLSearchParams();
    if (params?.teacher_id) query.set('teacher_id', params.teacher_id);
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
    // fallback
  }

  let lessons = getStored<Lesson[]>(STORAGE_KEY_LESSONS, initMockLessons());
  if (params?.teacher_id) {
    lessons = lessons.filter((l) => l.teacher_id === params.teacher_id);
  }
  if (params?.student_id) {
    lessons = lessons.filter((l) => l.student_id === params.student_id);
  }
  return lessons;
}

export async function createLesson(data: CreateLessonRequest): Promise<Lesson> {
  const classrooms = getStored<Classroom[]>(STORAGE_KEY_CLASSROOMS, defaultClassrooms);
  const selectedRoom = data.classroom_id
    ? classrooms.find((c) => c.id === data.classroom_id)
    : undefined;

  const payload = {
    ...data,
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
    // fallback
  }

  const student = mockUsers.find((u) => u.id === data.student_id);
  const storedLessons = getStored<Lesson[]>(STORAGE_KEY_LESSONS, initMockLessons());

  const newLesson: Lesson = {
    id: `lesson-${Date.now()}`,
    teacher_id: 'teacher-1',
    student_id: data.student_id,
    classroom_id: data.classroom_id || null,
    title: data.title,
    format: data.format,
    status: 'pending_confirmation',
    start_time: data.start_time,
    end_time: data.end_time,
    online_link: data.online_link || null,
    comment: data.comment || null,
    student_name: student?.full_name || 'Ученик',
    teacher_name: 'Преподаватель',
    classroom_name: selectedRoom?.name,
    classroom_color: selectedRoom?.color || (data.format === 'online' ? '#10B981' : '#4F46E5'),
    created_at: new Date().toISOString(),
  };

  const updated = [newLesson, ...storedLessons];
  setStored(STORAGE_KEY_LESSONS, updated);
  return newLesson;
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
    // fallback
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
    // fallback
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
    // fallback
  }

  const storedLessons = getStored<Lesson[]>(STORAGE_KEY_LESSONS, initMockLessons());
  const updated = storedLessons.map((l) =>
    l.id === lessonId ? { ...l, status: 'completed' as const } : l
  );
  setStored(STORAGE_KEY_LESSONS, updated);
  const found = updated.find((l) => l.id === lessonId);
  if (!found) throw new Error('Lesson not found');
  return found;
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
    // fallback
  }

  const storedLessons = getStored<Lesson[]>(STORAGE_KEY_LESSONS, initMockLessons());
  const updated = storedLessons.map((l) =>
    l.id === lessonId ? { ...l, status: 'no_show' as const } : l
  );
  setStored(STORAGE_KEY_LESSONS, updated);
  const found = updated.find((l) => l.id === lessonId);
  if (!found) throw new Error('Lesson not found');
  return found;
}
