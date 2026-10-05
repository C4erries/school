import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { GlassCard } from '../../shared/components/GlassCard';
import { GlassButton } from '../../shared/components/GlassButton';
import { GlassInput } from '../../shared/components/GlassInput';
import { GlassModal } from '../../shared/components/GlassModal';
import { Badge } from '../../shared/components/Badge';
import { AppNavbar } from '../../shared/components/AppNavbar';
import { LiquidBackground } from '../../shared/components/LiquidBackground';
import { Classroom, TeacherStudent, Lesson, LessonFormat } from '../../types/schedule';
import {
  getClassrooms,
  getTeacherStudents,
  getLessons,
  createLesson,
  completeLesson,
  markNoShow,
} from '../../api/schedule';
import {
  Calendar as CalendarIcon,
  Plus,
  Video,
  MapPin,
  Clock,
  CheckCircle,
  AlertTriangle,
  ChevronLeft,
  ChevronRight,
  Columns,
  List,
  Search,
} from 'lucide-react';

interface PositionedLesson extends Lesson {
  column: number;
  totalColumns: number;
  startMinutes: number;
  durationMinutes: number;
}

export const TeacherSchedulePage: React.FC = () => {
  // Data
  const [students, setStudents] = useState<TeacherStudent[]>([]);
  const [classrooms, setClassrooms] = useState<Classroom[]>([]);
  const [lessons, setLessons] = useState<Lesson[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Selected date for calendar view
  const [selectedDate, setSelectedDate] = useState<Date>(new Date());
  const [viewMode, setViewMode] = useState<'timeline' | 'list'>('timeline');

  // New Lesson Modal State
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedStudentId, setSelectedStudentId] = useState('');
  const [lessonTitle, setLessonTitle] = useState('');
  const [lessonDate, setLessonDate] = useState(
    new Date().toISOString().split('T')[0]
  );
  const [lessonStartTime, setLessonStartTime] = useState('14:00');
  const [lessonDuration, setLessonDuration] = useState('60'); // minutes
  const [lessonFormat, setLessonFormat] = useState<LessonFormat>('offline');
  const [selectedClassroomId, setSelectedClassroomId] = useState('');
  const [onlineLink, setOnlineLink] = useState('https://telemost.yandex.ru/j/school-');
  const [lessonComment, setLessonComment] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [roomSearch, setRoomSearch] = useState('');

  // Filter classrooms by search query
  const filteredClassrooms = useMemo(() => {
    if (!roomSearch.trim()) return classrooms;
    const q = roomSearch.toLowerCase();
    return classrooms.filter(
      (c) => c.name.toLowerCase().includes(q) || String(c.capacity).includes(q)
    );
  }, [classrooms, roomSearch]);

  // Load teacher data
  const loadData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [allRooms, myStudents, allLessons] = await Promise.all([
        getClassrooms(),
        getTeacherStudents(),
        getLessons(),
      ]);

      setClassrooms(allRooms);
      setStudents(myStudents);
      setLessons(allLessons);

      if (myStudents.length > 0 && !selectedStudentId) {
        setSelectedStudentId(myStudents[0].student_id);
      }
      if (allRooms.length > 0 && !selectedClassroomId) {
        setSelectedClassroomId(allRooms[0].id);
      }
    } catch (err) {
      console.error('Failed to load teacher schedule', err);
    } finally {
      setIsLoading(false);
    }
  }, [selectedClassroomId, selectedStudentId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Selected date string YYYY-MM-DD
  const selectedDateStr = useMemo(() => {
    const y = selectedDate.getFullYear();
    const m = String(selectedDate.getMonth() + 1).padStart(2, '0');
    const d = String(selectedDate.getDate()).padStart(2, '0');
    return `${y}-${m}-${d}`;
  }, [selectedDate]);

  // Filter lessons for selected date
  const dayLessons = useMemo(() => {
    return lessons.filter((l) => l.start_time.startsWith(selectedDateStr));
  }, [lessons, selectedDateStr]);

  // Apple Calendar Overlap Layout Engine:
  // Detects overlapping time windows and calculates (column, totalColumns)
  const positionedLessons = useMemo<PositionedLesson[]>(() => {
    if (dayLessons.length === 0) return [];

    const parsed = dayLessons.map((l) => {
      const start = new Date(l.start_time);
      const end = new Date(l.end_time);
      const startMinutes = start.getHours() * 60 + start.getMinutes();
      const endMinutes = end.getHours() * 60 + end.getMinutes();
      const durationMinutes = Math.max(30, endMinutes - startMinutes);

      return {
        ...l,
        startMinutes,
        durationMinutes,
        endMinutes,
        column: 0,
        totalColumns: 1,
      };
    });

    // Sort by start time, then by duration desc
    parsed.sort((a, b) => a.startMinutes - b.startMinutes || b.durationMinutes - a.durationMinutes);

    // Group overlapping intervals
    const clusters: (typeof parsed)[] = [];
    let currentCluster: typeof parsed = [];
    let clusterEnd = -1;

    for (const item of parsed) {
      if (currentCluster.length === 0 || item.startMinutes < clusterEnd) {
        currentCluster.push(item);
        clusterEnd = Math.max(clusterEnd, item.endMinutes);
      } else {
        clusters.push(currentCluster);
        currentCluster = [item];
        clusterEnd = item.endMinutes;
      }
    }
    if (currentCluster.length > 0) {
      clusters.push(currentCluster);
    }

    // Allocate columns within each overlapping cluster
    const result: PositionedLesson[] = [];
    for (const cluster of clusters) {
      const columns: number[] = []; // tracks end minute of current column

      for (const item of cluster) {
        let placedCol = -1;
        for (let col = 0; col < columns.length; col++) {
          if (columns[col] <= item.startMinutes) {
            placedCol = col;
            columns[col] = item.endMinutes;
            break;
          }
        }
        if (placedCol === -1) {
          placedCol = columns.length;
          columns.push(item.endMinutes);
        }
        item.column = placedCol;
      }

      const totalCols = Math.max(1, columns.length);
      for (const item of cluster) {
        item.totalColumns = totalCols;
        result.push(item);
      }
    }

    return result;
  }, [dayLessons]);

  // Quick Action: Complete Lesson
  const handleComplete = async (lessonId: string) => {
    try {
      await completeLesson(lessonId);
      await loadData();
    } catch (err) {
      console.error('Error completing lesson', err);
    }
  };

  // Quick Action: No Show
  const handleNoShow = async (lessonId: string) => {
    if (!window.confirm('Отметить неявку ученика на занятие?')) return;
    try {
      await markNoShow(lessonId);
      await loadData();
    } catch (err) {
      console.error('Error marking no show', err);
    }
  };

  // Handle Create Lesson Form
  const handleCreateLesson = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedStudentId || !lessonTitle.trim()) return;

    setIsSubmitting(true);
    try {
      const [hours, mins] = lessonStartTime.split(':').map(Number);
      const start = new Date(lessonDate);
      start.setHours(hours, mins, 0, 0);

      const end = new Date(start.getTime() + Number(lessonDuration) * 60 * 1000);

      await createLesson({
        student_id: selectedStudentId,
        title: lessonTitle.trim(),
        format: lessonFormat,
        classroom_id: lessonFormat === 'offline' ? selectedClassroomId : null,
        online_link: lessonFormat === 'online' ? onlineLink : undefined,
        comment: lessonComment.trim() || undefined,
        start_time: start.toISOString(),
        end_time: end.toISOString(),
      });

      setIsModalOpen(false);
      setLessonTitle('');
      setLessonComment('');
      setRoomSearch('');
      await loadData();
    } catch (err) {
      console.error('Failed to create lesson', err);
    } finally {
      setIsSubmitting(false);
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

  const changeDateBy = (days: number) => {
    const next = new Date(selectedDate);
    next.setDate(next.getDate() + days);
    setSelectedDate(next);
  };

  // Timeline hours from 08:00 to 20:00
  const hours = Array.from({ length: 13 }, (_, i) => i + 8);
  const START_HOUR = 8;
  const HOUR_HEIGHT = 70; // px per hour

  return (
    <div className="min-h-screen text-slate-800 p-4 sm:p-8 relative overflow-hidden">
      <LiquidBackground />

      <div className="max-w-6xl mx-auto space-y-6 relative z-10">
        <AppNavbar />

        {/* Top Header & Quick Schedule */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">
              Расписание Преподавателя
            </h1>
            <p className="text-sm text-slate-500 mt-1">
              Планирование занятий, нахлёст расписания и фиксация проведения уроков
            </p>
          </div>

          <GlassButton
            variant="primary"
            onClick={() => setIsModalOpen(true)}
            icon={<Plus className="w-4 h-4" />}
          >
            Назначить урок
          </GlassButton>
        </div>

        {/* Закрепленные ученики (горизонтальная карусель) */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-500">
              Закреплённые ученики ({students.length})
            </span>
          </div>
          <div className="flex items-center gap-3 overflow-x-auto pb-2 scrollbar-none">
            {students.map((rel) => (
              <div
                key={rel.id}
                onClick={() => {
                  setSelectedStudentId(rel.student_id);
                  setIsModalOpen(true);
                }}
                className="liquid-glass hover:bg-white/90 p-3 rounded-2xl flex items-center gap-3 shrink-0 cursor-pointer border border-white/60 transition-all shadow-sm hover:scale-[1.02]"
              >
                <div className="w-8 h-8 rounded-xl bg-indigo-50 border border-indigo-200/50 text-indigo-600 flex items-center justify-center font-bold text-xs">
                  {rel.student_name?.charAt(0) || 'У'}
                </div>
                <div className="text-left pr-2">
                  <div className="text-xs font-bold text-slate-900 leading-tight">
                    {rel.student_name || 'Ученик'}
                  </div>
                  <div className="text-[10px] text-indigo-600 font-medium">
                    + Запланировать
                  </div>
                </div>
              </div>
            ))}

            {students.length === 0 && (
              <div className="text-xs text-slate-400 p-2">
                Ученики ещё не прикреплены. Администратор может закрепить учеников в панели управления.
              </div>
            )}
          </div>
        </div>

        {/* Date Selector & View Switcher */}
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4 p-3 rounded-3xl liquid-glass border border-white/60 shadow-sm">
          <div className="flex items-center gap-2">
            <button
              onClick={() => changeDateBy(-1)}
              className="p-2 rounded-xl hover:bg-black/5 text-slate-600 transition-colors"
              title="Предыдущий день"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>

            <div className="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-white/70 border border-white/80">
              <CalendarIcon className="w-4 h-4 text-indigo-600" />
              <span className="font-bold text-sm text-slate-900 capitalize">
                {selectedDate.toLocaleDateString('ru-RU', {
                  weekday: 'short',
                  day: 'numeric',
                  month: 'long',
                  year: 'numeric',
                })}
              </span>
            </div>

            <button
              onClick={() => changeDateBy(1)}
              className="p-2 rounded-xl hover:bg-black/5 text-slate-600 transition-colors"
              title="Следующий день"
            >
              <ChevronRight className="w-4 h-4" />
            </button>

            <button
              onClick={() => setSelectedDate(new Date())}
              className="text-xs font-semibold px-3 py-1.5 rounded-xl text-indigo-600 hover:bg-indigo-50 transition-colors ml-1"
            >
              Сегодня
            </button>
          </div>

          <div className="flex items-center gap-1.5 p-1 rounded-2xl bg-black/[0.04] self-end sm:self-auto">
            <button
              onClick={() => setViewMode('timeline')}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
                viewMode === 'timeline'
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              <Columns className="w-3.5 h-3.5" />
              <span>Сетка</span>
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
                viewMode === 'list'
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              <List className="w-3.5 h-3.5" />
              <span>Список ({dayLessons.length})</span>
            </button>
          </div>
        </div>

        {/* ------------------------------------------------------------- */}
        {/* VIEW 1: APPLE CALENDAR TIMELINE WITH OVERLAPPING LESSONS */}
        {/* ------------------------------------------------------------- */}
        {viewMode === 'timeline' && (
          <GlassCard className="p-4 sm:p-6 overflow-hidden">
            <div className="relative border-t border-slate-200/60 select-none">
              {/* Timeline Grid Background */}
              <div className="relative">
                {hours.map((hour) => (
                  <div
                    key={hour}
                    className="flex items-start border-b border-slate-200/50"
                    style={{ height: `${HOUR_HEIGHT}px` }}
                  >
                    <div className="w-16 text-right pr-4 text-xs font-mono text-slate-400 -mt-2.5">
                      {String(hour).padStart(2, '0')}:00
                    </div>
                    <div className="flex-1 h-full border-l border-slate-200/60" />
                  </div>
                ))}
              </div>

              {/* Overlapping Lessons Layer */}
              <div
                className="absolute top-0 right-0 left-16 bottom-0 pointer-events-none"
                style={{ height: `${hours.length * HOUR_HEIGHT}px` }}
              >
                {positionedLessons.map((lesson) => {
                  const topOffset =
                    ((lesson.startMinutes - START_HOUR * 60) / 60) * HOUR_HEIGHT;
                  const cardHeight = (lesson.durationMinutes / 60) * HOUR_HEIGHT;

                  const widthPercent = 100 / lesson.totalColumns;
                  const leftPercent = lesson.column * widthPercent;

                  const accentColor =
                    lesson.classroom_color ||
                    (lesson.format === 'online' ? '#10B981' : '#4F46E5');

                  return (
                    <div
                      key={lesson.id}
                      className="absolute p-1 pointer-events-auto transition-all group z-10"
                      style={{
                        top: `${Math.max(0, topOffset)}px`,
                        height: `${Math.max(50, cardHeight)}px`,
                        left: `${leftPercent}%`,
                        width: `${widthPercent}%`,
                      }}
                    >
                      <div
                        className="h-full w-full rounded-2xl liquid-glass border border-white/80 p-3 flex flex-col justify-between shadow-md relative overflow-hidden transition-all group-hover:scale-[1.01] group-hover:shadow-lg"
                        style={{
                          borderLeftWidth: '5px',
                          borderLeftColor: accentColor,
                          backgroundColor: 'rgba(255, 255, 255, 0.85)',
                        }}
                      >
                        {/* Top Info */}
                        <div>
                          <div className="flex items-center justify-between gap-2">
                            <span className="font-bold text-xs sm:text-sm text-slate-900 truncate">
                              {lesson.title}
                            </span>
                            {getStatusBadge(lesson.status)}
                          </div>

                          <div className="flex items-center gap-2 text-[11px] text-slate-600 mt-0.5 truncate">
                            <span className="font-medium text-slate-800">
                              {lesson.student_name}
                            </span>
                            <span>•</span>
                            <span className="font-mono">
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

                          {lesson.format === 'offline' && lesson.classroom_name && (
                            <div className="flex items-center gap-1 text-[11px] font-medium text-slate-600 mt-1">
                              <MapPin className="w-3 h-3 text-slate-400" />
                              <span className="truncate">{lesson.classroom_name}</span>
                            </div>
                          )}

                          {lesson.format === 'online' && (
                            <div className="flex items-center gap-1 text-[11px] font-medium text-emerald-600 mt-1">
                              <Video className="w-3 h-3" />
                              <span>Онлайн урок</span>
                            </div>
                          )}
                        </div>

                        {/* Bottom Actions for Confirmed / Conducted */}
                        {lesson.status === 'confirmed' && (
                          <div className="flex items-center gap-2 pt-2 border-t border-black/[0.04]">
                            <button
                              onClick={() => handleComplete(lesson.id)}
                              className="px-2.5 py-1 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-white text-[11px] font-semibold flex items-center gap-1 shadow-sm transition-all active:scale-95"
                              title="Урок проведён"
                            >
                              <CheckCircle className="w-3 h-3" />
                              Проведён
                            </button>
                            <button
                              onClick={() => handleNoShow(lesson.id)}
                              className="px-2 py-1 rounded-lg bg-rose-50 hover:bg-rose-100 text-rose-600 text-[11px] font-medium flex items-center gap-1 transition-all active:scale-95"
                              title="Неявка ученика"
                            >
                              <AlertTriangle className="w-3 h-3" />
                              Неявка
                            </button>
                          </div>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          </GlassCard>
        )}

        {/* ------------------------------------------------------------- */}
        {/* VIEW 2: LIST VIEW */}
        {/* ------------------------------------------------------------- */}
        {viewMode === 'list' && (
          <div className="space-y-3">
            {dayLessons.map((lesson) => (
              <GlassCard
                key={lesson.id}
                className="p-5 flex flex-col md:flex-row items-start md:items-center justify-between gap-4"
              >
                <div className="flex items-start gap-4">
                  <div
                    className="w-12 h-12 rounded-2xl flex items-center justify-center text-white shrink-0 shadow-sm"
                    style={{
                      backgroundColor:
                        lesson.classroom_color ||
                        (lesson.format === 'online' ? '#10B981' : '#4F46E5'),
                    }}
                  >
                    {lesson.format === 'online' ? (
                      <Video className="w-6 h-6" />
                    ) : (
                      <MapPin className="w-6 h-6" />
                    )}
                  </div>

                  <div>
                    <div className="flex items-center gap-2 flex-wrap">
                      <h3 className="font-bold text-slate-900 text-base">
                        {lesson.title}
                      </h3>
                      {getStatusBadge(lesson.status)}
                    </div>

                    <div className="flex items-center gap-3 text-xs text-slate-500 mt-1 flex-wrap">
                      <span className="font-semibold text-slate-800">
                        Ученик: {lesson.student_name}
                      </span>
                      <span>•</span>
                      <span className="flex items-center gap-1 font-mono">
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
                      <div className="text-xs text-indigo-600 font-medium mt-1">
                        Аудитория: {lesson.classroom_name}
                      </div>
                    )}

                    {lesson.comment && (
                      <div className="text-xs text-slate-500 italic mt-1">
                        Примечание: {lesson.comment}
                      </div>
                    )}
                  </div>
                </div>

                {/* Действия */}
                <div className="flex items-center gap-2 self-end md:self-auto">
                  {lesson.status === 'confirmed' && (
                    <>
                      <GlassButton
                        variant="mint"
                        size="sm"
                        onClick={() => handleComplete(lesson.id)}
                        icon={<CheckCircle className="w-3.5 h-3.5" />}
                      >
                        Урок проведён
                      </GlassButton>
                      <GlassButton
                        variant="coral"
                        size="sm"
                        onClick={() => handleNoShow(lesson.id)}
                        icon={<AlertTriangle className="w-3.5 h-3.5" />}
                      >
                        Неявка
                      </GlassButton>
                    </>
                  )}
                </div>
              </GlassCard>
            ))}

            {dayLessons.length === 0 && !isLoading && (
              <div className="p-12 text-center liquid-glass rounded-3xl">
                <CalendarIcon className="w-12 h-12 text-slate-300 mx-auto mb-3" />
                <p className="text-base font-medium text-slate-700">На этот день уроков нет</p>
                <p className="text-xs text-slate-400 mt-1">
                  Нажмите «Назначить урок», чтобы запланировать занятие
                </p>
              </div>
            )}
          </div>
        )}
      </div>

      {/* ------------------------------------------------------------- */}
      {/* MODAL: НАЗНАЧЕНИЕ УРОКА */}
      {/* ------------------------------------------------------------- */}
      <GlassModal
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
          setRoomSearch('');
        }}
        title="Назначить занятие"
        description="Заполните параметры урока: ученик, время и формат проведения"
      >
        <form onSubmit={handleCreateLesson} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
              Ученик
            </label>
            <select
              value={selectedStudentId}
              onChange={(e) => setSelectedStudentId(e.target.value)}
              className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 liquid-glass-input focus:border-indigo-600 focus:outline-none"
              required
            >
              {students.map((rel) => (
                <option key={rel.student_id} value={rel.student_id}>
                  {rel.student_name}
                </option>
              ))}
            </select>
          </div>

          <GlassInput
            label="Тема / Название урока"
            placeholder="например: Подготовка к ОГЭ по математике"
            value={lessonTitle}
            onChange={(e) => setLessonTitle(e.target.value)}
            required
          />

          <div className="grid grid-cols-2 gap-3">
            <GlassInput
              label="Дата"
              type="date"
              value={lessonDate}
              onChange={(e) => setLessonDate(e.target.value)}
              required
            />
            <GlassInput
              label="Время начала"
              type="time"
              value={lessonStartTime}
              onChange={(e) => setLessonStartTime(e.target.value)}
              required
            />
          </div>

          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
              Длительность
            </label>
            <div className="grid grid-cols-4 gap-2">
              {['45', '60', '90', '120'].map((mins) => (
                <button
                  key={mins}
                  type="button"
                  onClick={() => setLessonDuration(mins)}
                  className={`py-2 rounded-xl text-xs font-semibold border transition-all ${
                    lessonDuration === mins
                      ? 'bg-indigo-600 text-white border-indigo-600 shadow-md shadow-indigo-600/20'
                      : 'liquid-glass text-slate-700 border-white/60 hover:bg-white/90'
                  }`}
                >
                  {mins} мин
                </button>
              ))}
            </div>
          </div>

          {/* Формат: Онлайн / Оффлайн */}
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
              Формат занятия
            </label>
            <div className="grid grid-cols-2 gap-3">
              <button
                type="button"
                onClick={() => setLessonFormat('offline')}
                className={`flex items-center justify-center gap-2 py-3 rounded-2xl border text-xs font-bold transition-all ${
                  lessonFormat === 'offline'
                    ? 'bg-indigo-600 text-white border-indigo-600 shadow-md shadow-indigo-600/20'
                    : 'liquid-glass text-slate-700 border-white/60 hover:bg-white/80'
                }`}
              >
                <MapPin className="w-4 h-4" />
                Оффлайн в школе
              </button>

              <button
                type="button"
                onClick={() => setLessonFormat('online')}
                className={`flex items-center justify-center gap-2 py-3 rounded-2xl border text-xs font-bold transition-all ${
                  lessonFormat === 'online'
                    ? 'bg-emerald-600 text-white border-emerald-600 shadow-md shadow-emerald-600/20'
                    : 'liquid-glass text-slate-700 border-white/60 hover:bg-white/80'
                }`}
              >
                <Video className="w-4 h-4" />
                Онлайн урок
              </button>
            </div>
          </div>

          {/* Оффлайн: Выбор кабинета */}
          {lessonFormat === 'offline' && (
            <div className="space-y-2">
              <div className="flex items-center justify-between ml-1">
                <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500">
                  Кабинет школы
                </label>
                <span className="text-[11px] text-slate-400 font-medium">
                  {classrooms.length} {classrooms.length === 1 ? 'аудитория' : 'доступно'}
                </span>
              </div>

              {classrooms.length > 4 && (
                <div className="relative">
                  <Search className="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
                  <input
                    type="text"
                    placeholder="Поиск по названию или вместимости..."
                    value={roomSearch}
                    onChange={(e) => setRoomSearch(e.target.value)}
                    className="w-full pl-8 pr-3 py-1.5 rounded-xl text-xs text-slate-800 liquid-glass-input focus:outline-none placeholder:text-slate-400"
                  />
                </div>
              )}

              {classrooms.length === 0 ? (
                <div className="p-3 text-center text-xs text-slate-400 liquid-glass rounded-2xl">
                  Кабинеты пока не добавлены администратором
                </div>
              ) : filteredClassrooms.length === 0 ? (
                <div className="p-3 text-center text-xs text-slate-400 liquid-glass rounded-2xl">
                  Кабинет по запросу «{roomSearch}» не найден
                </div>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 max-h-44 overflow-y-auto p-1.5 rounded-2xl bg-white/20 border border-white/50 custom-scrollbar">
                  {filteredClassrooms.map((room) => {
                    const isSelected = selectedClassroomId === room.id;
                    return (
                      <button
                        key={room.id}
                        type="button"
                        onClick={() => setSelectedClassroomId(room.id)}
                        className={`p-2.5 rounded-xl flex items-center gap-2.5 text-left transition-all border ${
                          isSelected
                            ? 'border-indigo-500 bg-white/95 shadow-sm ring-1 ring-indigo-500/30'
                            : 'liquid-glass border-white/40 hover:bg-white/70 hover:border-white/70'
                        }`}
                      >
                        <span
                          className="w-3.5 h-3.5 rounded-full shrink-0 shadow-xs"
                          style={{ backgroundColor: room.color }}
                        />
                        <div className="min-w-0 flex-1">
                          <span className="font-semibold text-xs text-slate-800 block truncate">
                            {room.name}
                          </span>
                          <span className="text-[10px] text-slate-500 block truncate">
                            до {room.capacity} чел
                          </span>
                        </div>
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
          )}

          {/* Онлайн: Ссылка на звонок */}
          {lessonFormat === 'online' && (
            <GlassInput
              label="Ссылка на видеоконференцию"
              placeholder="https://telemost.yandex.ru/... или Zoom"
              value={onlineLink}
              onChange={(e) => setOnlineLink(e.target.value)}
              required
            />
          )}

          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
              Комментарий для ученика
            </label>
            <textarea
              rows={2}
              placeholder="Взять с собой тетрадь, ручку или подготовить вопросы"
              value={lessonComment}
              onChange={(e) => setLessonComment(e.target.value)}
              className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 liquid-glass-input focus:border-indigo-600 focus:outline-none resize-none"
            />
          </div>

          <div className="flex items-center justify-end gap-3 pt-3 border-t border-black/[0.05]">
            <GlassButton
              type="button"
              variant="secondary"
              onClick={() => {
                setIsModalOpen(false);
                setRoomSearch('');
              }}
            >
              Отмена
            </GlassButton>
            <GlassButton
              type="submit"
              variant="primary"
              isLoading={isSubmitting}
            >
              Назначить урок
            </GlassButton>
          </div>
        </form>
      </GlassModal>
    </div>
  );
};
