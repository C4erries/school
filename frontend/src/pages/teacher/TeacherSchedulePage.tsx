import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { GlassButton } from '../../shared/components/GlassButton';
import { Classroom, Client, Lesson } from '../../types/schedule';
import { getClassrooms, getClients, getLessons, completeLesson } from '../../api/schedule';
import { useSchedulePositioning } from '../../features/schedule/hooks/useSchedulePositioning';
import { ScheduleToolbar } from '../../features/schedule/components/ScheduleToolbar';
import { ScheduleWeekView } from '../../features/schedule/components/ScheduleWeekView';
import { ScheduleDayView } from '../../features/schedule/components/ScheduleDayView';
import { ScheduleListView } from '../../features/schedule/components/ScheduleListView';
import { CreateLessonModal } from '../../features/schedule/components/CreateLessonModal';
import { EditLessonModal } from '../../features/schedule/components/EditLessonModal';
import { CalendarSyncModal } from '../../features/calendar/components/CalendarSyncModal';
import { Plus, RefreshCw, AlertCircle, Calendar as CalendarIcon } from 'lucide-react';

export const TeacherSchedulePage: React.FC = () => {
  const [clients, setClients] = useState<Client[]>([]);
  const [classrooms, setClassrooms] = useState<Classroom[]>([]);
  const [lessons, setLessons] = useState<Lesson[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'day' | 'week' | 'list'>('week');
  const [currentDate, setCurrentDate] = useState<Date>(new Date());

  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [createInitial, setCreateInitial] = useState<{
    clientId?: string;
    title?: string;
    date?: string;
    startTime?: string;
    duration?: string;
  }>({});

  const [editingLesson, setEditingLesson] = useState<Lesson | null>(null);
  const [isSyncModalOpen, setIsSyncModalOpen] = useState(false);
  const { hours, positionLessons } = useSchedulePositioning();

  const loadData = useCallback(async () => {
    setIsLoading(true);
    setErrorMessage(null);
    try {
      const [cls, rms, les] = await Promise.all([getClients(), getClassrooms(), getLessons()]);
      setClients(cls);
      setClassrooms(rms);
      setLessons(les);
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Ошибка загрузки расписания');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const weekDays = useMemo(() => {
    const d = new Date(currentDate);
    const dayOfWeek = d.getDay();
    const diff = d.getDate() - dayOfWeek + (dayOfWeek === 0 ? -6 : 1);
    const monday = new Date(d.setDate(diff));
    monday.setHours(0, 0, 0, 0);
    return Array.from({ length: 7 }, (_, i) => {
      const day = new Date(monday);
      day.setDate(monday.getDate() + i);
      return day;
    });
  }, [currentDate]);

  const dayLessons = useMemo(() => {
    const targetStr = currentDate.toISOString().split('T')[0];
    return lessons.filter((l) => l.start_time.startsWith(targetStr));
  }, [lessons, currentDate]);

  const positionedDayLessons = useMemo(() => positionLessons(dayLessons), [dayLessons, positionLessons]);

  const handleNavigate = (direction: -1 | 1) => {
    const next = new Date(currentDate);
    next.setDate(next.getDate() + direction * (viewMode === 'week' ? 7 : 1));
    setCurrentDate(next);
  };

  const handleSlotClick = (date: Date, hour: number, minute: number = 0) => {
    const y = date.getFullYear();
    const m = String(date.getMonth() + 1).padStart(2, '0');
    const d = String(date.getDate()).padStart(2, '0');
    setCreateInitial({
      date: `${y}-${m}-${d}`,
      startTime: `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`,
      duration: '60',
      title: 'Урок',
      clientId: '',
    });
    setIsCreateOpen(true);
  };

  const handleSlotDragSelect = (
    date: Date,
    startHour: number,
    startMinute: number,
    durationMinutes: number
  ) => {
    const y = date.getFullYear();
    const m = String(date.getMonth() + 1).padStart(2, '0');
    const d = String(date.getDate()).padStart(2, '0');
    setCreateInitial({
      date: `${y}-${m}-${d}`,
      startTime: `${String(startHour).padStart(2, '0')}:${String(startMinute).padStart(2, '0')}`,
      duration: String(durationMinutes),
      title: 'Урок',
      clientId: '',
    });
    setIsCreateOpen(true);
  };

  const handleQuickComplete = async (e: React.MouseEvent, lessonId: string) => {
    e.stopPropagation();
    try {
      await completeLesson(lessonId);
      await loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка отметки проведения');
    }
  };

  const getClientDisplayName = useCallback((lesson: Lesson): string => {
    if (lesson.client_name) return lesson.client_name;
    if (lesson.student_name) return lesson.student_name;
    const found = clients.find((c) => c.id === lesson.client_id || c.id === lesson.student_id);
    return found ? found.name : 'Ученик';
  }, [clients]);

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">Расписание уроков</h1>
          <p className="text-sm text-slate-500 mt-1">Календарь занятий, нахлёсты, быстрое создание в 1 клик</p>
        </div>
        <div className="flex items-center gap-2.5 self-stretch sm:self-auto">
          <GlassButton
            variant="secondary"
            onClick={() => setIsSyncModalOpen(true)}
            icon={<CalendarIcon className="w-4 h-4 text-indigo-600" />}
          >
            Синхронизация
          </GlassButton>
          <GlassButton
            variant="primary"
            onClick={() => {
              setCreateInitial({
                date: new Date().toISOString().split('T')[0],
                startTime: '14:00',
                title: '',
                clientId: '',
              });
              setIsCreateOpen(true);
            }}
            icon={<Plus className="w-4 h-4" />}
          >
            Назначить урок
          </GlassButton>
        </div>
      </div>

      {errorMessage && (
        <div className="p-4 rounded-3xl bg-rose-500/10 border border-rose-500/20 backdrop-blur-md flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
            <p className="text-sm font-medium text-rose-900">{errorMessage}</p>
          </div>
          <GlassButton variant="secondary" size="sm" onClick={loadData} icon={<RefreshCw className="w-3.5 h-3.5" />}>
            Повторить
          </GlassButton>
        </div>
      )}

      <ScheduleToolbar
        currentDate={currentDate}
        viewMode={viewMode}
        weekDays={weekDays}
        clients={clients}
        onNavigate={handleNavigate}
        onGoToday={() => setCurrentDate(new Date())}
        onViewModeChange={setViewMode}
        onQuickAssignClient={(c) => {
          setCreateInitial({ clientId: c.id, title: `Урок: ${c.name}` });
          setIsCreateOpen(true);
        }}
      />

      {viewMode === 'week' && (
        <ScheduleWeekView
          currentDate={currentDate}
          setCurrentDate={setCurrentDate}
          weekDays={weekDays}
          hours={hours}
          lessons={lessons}
          positionLessons={positionLessons}
          onSlotClick={handleSlotClick}
          onSlotDragSelect={handleSlotDragSelect}
          onQuickComplete={handleQuickComplete}
          onOpenEdit={setEditingLesson}
          getClientDisplayName={getClientDisplayName}
        />
      )}

      {viewMode === 'day' && (
        <ScheduleDayView
          currentDate={currentDate}
          hours={hours}
          positionedDayLessons={positionedDayLessons}
          onSlotClick={handleSlotClick}
          onSlotDragSelect={handleSlotDragSelect}
          onQuickComplete={handleQuickComplete}
          onOpenEdit={setEditingLesson}
          getClientDisplayName={getClientDisplayName}
        />
      )}

      {viewMode === 'list' && (
        <ScheduleListView
          dayLessons={dayLessons}
          isLoading={isLoading}
          onQuickComplete={handleQuickComplete}
          onOpenEdit={setEditingLesson}
          getClientDisplayName={getClientDisplayName}
        />
      )}

      <CreateLessonModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        clients={clients}
        classrooms={classrooms}
        initialClientId={createInitial.clientId}
        initialTitle={createInitial.title}
        initialDate={createInitial.date}
        initialStartTime={createInitial.startTime}
        initialDuration={createInitial.duration || '60'}
        onCreated={loadData}
      />

      <EditLessonModal
        isOpen={Boolean(editingLesson)}
        onClose={() => setEditingLesson(null)}
        lesson={editingLesson}
        classrooms={classrooms}
        clientDisplayName={editingLesson ? getClientDisplayName(editingLesson) : ''}
        onUpdated={loadData}
      />

      <CalendarSyncModal
        isOpen={isSyncModalOpen}
        onClose={() => setIsSyncModalOpen(false)}
        onImportSuccess={loadData}
      />
    </div>
  );
};

export default TeacherSchedulePage;
