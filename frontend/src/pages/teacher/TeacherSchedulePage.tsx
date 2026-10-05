import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { GlassCard } from '../../shared/components/GlassCard';
import { GlassButton } from '../../shared/components/GlassButton';
import { GlassInput } from '../../shared/components/GlassInput';
import { GlassModal } from '../../shared/components/GlassModal';
import { Badge } from '../../shared/components/Badge';
import { AppNavbar } from '../../shared/components/AppNavbar';
import { LiquidBackground } from '../../shared/components/LiquidBackground';
import { Classroom, Client, Lesson, LessonFormat } from '../../types/schedule';
import {
  getClassrooms,
  getClients,
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
  const [clients, setClients] = useState<Client[]>([]);
  const [classrooms, setClassrooms] = useState<Classroom[]>([]);
  const [lessons, setLessons] = useState<Lesson[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const [selectedDate, setSelectedDate] = useState<Date>(new Date());
  const [viewMode, setViewMode] = useState<'timeline' | 'list'>('timeline');

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedClientId, setSelectedClientId] = useState('');
  const [lessonTitle, setLessonTitle] = useState('');
  const [lessonDate, setLessonDate] = useState(new Date().toISOString().split('T')[0]);
  const [lessonStartTime, setLessonStartTime] = useState('14:00');
  const [lessonDuration, setLessonDuration] = useState('60');
  const [lessonFormat, setLessonFormat] = useState<LessonFormat>('offline');
  const [selectedClassroomId, setSelectedClassroomId] = useState('');
  const [onlineLink, setOnlineLink] = useState('https://telemost.yandex.ru/j/school-');
  const [lessonComment, setLessonComment] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [roomSearch, setRoomSearch] = useState('');

  const filteredClassrooms = useMemo(() => {
    if (!roomSearch.trim()) return classrooms;
    const q = roomSearch.toLowerCase();
    return classrooms.filter(
      (c) => c.name.toLowerCase().includes(q) || String(c.capacity).includes(q)
    );
  }, [classrooms, roomSearch]);

  const loadData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [allRooms, myClients, allLessons] = await Promise.all([
        getClassrooms(),
        getClients(),
        getLessons(),
      ]);

      setClassrooms(allRooms);
      setClients(myClients);
      setLessons(allLessons);

      if (myClients.length > 0 && !selectedClientId) {
        setSelectedClientId(myClients[0].id);
      }
      if (allRooms.length > 0 && !selectedClassroomId) {
        setSelectedClassroomId(allRooms[0].id);
      }
    } catch (err) {
      console.error('Failed to load schedule', err);
    } finally {
      setIsLoading(false);
    }
  }, [selectedClassroomId, selectedClientId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const selectedDateStr = useMemo(() => {
    const y = selectedDate.getFullYear();
    const m = String(selectedDate.getMonth() + 1).padStart(2, '0');
    const d = String(selectedDate.getDate()).padStart(2, '0');
    return `${y}-${m}-${d}`;
  }, [selectedDate]);

  const dayLessons = useMemo(() => {
    return lessons.filter((l) => l.start_time.startsWith(selectedDateStr));
  }, [lessons, selectedDateStr]);

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

    parsed.sort((a, b) => a.startMinutes - b.startMinutes || b.durationMinutes - a.durationMinutes);

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
  }, [dayLessons]);

  const handleComplete = async (lessonId: string) => {
    try {
      await completeLesson(lessonId);
      await loadData();
    } catch (err) {
      console.error('Error completing lesson', err);
    }
  };

  const handleNoShow = async (lessonId: string) => {
    if (!window.confirm('Отметить неявку ученика на занятие?')) return;
    try {
      await markNoShow(lessonId);
      await loadData();
    } catch (err) {
      console.error('Error marking no show', err);
    }
  };

  const handleCreateLesson = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedClientId || !lessonTitle.trim()) return;

    setIsSubmitting(true);
    try {
      const [hours, mins] = lessonStartTime.split(':').map(Number);
      const start = new Date(lessonDate);
      start.setHours(hours, mins, 0, 0);

      const end = new Date(start.getTime() + Number(lessonDuration) * 60 * 1000);

      await createLesson({
        client_id: selectedClientId,
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
      case 'scheduled':
        return <Badge variant="mint">Запланирован</Badge>;
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

  const changeDateBy = (days: number) => {
    const next = new Date(selectedDate);
    next.setDate(next.getDate() + days);
    setSelectedDate(next);
  };

  const hours = Array.from({ length: 13 }, (_, i) => i + 8);
  const START_HOUR = 8;
  const HOUR_HEIGHT = 70;

  return (
    <div className="min-h-screen text-slate-800 p-4 sm:p-8 relative overflow-hidden">
      <LiquidBackground />
      <div className="max-w-6xl mx-auto space-y-6 relative z-10">
        <AppNavbar />

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
                  setIsModalOpen(true);
                }}
                className="liquid-glass hover:bg-white/90 p-3 rounded-2xl flex items-center gap-3 shrink-0 cursor-pointer border border-white/60 transition-all shadow-sm hover:scale-[1.02]"
              >
                <div className="w-8 h-8 rounded-xl bg-indigo-50 border border-indigo-200/50 text-indigo-600 flex items-center justify-center font-bold text-xs">
                  {client.name.charAt(0)}
                </div>
                <div className="text-left pr-2">
                  <div className="text-xs font-bold text-slate-900 leading-tight">
                    {client.name}
                  </div>
                  <div className="text-[10px] text-indigo-600 font-medium">
                    + Запланировать
                  </div>
                </div>
              </div>
            ))}
            {clients.length === 0 && (
              <div className="text-xs text-slate-400 p-2">
                Ученики не добавлены. Перейдите в CRM для добавления.
              </div>
            )}
          </div>
        </div>

        <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4 p-3 rounded-3xl liquid-glass border border-white/60 shadow-sm">
          <div className="flex items-center gap-2">
            <button onClick={() => changeDateBy(-1)} className="p-2 rounded-xl hover:bg-black/5 text-slate-600 transition-colors">
              <ChevronLeft className="w-4 h-4" />
            </button>
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-white/70 border border-white/80">
              <CalendarIcon className="w-4 h-4 text-indigo-600" />
              <span className="font-bold text-sm text-slate-900 capitalize">
                {selectedDate.toLocaleDateString('ru-RU', { weekday: 'short', day: 'numeric', month: 'long', year: 'numeric' })}
              </span>
            </div>
            <button onClick={() => changeDateBy(1)} className="p-2 rounded-xl hover:bg-black/5 text-slate-600 transition-colors">
              <ChevronRight className="w-4 h-4" />
            </button>
            <button onClick={() => setSelectedDate(new Date())} className="text-xs font-semibold px-3 py-1.5 rounded-xl text-indigo-600 hover:bg-indigo-50 transition-colors ml-1">
              Сегодня
            </button>
          </div>

          <div className="flex items-center gap-1.5 p-1 rounded-2xl bg-black/[0.04] self-end sm:self-auto">
            <button onClick={() => setViewMode('timeline')} className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${viewMode === 'timeline' ? 'bg-white text-indigo-600 shadow-sm' : 'text-slate-600 hover:text-slate-900'}`}>
              <Columns className="w-3.5 h-3.5" />
              <span>Сетка</span>
            </button>
            <button onClick={() => setViewMode('list')} className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${viewMode === 'list' ? 'bg-white text-indigo-600 shadow-sm' : 'text-slate-600 hover:text-slate-900'}`}>
              <List className="w-3.5 h-3.5" />
              <span>Список ({dayLessons.length})</span>
            </button>
          </div>
        </div>

        {viewMode === 'timeline' && (
          <GlassCard className="p-4 sm:p-6 overflow-hidden">
            <div className="relative border-t border-slate-200/60 select-none">
              <div className="relative">
                {hours.map((hour) => (
                  <div key={hour} className="flex items-start border-b border-slate-200/50" style={{ height: `${HOUR_HEIGHT}px` }}>
                    <div className="w-16 text-right pr-4 text-xs font-mono text-slate-400 -mt-2.5">
                      {String(hour).padStart(2, '0')}:00
                    </div>
                    <div className="flex-1 h-full border-l border-slate-200/60" />
                  </div>
                ))}
              </div>

              <div className="absolute top-0 right-0 left-16 bottom-0 pointer-events-none" style={{ height: `${hours.length * HOUR_HEIGHT}px` }}>
                {positionedLessons.map((lesson) => {
                  const topOffset = ((lesson.startMinutes - START_HOUR * 60) / 60) * HOUR_HEIGHT;
                  const cardHeight = (lesson.durationMinutes / 60) * HOUR_HEIGHT;
                  const widthPercent = 100 / lesson.totalColumns;
                  const leftPercent = lesson.column * widthPercent;
                  const accentColor = lesson.classroom_color || (lesson.format === 'online' ? '#10B981' : '#4F46E5');

                  return (
                    <div
                      key={lesson.id}
                      className="absolute p-1 pointer-events-auto transition-all group z-10"
                      style={{ top: `${Math.max(0, topOffset)}px`, height: `${Math.max(50, cardHeight)}px`, left: `${leftPercent}%`, width: `${widthPercent}%` }}
                    >
                      <div
                        className="h-full w-full rounded-2xl liquid-glass border border-white/80 p-3 flex flex-col justify-between shadow-md relative overflow-hidden transition-all group-hover:scale-[1.01] group-hover:shadow-lg"
                        style={{ borderLeftWidth: '5px', borderLeftColor: accentColor, backgroundColor: 'rgba(255, 255, 255, 0.85)' }}
                      >
                        <div>
                          <div className="flex items-center justify-between gap-2">
                            <span className="font-bold text-xs sm:text-sm text-slate-900 truncate">{lesson.title}</span>
                            {getStatusBadge(lesson.status)}
                          </div>
                          <div className="flex items-center gap-2 text-[11px] text-slate-600 mt-0.5 truncate">
                            <span className="font-medium text-slate-800">{lesson.client_name}</span>
                            <span>•</span>
                            <span className="font-mono">
                              {new Date(lesson.start_time).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })} – {new Date(lesson.end_time).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
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

                        {lesson.status === 'scheduled' && (
                          <div className="flex items-center gap-2 pt-2 border-t border-black/[0.04]">
                            <button onClick={() => handleComplete(lesson.id)} className="px-2.5 py-1 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-white text-[11px] font-semibold flex items-center gap-1 transition-all">
                              <CheckCircle className="w-3 h-3" /> Проведён
                            </button>
                            <button onClick={() => handleNoShow(lesson.id)} className="px-2 py-1 rounded-lg bg-rose-50 hover:bg-rose-100 text-rose-600 text-[11px] font-medium flex items-center gap-1 transition-all">
                              <AlertTriangle className="w-3 h-3" /> Неявка
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

        {viewMode === 'list' && (
          <div className="space-y-3">
            {dayLessons.map((lesson) => (
              <GlassCard key={lesson.id} className="p-5 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
                <div className="flex items-start gap-4">
                  <div
                    className="w-12 h-12 rounded-2xl flex items-center justify-center text-white shrink-0 shadow-sm"
                    style={{ backgroundColor: lesson.classroom_color || (lesson.format === 'online' ? '#10B981' : '#4F46E5') }}
                  >
                    {lesson.format === 'online' ? <Video className="w-6 h-6" /> : <MapPin className="w-6 h-6" />}
                  </div>
                  <div>
                    <div className="flex items-center gap-2 flex-wrap">
                      <h3 className="font-bold text-slate-900 text-base">{lesson.title}</h3>
                      {getStatusBadge(lesson.status)}
                    </div>
                    <div className="flex items-center gap-3 text-xs text-slate-500 mt-1 flex-wrap">
                      <span className="font-semibold text-slate-800">Ученик: {lesson.client_name}</span>
                      <span>•</span>
                      <span className="flex items-center gap-1 font-mono">
                        <Clock className="w-3.5 h-3.5 text-slate-400" />
                        {new Date(lesson.start_time).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })} – {new Date(lesson.end_time).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
                      </span>
                    </div>
                    {lesson.classroom_name && <div className="text-xs text-indigo-600 font-medium mt-1">Аудитория: {lesson.classroom_name}</div>}
                    {lesson.comment && <div className="text-xs text-slate-500 italic mt-1">Примечание: {lesson.comment}</div>}
                  </div>
                </div>
                <div className="flex items-center gap-2 self-end md:self-auto">
                  {lesson.status === 'scheduled' && (
                    <>
                      <GlassButton variant="mint" size="sm" onClick={() => handleComplete(lesson.id)} icon={<CheckCircle className="w-3.5 h-3.5" />}>Проведён</GlassButton>
                      <GlassButton variant="coral" size="sm" onClick={() => handleNoShow(lesson.id)} icon={<AlertTriangle className="w-3.5 h-3.5" />}>Неявка</GlassButton>
                    </>
                  )}
                </div>
              </GlassCard>
            ))}
            {dayLessons.length === 0 && !isLoading && (
              <div className="p-12 text-center liquid-glass rounded-3xl">
                <CalendarIcon className="w-12 h-12 text-slate-300 mx-auto mb-3" />
                <p className="text-base font-medium text-slate-700">На этот день уроков нет</p>
                <p className="text-xs text-slate-400 mt-1">Нажмите «Назначить урок», чтобы запланировать занятие</p>
              </div>
            )}
          </div>
        )}
      </div>

      <GlassModal
        isOpen={isModalOpen}
        onClose={() => { setIsModalOpen(false); setRoomSearch(''); }}
        title="Назначить занятие"
        description="Заполните параметры урока: ученик, время и формат проведения"
      >
        <form onSubmit={handleCreateLesson} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">Ученик</label>
            <select
              value={selectedClientId}
              onChange={(e) => setSelectedClientId(e.target.value)}
              className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 liquid-glass-input focus:border-indigo-600 focus:outline-none"
              required
            >
              <option value="" disabled>Выберите ученика</option>
              {clients.map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
          </div>
          <GlassInput label="Тема / Название урока" placeholder="например: Подготовка к ОГЭ" value={lessonTitle} onChange={(e) => setLessonTitle(e.target.value)} required />
          <div className="grid grid-cols-2 gap-3">
            <GlassInput label="Дата" type="date" value={lessonDate} onChange={(e) => setLessonDate(e.target.value)} required />
            <GlassInput label="Время начала" type="time" value={lessonStartTime} onChange={(e) => setLessonStartTime(e.target.value)} required />
          </div>
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">Длительность</label>
            <div className="grid grid-cols-4 gap-2">
              {['45', '60', '90', '120'].map((mins) => (
                <button
                  key={mins}
                  type="button"
                  onClick={() => setLessonDuration(mins)}
                  className={`py-2 rounded-xl text-xs font-semibold border transition-all ${lessonDuration === mins ? 'bg-indigo-600 text-white border-indigo-600 shadow-md shadow-indigo-600/20' : 'liquid-glass text-slate-700 border-white/60 hover:bg-white/90'}`}
                >
                  {mins} мин
                </button>
              ))}
            </div>
          </div>
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">Формат занятия</label>
            <div className="grid grid-cols-2 gap-3">
              <button
                type="button"
                onClick={() => setLessonFormat('offline')}
                className={`flex items-center justify-center gap-2 py-3 rounded-2xl border text-xs font-bold transition-all ${lessonFormat === 'offline' ? 'bg-indigo-600 text-white border-indigo-600 shadow-md shadow-indigo-600/20' : 'liquid-glass text-slate-700 border-white/60 hover:bg-white/80'}`}
              >
                <MapPin className="w-4 h-4" /> Оффлайн
              </button>
              <button
                type="button"
                onClick={() => setLessonFormat('online')}
                className={`flex items-center justify-center gap-2 py-3 rounded-2xl border text-xs font-bold transition-all ${lessonFormat === 'online' ? 'bg-emerald-600 text-white border-emerald-600 shadow-md shadow-emerald-600/20' : 'liquid-glass text-slate-700 border-white/60 hover:bg-white/80'}`}
              >
                <Video className="w-4 h-4" /> Онлайн
              </button>
            </div>
          </div>

          {lessonFormat === 'offline' && (
            <div className="space-y-2">
              <div className="flex items-center justify-between ml-1">
                <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500">Кабинет школы</label>
                <span className="text-[11px] text-slate-400 font-medium">{classrooms.length} {classrooms.length === 1 ? 'аудитория' : 'доступно'}</span>
              </div>
              {classrooms.length > 4 && (
                <div className="relative">
                  <Search className="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
                  <input type="text" placeholder="Поиск..." value={roomSearch} onChange={(e) => setRoomSearch(e.target.value)} className="w-full pl-8 pr-3 py-1.5 rounded-xl text-xs text-slate-800 liquid-glass-input focus:outline-none placeholder:text-slate-400" />
                </div>
              )}
              {classrooms.length === 0 ? (
                <div className="p-3 text-center text-xs text-slate-400 liquid-glass rounded-2xl">Кабинеты пока не добавлены</div>
              ) : filteredClassrooms.length === 0 ? (
                <div className="p-3 text-center text-xs text-slate-400 liquid-glass rounded-2xl">Кабинет не найден</div>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 max-h-44 overflow-y-auto p-1.5 rounded-2xl bg-white/20 border border-white/50 custom-scrollbar">
                  {filteredClassrooms.map((room) => {
                    const isSelected = selectedClassroomId === room.id;
                    return (
                      <button
                        key={room.id}
                        type="button"
                        onClick={() => setSelectedClassroomId(room.id)}
                        className={`text-left p-2.5 rounded-xl border transition-all ${isSelected ? 'bg-indigo-600 border-indigo-600 text-white shadow-md' : 'bg-white/50 border-white/40 hover:bg-white/80 text-slate-700'}`}
                      >
                        <div className="flex items-center gap-2">
                          <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: room.color }} />
                          <div className="font-bold text-xs truncate">{room.name}</div>
                        </div>
                        <div className={`text-[10px] mt-1 ${isSelected ? 'text-indigo-100' : 'text-slate-500'}`}>До {room.capacity} чел.</div>
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
          )}

          {lessonFormat === 'online' && (
            <GlassInput label="Ссылка на видеоконференцию" placeholder="https://zoom.us/j/..." value={onlineLink} onChange={(e) => setOnlineLink(e.target.value)} icon={<Video className="w-4 h-4" />} required />
          )}

          <GlassInput label="Комментарий / Домашнее задание (необязательно)" placeholder="Что нужно подготовить к уроку..." value={lessonComment} onChange={(e) => setLessonComment(e.target.value)} />

          <div className="pt-2 flex justify-end gap-3">
            <GlassButton type="button" variant="secondary" onClick={() => setIsModalOpen(false)}>Отмена</GlassButton>
            <GlassButton type="submit" variant="primary" isLoading={isSubmitting}>Запланировать</GlassButton>
          </div>
        </form>
      </GlassModal>
    </div>
  );
};
