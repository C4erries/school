import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { GlassCard } from '../../shared/components/GlassCard';
import { GlassButton } from '../../shared/components/GlassButton';
import { GlassInput } from '../../shared/components/GlassInput';
import { GlassModal } from '../../shared/components/GlassModal';
import { Badge } from '../../shared/components/Badge';
import { Classroom, Client, Lesson, LessonFormat } from '../../types/schedule';
import {
  getClassrooms,
  getClients,
  getLessons,
  createLesson,
  updateLesson,
  completeLesson,
  markNoShow,
  cancelLesson,
} from '../../api/schedule';
import {
  Calendar as CalendarIcon,
  Plus,
  Video,
  MapPin,
  Clock,
  AlertTriangle,
  ChevronLeft,
  ChevronRight,
  Columns,
  CalendarDays,
  List,
  Check,
  Edit2,
  Trash2,
  RefreshCw,
  AlertCircle,
} from 'lucide-react';

interface PositionedLesson extends Lesson {
  column: number;
  totalColumns: number;
  startMinutes: number;
  durationMinutes: number;
  endMinutes: number;
}

const START_HOUR = 8;
const END_HOUR = 22;
const HOUR_HEIGHT = 64; // px per hour slot

export const TeacherSchedulePage: React.FC = () => {
  const [clients, setClients] = useState<Client[]>([]);
  const [classrooms, setClassrooms] = useState<Classroom[]>([]);
  const [lessons, setLessons] = useState<Lesson[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Вид: 'day' | 'week' | 'list'
  const [viewMode, setViewMode] = useState<'day' | 'week' | 'list'>('week');
  const [currentDate, setCurrentDate] = useState<Date>(new Date());

  // Модалка создания урока
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [selectedClientId, setSelectedClientId] = useState('');
  const [lessonTitle, setLessonTitle] = useState('');
  const [lessonDate, setLessonDate] = useState(new Date().toISOString().split('T')[0]);
  const [lessonStartTime, setLessonStartTime] = useState('14:00');
  const [lessonDuration, setLessonDuration] = useState('60');
  const [lessonFormat, setLessonFormat] = useState<LessonFormat>('offline');
  const [selectedClassroomId, setSelectedClassroomId] = useState('');
  const [onlineLink, setOnlineLink] = useState('https://telemost.yandex.ru/j/school-lesson');
  const [lessonComment, setLessonComment] = useState('');
  const [isSubmittingCreate, setIsSubmittingCreate] = useState(false);

  // Модалка редактирования урока
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [editingLesson, setEditingLesson] = useState<Lesson | null>(null);
  const [editTitle, setEditTitle] = useState('');
  const [editDate, setEditDate] = useState('');
  const [editStartTime, setEditStartTime] = useState('');
  const [editEndTime, setEditEndTime] = useState('');
  const [editFormat, setEditFormat] = useState<LessonFormat>('offline');
  const [editClassroomId, setEditClassroomId] = useState('');
  const [editOnlineLink, setEditOnlineLink] = useState('');
  const [editComment, setEditComment] = useState('');
  const [cancelReason, setCancelReason] = useState('');
  const [isCancelling, setIsCancelling] = useState(false);
  const [isSubmittingEdit, setIsSubmittingEdit] = useState(false);

  const hours = useMemo(() => {
    const list: number[] = [];
    for (let h = START_HOUR; h <= END_HOUR; h++) {
      list.push(h);
    }
    return list;
  }, []);

  const loadData = useCallback(async () => {
    setIsLoading(true);
    setErrorMessage(null);
    try {
      const [cls, rms, les] = await Promise.all([
        getClients(),
        getClassrooms(),
        getLessons(),
      ]);
      setClients(cls);
      setClassrooms(rms);
      setLessons(les);
    } catch (err: unknown) {
      console.error('Failed to load schedule data', err);
      setErrorMessage(err instanceof Error ? err.message : 'Ошибка загрузки расписания');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Дни текущей недели (Пн – Вс)
  const weekDays = useMemo(() => {
    const d = new Date(currentDate);
    const dayOfWeek = d.getDay(); // 0 is Sun, 1 is Mon...
    const diff = d.getDate() - dayOfWeek + (dayOfWeek === 0 ? -6 : 1);
    const monday = new Date(d.setDate(diff));
    monday.setHours(0, 0, 0, 0);

    const days: Date[] = [];
    for (let i = 0; i < 7; i++) {
      const day = new Date(monday);
      day.setDate(monday.getDate() + i);
      days.push(day);
    }
    return days;
  }, [currentDate]);

  // Уроки на выбранный день (для DayView и ListView)
  const dayLessons = useMemo(() => {
    const targetDateStr = currentDate.toISOString().split('T')[0];
    return lessons.filter((l) => l.start_time.startsWith(targetDateStr));
  }, [lessons, currentDate]);

  // Алгоритм каскадного размещения нахлёстов для заданного списка уроков
  const positionLessons = useCallback((dayItems: Lesson[]): PositionedLesson[] => {
    if (dayItems.length === 0) return [];

    const sorted = [...dayItems].sort(
      (a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
    );

    const clusters: PositionedLesson[][] = [];
    let currentCluster: PositionedLesson[] = [];
    let clusterEnd = 0;

    for (const lesson of sorted) {
      const start = new Date(lesson.start_time);
      const end = new Date(lesson.end_time);
      const startMinutes = start.getHours() * 60 + start.getMinutes();
      const endMinutes = end.getHours() * 60 + end.getMinutes();
      const durationMinutes = Math.max(30, endMinutes - startMinutes);

      const item: PositionedLesson = {
        ...lesson,
        column: 0,
        totalColumns: 1,
        startMinutes,
        durationMinutes,
        endMinutes: startMinutes + durationMinutes,
      };

      if (currentCluster.length === 0) {
        currentCluster.push(item);
        clusterEnd = item.endMinutes;
      } else {
        if (startMinutes < clusterEnd) {
          currentCluster.push(item);
          clusterEnd = Math.max(clusterEnd, item.endMinutes);
        } else {
          clusters.push(currentCluster);
          currentCluster = [item];
          clusterEnd = item.endMinutes;
        }
      }
    }
    if (currentCluster.length > 0) {
      clusters.push(currentCluster);
    }

    const result: PositionedLesson[] = [];
    for (const cluster of clusters) {
      const columns: number[] = [];

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
  }, []);

  const positionedDayLessons = useMemo(() => {
    return positionLessons(dayLessons);
  }, [dayLessons, positionLessons]);

  // Навигация по датам
  const handleNavigate = (direction: -1 | 1) => {
    const next = new Date(currentDate);
    if (viewMode === 'week') {
      next.setDate(next.getDate() + direction * 7);
    } else {
      next.setDate(next.getDate() + direction);
    }
    setCurrentDate(next);
  };

  const handleGoToday = () => {
    setCurrentDate(new Date());
  };

  // Быстрый клик по свободному временному слоту
  const handleSlotClick = (date: Date, hour: number) => {
    const y = date.getFullYear();
    const m = String(date.getMonth() + 1).padStart(2, '0');
    const d = String(date.getDate()).padStart(2, '0');
    setLessonDate(`${y}-${m}-${d}`);
    setLessonStartTime(`${String(hour).padStart(2, '0')}:00`);
    setLessonDuration('60');
    if (!lessonTitle) setLessonTitle('Урок');
    setIsCreateModalOpen(true);
  };

  // Быстрое подтверждение проведения (кнопка ✓)
  const handleQuickComplete = async (e: React.MouseEvent, lessonId: string) => {
    e.stopPropagation();
    try {
      await completeLesson(lessonId);
      await loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка отметки проведения');
    }
  };

  // Открытие модалки редактирования
  const openEditModal = (lesson: Lesson) => {
    setEditingLesson(lesson);
    setEditTitle(lesson.title);
    const start = new Date(lesson.start_time);
    const end = new Date(lesson.end_time);
    setEditDate(start.toISOString().split('T')[0]);
    setEditStartTime(
      `${String(start.getHours()).padStart(2, '0')}:${String(start.getMinutes()).padStart(2, '0')}`
    );
    setEditEndTime(
      `${String(end.getHours()).padStart(2, '0')}:${String(end.getMinutes()).padStart(2, '0')}`
    );
    setEditFormat(lesson.format);
    setEditClassroomId(lesson.classroom_id || '');
    setEditOnlineLink(lesson.online_link || '');
    setEditComment(lesson.comment || '');
    setCancelReason('');
    setIsCancelling(false);
    setIsEditModalOpen(true);
  };

  // Сохранение изменений в уроке
  const handleSaveEdit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingLesson) return;

    setIsSubmittingEdit(true);
    try {
      const [startH, startM] = editStartTime.split(':').map(Number);
      const [endH, endM] = editEndTime.split(':').map(Number);
      const start = new Date(editDate);
      start.setHours(startH, startM, 0, 0);
      const end = new Date(editDate);
      end.setHours(endH, endM, 0, 0);

      await updateLesson(editingLesson.id, {
        title: editTitle.trim(),
        start_time: start.toISOString(),
        end_time: end.toISOString(),
        format: editFormat,
        classroom_id: editFormat === 'offline' ? editClassroomId || null : null,
        online_link: editFormat === 'online' ? editOnlineLink : undefined,
        comment: editComment.trim() || undefined,
      });

      setIsEditModalOpen(false);
      await loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка сохранения занятия');
    } finally {
      setIsSubmittingEdit(false);
    }
  };

  // Отмена урока
  const handleCancelLesson = async () => {
    if (!editingLesson) return;
    try {
      await cancelLesson(editingLesson.id, cancelReason.trim() || undefined);
      setIsEditModalOpen(false);
      await loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка отмены занятия');
    }
  };

  // Неявка ученика
  const handleNoShow = async () => {
    if (!editingLesson) return;
    if (!window.confirm('Отметить неявку ученика на занятие?')) return;
    try {
      await markNoShow(editingLesson.id);
      setIsEditModalOpen(false);
      await loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка отметки неявки');
    }
  };

  // Создание урока
  const handleCreateLesson = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedClientId || !lessonTitle.trim()) return;

    setIsSubmittingCreate(true);
    try {
      const [hoursVal, minsVal] = lessonStartTime.split(':').map(Number);
      const start = new Date(lessonDate);
      start.setHours(hoursVal, minsVal, 0, 0);

      const end = new Date(start.getTime() + Number(lessonDuration) * 60 * 1000);

      await createLesson({
        client_id: selectedClientId,
        title: lessonTitle.trim(),
        format: lessonFormat,
        classroom_id: lessonFormat === 'offline' ? selectedClassroomId || null : null,
        online_link: lessonFormat === 'online' ? onlineLink : undefined,
        comment: lessonComment.trim() || undefined,
        start_time: start.toISOString(),
        end_time: end.toISOString(),
      });

      setIsCreateModalOpen(false);
      setLessonTitle('');
      setLessonComment('');
      await loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка создания занятия');
    } finally {
      setIsSubmittingCreate(false);
    }
  };

  const getStatusBadge = (status: Lesson['status']) => {
    switch (status) {
      case 'scheduled':
      case 'confirmed':
        return <Badge variant="mint">Запланирован</Badge>;
      case 'pending_confirmation':
        return <Badge variant="amber">Ожидание</Badge>;
      case 'completed':
        return <Badge variant="indigo">Проведён</Badge>;
      case 'no_show':
        return <Badge variant="coral">Неявка</Badge>;
      case 'cancelled':
        return <Badge variant="neutral">Отменён</Badge>;
      default:
        return <Badge variant="neutral">{status}</Badge>;
    }
  };

  // Поиск имени клиента по ID или привязке
  const getClientDisplayName = (lesson: Lesson): string => {
    if (lesson.client_name) return lesson.client_name;
    if (lesson.student_name) return lesson.student_name;
    const found = clients.find((c) => c.id === lesson.client_id || c.id === lesson.student_id);
    return found ? found.name : 'Ученик';
  };

  return (
    <div className="space-y-6">
      {/* Заголовок и панель действий */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">
            Расписание уроков
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Календарь занятий, нахлёсты, быстрое создание в 1 клик и отметка проведения
          </p>
        </div>

        <GlassButton
          variant="primary"
          onClick={() => {
            const now = new Date();
            setLessonDate(now.toISOString().split('T')[0]);
            setLessonStartTime('14:00');
            setLessonDuration('60');
            setIsCreateModalOpen(true);
          }}
          icon={<Plus className="w-4 h-4" />}
        >
          Назначить урок
        </GlassButton>
      </div>

      {/* Ошибка подключения */}
      {errorMessage && (
        <div className="p-4 rounded-3xl bg-rose-500/10 border border-rose-500/20 backdrop-blur-md flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
            <p className="text-sm font-medium text-rose-900">{errorMessage}</p>
          </div>
          <GlassButton
            variant="secondary"
            size="sm"
            onClick={loadData}
            icon={<RefreshCw className="w-3.5 h-3.5" />}
          >
            Повторить запрос
          </GlassButton>
        </div>
      )}

      {/* Быстрое назначение ученику */}
      {clients.length > 0 && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-500">
              Быстрое назначение ({clients.length})
            </span>
          </div>
          <div className="flex items-center gap-3 overflow-x-auto pb-2 scrollbar-none">
            {clients.map((client) => (
              <div
                key={client.id}
                onClick={() => {
                  setSelectedClientId(client.id);
                  setLessonTitle(`Урок: ${client.name}`);
                  setIsCreateModalOpen(true);
                }}
                className="p-3 rounded-2xl bg-white/30 backdrop-blur-md hover:bg-white/60 border border-white/40 flex items-center gap-3 shrink-0 cursor-pointer transition-all shadow-sm hover:scale-[1.02]"
              >
                <div className="w-8 h-8 rounded-xl bg-indigo-500/15 border border-indigo-400/30 text-indigo-700 flex items-center justify-center font-bold text-xs shadow-sm">
                  {client.name.charAt(0)}
                </div>
                <div className="text-left pr-2">
                  <div className="text-xs font-bold text-slate-900 leading-tight">
                    {client.name}
                  </div>
                  <div className="text-[10px] text-indigo-600 font-semibold">
                    + Запланировать
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Панель навигации по календарю и переключатель режимов [ День | Неделя | Список ] */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4 p-3 rounded-3xl bg-white/30 backdrop-blur-md border border-white/40 shadow-sm">
        {/* Стрелки переключения и текущий период */}
        <div className="flex items-center gap-2">
          <button
            onClick={() => handleNavigate(-1)}
            className="p-2 rounded-xl hover:bg-black/5 text-slate-600 transition-colors"
            title="Назад"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>

          <div className="flex items-center gap-2 px-3.5 py-1.5 rounded-xl bg-white/60 border border-white/80 shadow-sm">
            <CalendarIcon className="w-4 h-4 text-indigo-600" />
            <span className="font-bold text-sm text-slate-900 capitalize">
              {viewMode === 'week' ? (
                <>
                  {weekDays[0].toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })} –{' '}
                  {weekDays[6].toLocaleDateString('ru-RU', {
                    day: 'numeric',
                    month: 'short',
                    year: 'numeric',
                  })}
                </>
              ) : (
                currentDate.toLocaleDateString('ru-RU', {
                  weekday: 'short',
                  day: 'numeric',
                  month: 'long',
                  year: 'numeric',
                })
              )}
            </span>
          </div>

          <button
            onClick={() => handleNavigate(1)}
            className="p-2 rounded-xl hover:bg-black/5 text-slate-600 transition-colors"
            title="Вперёд"
          >
            <ChevronRight className="w-4 h-4" />
          </button>

          <button
            onClick={handleGoToday}
            className="text-xs font-semibold px-3 py-1.5 rounded-xl text-indigo-600 hover:bg-indigo-50 transition-colors ml-1"
          >
            Сегодня
          </button>
        </div>

        {/* Переключатель вида [ День | Неделя | Список ] в стиле Apple Segmented Control */}
        <div className="flex items-center gap-1.5 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05] self-end sm:self-auto">
          <button
            onClick={() => setViewMode('day')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              viewMode === 'day'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <Columns className="w-3.5 h-3.5" />
            <span>День</span>
          </button>

          <button
            onClick={() => setViewMode('week')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
              viewMode === 'week'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <CalendarDays className="w-3.5 h-3.5" />
            <span>Неделя</span>
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
            <span>Список</span>
          </button>
        </div>
      </div>

      {/* ============================================================= */}
      {/* РЕЖИМ: НЕДЕЛЬНЫЙ ВИД (WEEK VIEW)                              */}
      {/* ============================================================= */}
      {viewMode === 'week' && (
        <GlassCard className="p-3 sm:p-5 overflow-x-auto">
          <div className="min-w-[760px] select-none">
            {/* Шапка 7 колонок дней недели */}
            <div className="grid grid-cols-[56px_repeat(7,1fr)] gap-0 border-b border-slate-200/60 pb-2 mb-2">
              <div />
              {weekDays.map((day) => {
                const isToday =
                  day.toDateString() === new Date().toDateString();
                const isSelected =
                  day.toDateString() === currentDate.toDateString();

                return (
                  <div
                    key={day.toISOString()}
                    onClick={() => setCurrentDate(day)}
                    className={`text-center py-2 px-1 rounded-2xl cursor-pointer transition-all ${
                      isSelected
                        ? 'bg-indigo-600/10 border border-indigo-500/20'
                        : 'hover:bg-white/40'
                    }`}
                  >
                    <span className="block text-[11px] font-semibold text-slate-500 uppercase tracking-wider">
                      {day.toLocaleDateString('ru-RU', { weekday: 'short' })}
                    </span>
                    <span
                      className={`inline-flex items-center justify-center w-7 h-7 mt-0.5 rounded-full text-sm font-bold ${
                        isToday
                          ? 'bg-indigo-600 text-white shadow-sm'
                          : 'text-slate-800'
                      }`}
                    >
                      {day.getDate()}
                    </span>
                  </div>
                );
              })}
            </div>

            {/* Сетка времени и колонок */}
            <div className="relative grid grid-cols-[56px_repeat(7,1fr)] gap-0">
              {/* Левая шкала времени 08:00 – 22:00 */}
              <div className="relative">
                {hours.map((hour) => (
                  <div
                    key={hour}
                    className="border-b border-transparent text-right pr-3 text-xs font-mono text-slate-400 -mt-2.5"
                    style={{ height: `${HOUR_HEIGHT}px` }}
                  >
                    {String(hour).padStart(2, '0')}:00
                  </div>
                ))}
              </div>

              {/* 7 колонок дней недели */}
              {weekDays.map((day) => {
                const dayStr = day.toISOString().split('T')[0];
                const dayItems = lessons.filter((l) => l.start_time.startsWith(dayStr));
                const positionedItems = positionLessons(dayItems);

                return (
                  <div
                    key={day.toISOString()}
                    className="relative border-l border-slate-200/60"
                    style={{ height: `${hours.length * HOUR_HEIGHT}px` }}
                  >
                    {/* Фоновые горизонтальные линии и кликабельные слоты */}
                    {hours.map((hour) => (
                      <div
                        key={hour}
                        onClick={() => handleSlotClick(day, hour)}
                        className="border-b border-slate-200/50 hover:bg-indigo-50/30 cursor-pointer transition-colors"
                        style={{ height: `${HOUR_HEIGHT}px` }}
                        title={`Назначить на ${day.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })} в ${hour}:00`}
                      />
                    ))}

                    {/* Размещение уроков в колонке дня */}
                    {positionedItems.map((lesson) => {
                      const topOffset =
                        ((lesson.startMinutes - START_HOUR * 60) / 60) * HOUR_HEIGHT;
                      const cardHeight = (lesson.durationMinutes / 60) * HOUR_HEIGHT;
                      const widthPercent = 100 / lesson.totalColumns;
                      const leftPercent = lesson.column * widthPercent;
                      const studentName = getClientDisplayName(lesson);
                      const isCompleted = lesson.status === 'completed';

                      return (
                        <div
                          key={lesson.id}
                          onClick={() => openEditModal(lesson)}
                          className="absolute p-0.5 pointer-events-auto transition-all group z-10 cursor-pointer"
                          style={{
                            top: `${Math.max(0, topOffset)}px`,
                            height: `${Math.max(48, cardHeight)}px`,
                            left: `${leftPercent}%`,
                            width: `${widthPercent}%`,
                          }}
                        >
                          <div
                            className={`h-full w-full rounded-2xl p-2 sm:p-2.5 flex flex-col justify-between shadow-sm relative overflow-hidden transition-all group-hover:scale-[1.02] group-hover:shadow-md border ${
                              isCompleted
                                ? 'bg-emerald-500/20 border-emerald-400/50 text-emerald-950'
                                : 'bg-white/85 backdrop-blur-md border-white/80'
                            }`}
                            style={{
                              borderLeftWidth: '4px',
                              borderLeftColor:
                                lesson.classroom_color ||
                                (lesson.format === 'online' ? '#10B981' : '#4F46E5'),
                            }}
                          >
                            <div className="min-w-0">
                              {/* 1-я строка: Крупное имя ученика */}
                              <div className="flex items-start justify-between gap-1">
                                <span className="font-bold text-xs sm:text-sm text-slate-900 leading-tight truncate">
                                  {studentName}
                                </span>

                                {/* Кнопка быстрого подтверждения проведения ✓ */}
                                {lesson.status === 'scheduled' && (
                                  <button
                                    type="button"
                                    onClick={(e) => handleQuickComplete(e, lesson.id)}
                                    title="Отметить проведённым"
                                    className="w-5 h-5 rounded-md bg-emerald-500 hover:bg-emerald-600 text-white flex items-center justify-center shrink-0 transition-all shadow-sm active:scale-95"
                                  >
                                    <Check className="w-3 h-3 stroke-[3]" />
                                  </button>
                                )}
                              </div>

                              {/* 2-я строка: Время и формат */}
                              <div className="text-[10px] text-slate-600 font-mono mt-0.5 truncate">
                                {new Date(lesson.start_time).toLocaleTimeString('ru-RU', {
                                  hour: '2-digit',
                                  minute: '2-digit',
                                })}{' '}
                                –{' '}
                                {new Date(lesson.end_time).toLocaleTimeString('ru-RU', {
                                  hour: '2-digit',
                                  minute: '2-digit',
                                })}
                              </div>

                              {/* 3-я строка: Название предмета или кабинет */}
                              {lesson.classroom_name ? (
                                <div className="text-[10px] text-slate-500 truncate flex items-center gap-1 mt-0.5">
                                  <MapPin className="w-2.5 h-2.5 text-slate-400 shrink-0" />
                                  <span className="truncate">{lesson.classroom_name}</span>
                                </div>
                              ) : lesson.format === 'online' ? (
                                <div className="text-[10px] text-emerald-600 font-medium truncate flex items-center gap-1 mt-0.5">
                                  <Video className="w-2.5 h-2.5 shrink-0" />
                                  <span>Онлайн</span>
                                </div>
                              ) : null}
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                );
              })}
            </div>
          </div>
        </GlassCard>
      )}

      {/* ============================================================= */}
      {/* РЕЖИМ: ОДИН ДЕНЬ (DAY VIEW)                                   */}
      {/* ============================================================= */}
      {viewMode === 'day' && (
        <GlassCard className="p-4 sm:p-6 overflow-hidden">
          <div className="relative border-t border-slate-200/60 select-none">
            {/* Фоновые линии шкалы времени */}
            <div className="relative">
              {hours.map((hour) => (
                <div
                  key={hour}
                  onClick={() => handleSlotClick(currentDate, hour)}
                  className="flex items-start border-b border-slate-200/50 hover:bg-indigo-50/20 cursor-pointer transition-colors"
                  style={{ height: `${HOUR_HEIGHT}px` }}
                >
                  <div className="w-16 text-right pr-4 text-xs font-mono text-slate-400 -mt-2.5">
                    {String(hour).padStart(2, '0')}:00
                  </div>
                  <div className="flex-1 h-full border-l border-slate-200/60" />
                </div>
              ))}
            </div>

            {/* Карточки уроков в один день с каскадным нахлёстом */}
            <div
              className="absolute top-0 right-0 left-16 bottom-0 pointer-events-none"
              style={{ height: `${hours.length * HOUR_HEIGHT}px` }}
            >
              {positionedDayLessons.map((lesson) => {
                const topOffset =
                  ((lesson.startMinutes - START_HOUR * 60) / 60) * HOUR_HEIGHT;
                const cardHeight = (lesson.durationMinutes / 60) * HOUR_HEIGHT;
                const widthPercent = 100 / lesson.totalColumns;
                const leftPercent = lesson.column * widthPercent;
                const accentColor =
                  lesson.classroom_color ||
                  (lesson.format === 'online' ? '#10B981' : '#4F46E5');
                const studentName = getClientDisplayName(lesson);
                const isCompleted = lesson.status === 'completed';

                return (
                  <div
                    key={lesson.id}
                    onClick={() => openEditModal(lesson)}
                    className="absolute p-1 pointer-events-auto transition-all group z-10 cursor-pointer"
                    style={{
                      top: `${Math.max(0, topOffset)}px`,
                      height: `${Math.max(50, cardHeight)}px`,
                      left: `${leftPercent}%`,
                      width: `${widthPercent}%`,
                    }}
                  >
                    <div
                      className={`h-full w-full rounded-2xl p-3 flex flex-col justify-between shadow-md relative overflow-hidden transition-all group-hover:scale-[1.01] group-hover:shadow-lg border ${
                        isCompleted
                          ? 'bg-emerald-500/20 border-emerald-400/50'
                          : 'bg-white/85 backdrop-blur-md border-white/80'
                      }`}
                      style={{
                        borderLeftWidth: '5px',
                        borderLeftColor: accentColor,
                      }}
                    >
                      <div>
                        {/* 1-я строка: Крупное имя ученика */}
                        <div className="flex items-center justify-between gap-2">
                          <span className="font-bold text-sm sm:text-base text-slate-900 truncate">
                            {studentName}
                          </span>
                          <div className="flex items-center gap-1.5">
                            {getStatusBadge(lesson.status)}
                            {lesson.status === 'scheduled' && (
                              <button
                                type="button"
                                onClick={(e) => handleQuickComplete(e, lesson.id)}
                                title="Отметить проведённым"
                                className="w-6 h-6 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-white flex items-center justify-center transition-all shadow-sm active:scale-95"
                              >
                                <Check className="w-3.5 h-3.5 stroke-[3]" />
                              </button>
                            )}
                          </div>
                        </div>

                        {/* 2-я строка: Время и тема */}
                        <div className="flex items-center gap-2 text-xs text-slate-600 mt-1 truncate">
                          <span className="font-mono font-medium">
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
                          <span>•</span>
                          <span className="truncate">{lesson.title}</span>
                        </div>

                        {/* 3-я строка: Формат и кабинет */}
                        {lesson.format === 'offline' && lesson.classroom_name && (
                          <div className="flex items-center gap-1 text-xs font-medium text-slate-600 mt-1">
                            <MapPin className="w-3.5 h-3.5 text-slate-400" />
                            <span className="truncate">{lesson.classroom_name}</span>
                          </div>
                        )}
                        {lesson.format === 'online' && (
                          <div className="flex items-center gap-1 text-xs font-medium text-emerald-600 mt-1">
                            <Video className="w-3.5 h-3.5" />
                            <span>Онлайн занятие</span>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </GlassCard>
      )}

      {/* ============================================================= */}
      {/* РЕЖИМ: СПИСОК УРОКОВ (LIST VIEW)                              */}
      {/* ============================================================= */}
      {viewMode === 'list' && (
        <div className="space-y-3">
          {dayLessons.map((lesson) => {
            const studentName = getClientDisplayName(lesson);
            return (
              <GlassCard
                key={lesson.id}
                onClick={() => openEditModal(lesson)}
                className="p-5 flex flex-col md:flex-row items-start md:items-center justify-between gap-4 cursor-pointer hover:shadow-md transition-all"
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
                      {/* Крупное имя ученика */}
                      <h3 className="font-bold text-slate-900 text-base">
                        {studentName}
                      </h3>
                      {getStatusBadge(lesson.status)}
                    </div>
                    <div className="flex items-center gap-3 text-xs text-slate-500 mt-1 flex-wrap">
                      <span className="font-semibold text-slate-800">{lesson.title}</span>
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
                  </div>
                </div>

                <div className="flex items-center gap-2 self-end md:self-auto">
                  {lesson.status === 'scheduled' && (
                    <GlassButton
                      variant="mint"
                      size="sm"
                      onClick={(e) => handleQuickComplete(e, lesson.id)}
                      icon={<Check className="w-3.5 h-3.5" />}
                    >
                      Проведён
                    </GlassButton>
                  )}
                  <GlassButton
                    variant="secondary"
                    size="sm"
                    onClick={(e) => {
                      e.stopPropagation();
                      openEditModal(lesson);
                    }}
                    icon={<Edit2 className="w-3.5 h-3.5" />}
                  >
                    Изменить
                  </GlassButton>
                </div>
              </GlassCard>
            );
          })}

          {dayLessons.length === 0 && !isLoading && (
            <div className="p-12 text-center liquid-glass rounded-3xl">
              <CalendarIcon className="w-12 h-12 text-slate-300 mx-auto mb-3" />
              <p className="text-base font-medium text-slate-700">На этот день уроков нет</p>
              <p className="text-xs text-slate-400 mt-1">
                Кликните на свободное время в сетке или нажмите «Назначить урок»
              </p>
            </div>
          )}
        </div>
      )}

      {/* ============================================================= */}
      {/* МОДАЛКА: СОЗДАНИЕ УРОКА (В 1 КЛИК ПО ТАЙМСЛОТУ)               */}
      {/* ============================================================= */}
      <GlassModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="Назначить занятие"
        description="Выберите ученика, дату, формат и кабинет школы"
        maxWidth="lg"
      >
        <form onSubmit={handleCreateLesson} className="space-y-4">
          {/* Выбор ученика */}
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
              Ученик *
            </label>
            <select
              value={selectedClientId}
              onChange={(e) => setSelectedClientId(e.target.value)}
              className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 bg-white/70 border border-slate-200 focus:outline-none focus:border-indigo-500"
              required
            >
              <option value="">-- Выберите ученика --</option>
              {clients.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name} ({c.balance ?? 0} ч на балансе)
                </option>
              ))}
            </select>
          </div>

          {/* Тема урока */}
          <GlassInput
            label="Тема / Название занятия *"
            placeholder="например: Подготовка к ОГЭ: Алгебра"
            value={lessonTitle}
            onChange={(e) => setLessonTitle(e.target.value)}
            required
          />

          {/* Дата и время */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <GlassInput
              label="Дата *"
              type="date"
              value={lessonDate}
              onChange={(e) => setLessonDate(e.target.value)}
              required
            />
            <GlassInput
              label="Время начала *"
              type="time"
              value={lessonStartTime}
              onChange={(e) => setLessonStartTime(e.target.value)}
              required
            />
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
                Длительность
              </label>
              <select
                value={lessonDuration}
                onChange={(e) => setLessonDuration(e.target.value)}
                className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 bg-white/70 border border-slate-200 focus:outline-none focus:border-indigo-500"
              >
                <option value="45">45 минут (0.75 ч)</option>
                <option value="60">60 минут (1.0 ч)</option>
                <option value="90">90 минут (1.5 ч)</option>
                <option value="120">120 минут (2.0 ч)</option>
              </select>
            </div>
          </div>

          {/* Формат занятия */}
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
              Формат проведения
            </label>
            <div className="grid grid-cols-2 gap-2 p-1 rounded-2xl bg-black/[0.04]">
              <button
                type="button"
                onClick={() => setLessonFormat('offline')}
                className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                  lessonFormat === 'offline'
                    ? 'bg-white text-indigo-600 shadow-sm'
                    : 'text-slate-600'
                }`}
              >
                Оффлайн в школе
              </button>
              <button
                type="button"
                onClick={() => setLessonFormat('online')}
                className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                  lessonFormat === 'online'
                    ? 'bg-white text-indigo-600 shadow-sm'
                    : 'text-slate-600'
                }`}
              >
                Онлайн урок
              </button>
            </div>
          </div>

          {/* Кабинет (если оффлайн) */}
          {lessonFormat === 'offline' && (
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
                Аудитория школы
              </label>
              <select
                value={selectedClassroomId}
                onChange={(e) => setSelectedClassroomId(e.target.value)}
                className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 bg-white/70 border border-slate-200 focus:outline-none focus:border-indigo-500"
              >
                <option value="">Без закрепления кабинета</option>
                {classrooms.map((room) => (
                  <option key={room.id} value={room.id}>
                    {room.name} (вместимость: {room.capacity})
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Ссылка (если онлайн) */}
          {lessonFormat === 'online' && (
            <GlassInput
              label="Ссылка на видеозвонок"
              placeholder="https://telemost.yandex.ru/j/..."
              value={onlineLink}
              onChange={(e) => setOnlineLink(e.target.value)}
            />
          )}

          <GlassInput
            label="Заметка / ДЗ (необязательно)"
            placeholder="например: Проверить вариант №4"
            value={lessonComment}
            onChange={(e) => setLessonComment(e.target.value)}
          />

          <div className="pt-3 flex justify-end gap-3 border-t border-black/[0.05]">
            <GlassButton
              type="button"
              variant="secondary"
              onClick={() => setIsCreateModalOpen(false)}
            >
              Отмена
            </GlassButton>
            <GlassButton
              type="submit"
              variant="primary"
              isLoading={isSubmittingCreate}
            >
              Запланировать урок
            </GlassButton>
          </div>
        </form>
      </GlassModal>

      {/* ============================================================= */}
      {/* МОДАЛКА: РЕДАКТИРОВАНИЕ / ОТМЕНА УРОКА                        */}
      {/* ============================================================= */}
      <GlassModal
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        title={editingLesson ? `Урок: ${getClientDisplayName(editingLesson)}` : 'Редактирование урока'}
        description="Измените время, аудиторию или отмените занятие."
        maxWidth="lg"
      >
        {editingLesson && (
          <form onSubmit={handleSaveEdit} className="space-y-4">
            <GlassInput
              label="Тема занятия"
              value={editTitle}
              onChange={(e) => setEditTitle(e.target.value)}
              required
            />

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <GlassInput
                label="Дата"
                type="date"
                value={editDate}
                onChange={(e) => setEditDate(e.target.value)}
                required
              />
              <GlassInput
                label="Начало"
                type="time"
                value={editStartTime}
                onChange={(e) => setEditStartTime(e.target.value)}
                required
              />
              <GlassInput
                label="Конец"
                type="time"
                value={editEndTime}
                onChange={(e) => setEditEndTime(e.target.value)}
                required
              />
            </div>

            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
                Формат
              </label>
              <div className="grid grid-cols-2 gap-2 p-1 rounded-2xl bg-black/[0.04]">
                <button
                  type="button"
                  onClick={() => setEditFormat('offline')}
                  className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                    editFormat === 'offline'
                      ? 'bg-white text-indigo-600 shadow-sm'
                      : 'text-slate-600'
                  }`}
                >
                  Оффлайн в кабинете
                </button>
                <button
                  type="button"
                  onClick={() => setEditFormat('online')}
                  className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                    editFormat === 'online'
                      ? 'bg-white text-indigo-600 shadow-sm'
                      : 'text-slate-600'
                  }`}
                >
                  Онлайн урок
                </button>
              </div>
            </div>

            {editFormat === 'offline' ? (
              <div>
                <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
                  Кабинет
                </label>
                <select
                  value={editClassroomId}
                  onChange={(e) => setEditClassroomId(e.target.value)}
                  className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 bg-white/70 border border-slate-200 focus:outline-none focus:border-indigo-500"
                >
                  <option value="">Без закрепления кабинета</option>
                  {classrooms.map((room) => (
                    <option key={room.id} value={room.id}>
                      {room.name}
                    </option>
                  ))}
                </select>
              </div>
            ) : (
              <GlassInput
                label="Ссылка на созвон"
                value={editOnlineLink}
                onChange={(e) => setEditOnlineLink(e.target.value)}
              />
            )}

            <GlassInput
              label="Заметка"
              value={editComment}
              onChange={(e) => setEditComment(e.target.value)}
            />

            {/* Блок отмены урока */}
            {editingLesson.status === 'scheduled' && (
              <div className="p-3.5 rounded-2xl bg-rose-500/10 border border-rose-500/20 space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold text-rose-800 flex items-center gap-1.5">
                    <Trash2 className="w-3.5 h-3.5" /> Отмена занятия
                  </span>
                  {!isCancelling && (
                    <div className="flex items-center gap-3">
                      <button
                        type="button"
                        onClick={handleNoShow}
                        className="text-xs font-semibold text-amber-700 hover:text-amber-800 flex items-center gap-1"
                      >
                        <AlertTriangle className="w-3.5 h-3.5" />
                        Неявка
                      </button>
                      <button
                        type="button"
                        onClick={() => setIsCancelling(true)}
                        className="text-xs font-semibold text-rose-600 hover:text-rose-700 underline"
                      >
                        Отменить этот урок
                      </button>
                    </div>
                  )}
                </div>

                {isCancelling && (
                  <div className="space-y-2 pt-1 animate-in fade-in">
                    <input
                      type="text"
                      placeholder="Укажите причину отмены (необязательно)"
                      value={cancelReason}
                      onChange={(e) => setCancelReason(e.target.value)}
                      className="w-full rounded-xl px-3 py-2 text-xs text-slate-800 bg-white/80 border border-rose-200 focus:outline-none"
                    />
                    <div className="flex justify-end gap-2">
                      <button
                        type="button"
                        onClick={() => setIsCancelling(false)}
                        className="px-3 py-1 rounded-lg text-xs font-medium text-slate-600 hover:bg-black/5"
                      >
                        Назад
                      </button>
                      <button
                        type="button"
                        onClick={handleCancelLesson}
                        className="px-3 py-1 rounded-lg text-xs font-bold text-white bg-rose-500 hover:bg-rose-600 shadow-sm"
                      >
                        Подтвердить отмену
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}

            <div className="pt-3 flex justify-end gap-3 border-t border-black/[0.05]">
              <GlassButton
                type="button"
                variant="secondary"
                onClick={() => setIsEditModalOpen(false)}
              >
                Закрыть
              </GlassButton>
              <GlassButton
                type="submit"
                variant="primary"
                isLoading={isSubmittingEdit}
              >
                Сохранить
              </GlassButton>
            </div>
          </form>
        )}
      </GlassModal>
    </div>
  );
};

export default TeacherSchedulePage;
