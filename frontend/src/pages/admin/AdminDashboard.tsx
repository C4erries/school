import React, { useState, useEffect, useCallback } from 'react';
import { GlassCard } from '../../shared/components/GlassCard';
import { GlassButton } from '../../shared/components/GlassButton';
import { GlassInput } from '../../shared/components/GlassInput';
import { GlassModal } from '../../shared/components/GlassModal';
import { Badge } from '../../shared/components/Badge';
import { Classroom, TeacherStudent, Lesson } from '../../types/schedule';
import { User } from '../../types/auth';
import {
  getClassrooms,
  createClassroom,
  deleteClassroom,
  getTeacherStudents,
  assignStudentToTeacher,
  removeTeacherStudent,
  getUsersByRole,
  getLessons,
} from '../../api/schedule';
import {
  Building2,
  Users,
  Plus,
  Trash2,
  UserCheck,
  Calendar,
  MapPin,
  Video,
  Clock,
  RefreshCw,
} from 'lucide-react';

const PRESET_COLORS = [
  { name: 'Indigo', value: '#4F46E5' },
  { name: 'Mint', value: '#10B981' },
  { name: 'Amber', value: '#F59E0B' },
  { name: 'Coral', value: '#F43F5E' },
  { name: 'Purple', value: '#8B5CF6' },
  { name: 'Cyan', value: '#06B6D4' },
  { name: 'Rose', value: '#E11D48' },
];

export const AdminDashboard: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'classrooms' | 'relations' | 'overview'>('classrooms');

  // Classrooms State
  const [classrooms, setClassrooms] = useState<Classroom[]>([]);
  const [isClassroomModalOpen, setIsClassroomModalOpen] = useState(false);
  const [newRoomName, setNewRoomName] = useState('');
  const [newRoomCapacity, setNewRoomCapacity] = useState(4);
  const [newRoomColor, setNewRoomColor] = useState(PRESET_COLORS[0].value);
  const [newRoomDescription, setNewRoomDescription] = useState('');
  const [isSubmittingRoom, setIsSubmittingRoom] = useState(false);

  // Relations State
  const [relations, setRelations] = useState<TeacherStudent[]>([]);
  const [teachers, setTeachers] = useState<User[]>([]);
  const [students, setStudents] = useState<User[]>([]);
  const [selectedTeacherId, setSelectedTeacherId] = useState('');
  const [selectedStudentId, setSelectedStudentId] = useState('');
  const [isAssigning, setIsAssigning] = useState(false);

  // Overview State
  const [lessons, setLessons] = useState<Lesson[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const loadData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [roomsData, relsData, teachersData, studentsData, lessonsData] = await Promise.all([
        getClassrooms(),
        getTeacherStudents(),
        getUsersByRole('teacher'),
        getUsersByRole('student'),
        getLessons(),
      ]);

      setClassrooms(roomsData);
      setRelations(relsData);
      setTeachers(teachersData);
      setStudents(studentsData);
      setLessons(lessonsData);

      if (teachersData.length > 0 && !selectedTeacherId) {
        setSelectedTeacherId(teachersData[0].id);
      }
      if (studentsData.length > 0 && !selectedStudentId) {
        setSelectedStudentId(studentsData[0].id);
      }
    } catch (err) {
      console.error('Failed to load admin dashboard data', err);
    } finally {
      setIsLoading(false);
    }
  }, [selectedTeacherId, selectedStudentId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Handle classroom create
  const handleCreateClassroom = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newRoomName.trim()) return;

    setIsSubmittingRoom(true);
    try {
      await createClassroom({
        name: newRoomName.trim(),
        capacity: Number(newRoomCapacity) || 1,
        color: newRoomColor,
        description: newRoomDescription.trim() || undefined,
      });
      setNewRoomName('');
      setNewRoomCapacity(4);
      setNewRoomDescription('');
      setIsClassroomModalOpen(false);
      await loadData();
    } catch (err) {
      console.error('Failed to create classroom', err);
    } finally {
      setIsSubmittingRoom(false);
    }
  };

  const handleDeleteClassroom = async (id: string) => {
    if (!window.confirm('Вы действительно хотите удалить этот кабинет?')) return;
    try {
      await deleteClassroom(id);
      await loadData();
    } catch (err) {
      console.error('Failed to delete classroom', err);
    }
  };

  // Handle assign student
  const handleAssignStudent = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedTeacherId || !selectedStudentId) return;

    setIsAssigning(true);
    try {
      await assignStudentToTeacher(selectedTeacherId, selectedStudentId);
      await loadData();
    } catch (err) {
      console.error('Failed to assign student', err);
    } finally {
      setIsAssigning(false);
    }
  };

  const handleRemoveRelation = async (id: string) => {
    if (!window.confirm('Открепить ученика от преподавателя?')) return;
    try {
      await removeTeacherStudent(id);
      await loadData();
    } catch (err) {
      console.error('Failed to remove relation', err);
    }
  };

  const getStatusBadge = (status: Lesson['status']) => {
    switch (status) {
      case 'confirmed':
        return <Badge variant="mint">Подтверждён</Badge>;
      case 'pending_confirmation':
        return <Badge variant="amber">Ожидает согласия</Badge>;
      case 'completed':
        return <Badge variant="mint">Проведён</Badge>;
      case 'no_show':
        return <Badge variant="coral">Неявка</Badge>;
      case 'declined':
        return <Badge variant="coral">Отклонён</Badge>;
      default:
        return <Badge variant="neutral">{status}</Badge>;
    }
  };

  return (
    <div className="space-y-6">
      {/* Dashboard Title & Tabs */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">
              Панель Администратора
            </h1>
            <p className="text-sm text-slate-500 mt-1">
              Управление аудиторным фондом, распределение учеников и мониторинг расписания школы
            </p>
          </div>

          <div className="flex items-center gap-2 p-1.5 rounded-2xl liquid-glass border border-white/60 self-start sm:self-auto shadow-sm">
            <button
              onClick={() => setActiveTab('classrooms')}
              className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-semibold transition-all ${
                activeTab === 'classrooms'
                  ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
              }`}
            >
              <Building2 className="w-4 h-4" />
              Кабинеты ({classrooms.length})
            </button>
            <button
              onClick={() => setActiveTab('relations')}
              className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-semibold transition-all ${
                activeTab === 'relations'
                  ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
              }`}
            >
              <UserCheck className="w-4 h-4" />
              Привязка ({relations.length})
            </button>
            <button
              onClick={() => setActiveTab('overview')}
              className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-semibold transition-all ${
                activeTab === 'overview'
                  ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
              }`}
            >
              <Calendar className="w-4 h-4" />
              Обзор уроков ({lessons.length})
            </button>
          </div>
        </div>

        {/* ------------------------------------------------------------- */}
        {/* TAB 1: КАБИНЕТЫ ШКОЛЫ */}
        {/* ------------------------------------------------------------- */}
        {activeTab === 'classrooms' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-bold text-slate-900">Аудитории и кабинеты</h2>
                <p className="text-xs text-slate-500">
                  Физические пространства школы с указанием вместимости и персональной цветовой индикацией
                </p>
              </div>
              <GlassButton
                variant="primary"
                onClick={() => setIsClassroomModalOpen(true)}
                icon={<Plus className="w-4 h-4" />}
              >
                Добавить кабинет
              </GlassButton>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
              {classrooms.map((room) => (
                <GlassCard key={room.id} interactive className="relative overflow-hidden group">
                  {/* Цветная акцентная плашка */}
                  <div
                    className="absolute top-0 left-0 right-0 h-2"
                    style={{ backgroundColor: room.color }}
                  />

                  <div className="flex items-start justify-between gap-3 pt-2">
                    <div className="flex items-center gap-3">
                      <div
                        className="w-10 h-10 rounded-2xl flex items-center justify-center text-white shadow-sm"
                        style={{ backgroundColor: room.color }}
                      >
                        <Building2 className="w-5 h-5" />
                      </div>
                      <div>
                        <h3 className="font-bold text-slate-900 text-base">{room.name}</h3>
                        <div className="flex items-center gap-1.5 text-xs text-slate-500 mt-0.5">
                          <Users className="w-3.5 h-3.5 text-slate-400" />
                          <span>Вместимость: {room.capacity} {room.capacity === 1 ? 'место' : 'мест'}</span>
                        </div>
                      </div>
                    </div>

                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteClassroom(room.id);
                      }}
                      className="opacity-0 group-hover:opacity-100 p-2 text-slate-400 hover:text-rose-500 rounded-xl hover:bg-rose-50 transition-all"
                      title="Удалить кабинет"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>

                  {room.description && (
                    <p className="text-xs text-slate-600 mt-4 leading-relaxed line-clamp-2">
                      {room.description}
                    </p>
                  )}

                  <div className="mt-4 pt-3 border-t border-black/[0.05] flex items-center justify-between text-xs">
                    <span className="text-slate-400 font-mono">ID: {room.id}</span>
                    <span
                      className="px-2.5 py-0.5 rounded-full text-[11px] font-semibold"
                      style={{
                        backgroundColor: `${room.color}15`,
                        color: room.color,
                        border: `1px solid ${room.color}35`,
                      }}
                    >
                      Цветовая метка
                    </span>
                  </div>
                </GlassCard>
              ))}

              {classrooms.length === 0 && !isLoading && (
                <div className="col-span-full p-12 text-center liquid-glass rounded-3xl">
                  <Building2 className="w-12 h-12 text-slate-300 mx-auto mb-3" />
                  <p className="text-base font-medium text-slate-700">Кабинеты ещё не добавлены</p>
                  <p className="text-xs text-slate-400 mt-1">
                    Создайте первый кабинет, чтобы преподаватели могли бронировать оффлайн-аудитории
                  </p>
                </div>
              )}
            </div>
          </div>
        )}

        {/* ------------------------------------------------------------- */}
        {/* TAB 2: ПРИВЯЗКА УЧЕНИКОВ К ПРЕПОДАВАТЕЛЯМ */}
        {/* ------------------------------------------------------------- */}
        {activeTab === 'relations' && (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Форма прикрепления */}
            <div className="lg:col-span-1">
              <GlassCard className="sticky top-6">
                <div className="flex items-center gap-2.5 mb-4">
                  <div className="w-8 h-8 rounded-xl bg-indigo-600/10 text-indigo-600 flex items-center justify-center">
                    <UserCheck className="w-4 h-4" />
                  </div>
                  <div>
                    <h3 className="font-bold text-slate-900 text-base">Закрепить ученика</h3>
                    <p className="text-xs text-slate-500">Привязка для планирования уроков</p>
                  </div>
                </div>

                <form onSubmit={handleAssignStudent} className="space-y-4">
                  <div>
                    <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
                      Преподаватель
                    </label>
                    <select
                      value={selectedTeacherId}
                      onChange={(e) => setSelectedTeacherId(e.target.value)}
                      className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 liquid-glass-input focus:border-indigo-600 focus:outline-none"
                    >
                      {teachers.map((t) => (
                        <option key={t.id} value={t.id}>
                          {t.full_name} ({t.email})
                        </option>
                      ))}
                    </select>
                  </div>

                  <div>
                    <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
                      Ученик
                    </label>
                    <select
                      value={selectedStudentId}
                      onChange={(e) => setSelectedStudentId(e.target.value)}
                      className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 liquid-glass-input focus:border-indigo-600 focus:outline-none"
                    >
                      {students.map((s) => (
                        <option key={s.id} value={s.id}>
                          {s.full_name} ({s.email})
                        </option>
                      ))}
                    </select>
                  </div>

                  <GlassButton
                    type="submit"
                    variant="primary"
                    className="w-full mt-2"
                    isLoading={isAssigning}
                    icon={<Plus className="w-4 h-4" />}
                  >
                    Закрепить ученика
                  </GlassButton>
                </form>
              </GlassCard>
            </div>

            {/* Список закреплений */}
            <div className="lg:col-span-2 space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="font-bold text-slate-900 text-lg">Текущие связки преподаватель-ученик</h3>
                  <p className="text-xs text-slate-500">
                    Ученики, которым преподаватели могут назначать уроки
                  </p>
                </div>
                <Badge variant="indigo">{relations.length} связей</Badge>
              </div>

              <div className="space-y-3">
                {relations.map((rel) => (
                  <GlassCard key={rel.id} className="p-4 sm:p-5 flex items-center justify-between gap-4">
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 rounded-2xl bg-indigo-50 border border-indigo-200/60 flex items-center justify-center text-indigo-600 font-bold text-sm">
                        {rel.teacher_name?.charAt(0) || 'П'}
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-bold text-slate-900 text-sm">
                            {rel.teacher_name || 'Преподаватель'}
                          </span>
                          <span className="text-xs text-slate-400">→</span>
                          <span className="font-semibold text-emerald-700 bg-emerald-50 px-2.5 py-0.5 rounded-full text-xs border border-emerald-200/50">
                            {rel.student_name || 'Ученик'}
                          </span>
                        </div>
                        <span className="text-[11px] text-slate-400 block mt-0.5">
                          Закреплен: {new Date(rel.created_at).toLocaleDateString('ru-RU')}
                        </span>
                      </div>
                    </div>

                    <button
                      onClick={() => handleRemoveRelation(rel.id)}
                      className="p-2 text-slate-400 hover:text-rose-500 hover:bg-rose-50 rounded-xl transition-all"
                      title="Открепить"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </GlassCard>
                ))}

                {relations.length === 0 && (
                  <div className="p-10 text-center liquid-glass rounded-3xl text-slate-500 text-sm">
                    Связи пока не созданы. Выберите преподавателя и ученика в панели слева.
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {/* ------------------------------------------------------------- */}
        {/* TAB 3: СКВОЗНОЙ ОБЗОР РАСПИСАНИЯ */}
        {/* ------------------------------------------------------------- */}
        {activeTab === 'overview' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-bold text-slate-900">Сквозное расписание занятий</h2>
                <p className="text-xs text-slate-500">
                  Все запланированные, подтвержденные и проведенные уроки по всей школе
                </p>
              </div>
              <GlassButton
                variant="secondary"
                size="sm"
                onClick={loadData}
                icon={<RefreshCw className="w-4 h-4" />}
              >
                Обновить
              </GlassButton>
            </div>

            <div className="space-y-3">
              {lessons.map((lesson) => (
                <GlassCard key={lesson.id} className="p-5 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
                  <div className="flex items-start gap-3.5">
                    <div
                      className="w-11 h-11 rounded-2xl flex items-center justify-center text-white shrink-0 mt-0.5 shadow-sm"
                      style={{
                        backgroundColor:
                          lesson.classroom_color ||
                          (lesson.format === 'online' ? '#10B981' : '#4F46E5'),
                      }}
                    >
                      {lesson.format === 'online' ? (
                        <Video className="w-5 h-5" />
                      ) : (
                        <MapPin className="w-5 h-5" />
                      )}
                    </div>
                    <div>
                      <div className="flex items-center gap-2 flex-wrap">
                        <h4 className="font-bold text-slate-900 text-sm sm:text-base">
                          {lesson.title}
                        </h4>
                        {getStatusBadge(lesson.status)}
                      </div>

                      <div className="flex items-center gap-3 text-xs text-slate-500 mt-1 flex-wrap">
                        <span className="font-medium text-slate-700">
                          Преподаватель: {lesson.teacher_name || 'Преподаватель'}
                        </span>
                        <span>•</span>
                        <span className="font-medium text-slate-700">
                          Ученик: {lesson.student_name || 'Ученик'}
                        </span>
                        <span>•</span>
                        <span className="flex items-center gap-1">
                          <Clock className="w-3.5 h-3.5 text-slate-400" />
                          {new Date(lesson.start_time).toLocaleTimeString('ru-RU', {
                            hour: '2-digit',
                            minute: '2-digit',
                          })}{' '}
                          –{' '}
                          {new Date(lesson.end_time).toLocaleTimeString('ru-RU', {
                            hour: '2-digit',
                            minute: '2-digit',
                          })}
                        </span>
                      </div>

                      {lesson.classroom_name && (
                        <div className="mt-1 text-xs text-indigo-600 font-medium flex items-center gap-1">
                          <Building2 className="w-3.5 h-3.5" />
                          {lesson.classroom_name}
                        </div>
                      )}
                    </div>
                  </div>

                  <div className="text-xs text-slate-400 font-mono">
                    {new Date(lesson.start_time).toLocaleDateString('ru-RU', {
                      day: 'numeric',
                      month: 'long',
                    })}
                  </div>
                </GlassCard>
              ))}

              {lessons.length === 0 && (
                <div className="p-12 text-center liquid-glass rounded-3xl text-slate-500 text-sm">
                  В расписании пока нет уроков. Преподаватели могут создавать их со своих страниц.
                </div>
              )}
            </div>
          </div>
        )}

      {/* ------------------------------------------------------------- */}
      {/* MODAL: ДОБАВЛЕНИЕ КАБИНЕТА */}
      {/* ------------------------------------------------------------- */}
      <GlassModal
        isOpen={isClassroomModalOpen}
        onClose={() => setIsClassroomModalOpen(false)}
        title="Новый кабинет школы"
        description="Задайте название, вместимость и цвет плашки для отображения в расписании"
      >
        <form onSubmit={handleCreateClassroom} className="space-y-4">
          <GlassInput
            label="Название аудитории"
            placeholder="например: Кабинет 204 — Пифагор"
            value={newRoomName}
            onChange={(e) => setNewRoomName(e.target.value)}
            required
          />

          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
              Вместимость (человек)
            </label>
            <input
              type="number"
              min={1}
              max={100}
              value={newRoomCapacity}
              onChange={(e) => setNewRoomCapacity(parseInt(e.target.value, 10) || 1)}
              className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 liquid-glass-input focus:border-indigo-600 focus:outline-none"
              required
            />
          </div>

          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
              Цветовая плашка
            </label>
            <div className="flex items-center gap-2.5 pt-1">
              {PRESET_COLORS.map((c) => (
                <button
                  key={c.value}
                  type="button"
                  onClick={() => setNewRoomColor(c.value)}
                  className={`w-8 h-8 rounded-full transition-transform ${
                    newRoomColor === c.value
                      ? 'scale-125 ring-2 ring-indigo-500 ring-offset-2'
                      : 'hover:scale-110'
                  }`}
                  style={{ backgroundColor: c.value }}
                  title={c.name}
                />
              ))}
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
              Описание и оснащение (опционально)
            </label>
            <textarea
              rows={3}
              placeholder="Интерактивная доска, проектор, кондиционер..."
              value={newRoomDescription}
              onChange={(e) => setNewRoomDescription(e.target.value)}
              className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 liquid-glass-input focus:border-indigo-600 focus:outline-none resize-none"
            />
          </div>

          <div className="flex items-center justify-end gap-3 pt-3 border-t border-black/[0.05]">
            <GlassButton
              type="button"
              variant="secondary"
              onClick={() => setIsClassroomModalOpen(false)}
            >
              Отмена
            </GlassButton>
            <GlassButton
              type="submit"
              variant="primary"
              isLoading={isSubmittingRoom}
            >
              Создать кабинет
            </GlassButton>
          </div>
        </form>
      </GlassModal>
    </div>
  );
};
