import React, { useState, useEffect, useMemo } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassButton } from '../../../shared/components/GlassButton';
import { Classroom, Client, SubscriptionFormat } from '../../../types/schedule';
import { createLesson, createSeries } from '../../../api/schedule';
import { SCHEDULE_TIME_OPTIONS } from '../types';
import { CreateLessonForm } from './CreateLessonForm';
import { getRRuleDayFromDate } from '../utils/recurrence';

interface CreateLessonModalProps {
  isOpen: boolean;
  onClose: () => void;
  clients: Client[];
  classrooms: Classroom[];
  initialClientId?: string;
  initialDate?: string;
  initialStartTime?: string;
  initialDuration?: string;
  initialTitle?: string;
  onCreated: () => Promise<void>;
}

export const CreateLessonModal: React.FC<CreateLessonModalProps> = ({
  isOpen,
  onClose,
  clients,
  classrooms,
  initialClientId = '',
  initialDate = new Date().toISOString().split('T')[0],
  initialStartTime = '14:00',
  initialDuration = '60',
  initialTitle = '',
  onCreated,
}) => {
  const [selectedClientId, setSelectedClientId] = useState(initialClientId);
  const [lessonTitle, setLessonTitle] = useState(initialTitle);
  const [lessonDate, setLessonDate] = useState(initialDate);
  const [lessonStartTime, setLessonStartTime] = useState(initialStartTime);
  const [lessonDuration, setLessonDuration] = useState(initialDuration);
  const [selectedFormat, setSelectedFormat] = useState<SubscriptionFormat>('individual');
  const [locationType, setLocationType] = useState<'offline' | 'online'>('offline');
  const [selectedClassroomId, setSelectedClassroomId] = useState('');
  const [onlineLink, setOnlineLink] = useState('https://telemost.yandex.ru/j/school-lesson');
  const [lessonComment, setLessonComment] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Recurrence states
  const [isRecurring, setIsRecurring] = useState(false);
  const [selectedDays, setSelectedDays] = useState<string[]>(['MO']);
  const [untilMode, setUntilMode] = useState<'never' | 'date'>('never');
  const [untilDate, setUntilDate] = useState('');

  useEffect(() => {
    if (isOpen) {
      setSelectedClientId(initialClientId);
      setLessonTitle(initialTitle);
      setLessonDate(initialDate);
      setLessonStartTime(initialStartTime);
      setLessonDuration(initialDuration);
      setSelectedFormat('individual');
      setLocationType('offline');
      setSelectedClassroomId('');
      setLessonComment('');
      setIsRecurring(false);
      setSelectedDays([getRRuleDayFromDate(initialDate)]);
      setUntilMode('never');
      setUntilDate('');
    }
  }, [isOpen, initialClientId, initialTitle, initialDate, initialStartTime, initialDuration]);

  // Sync selectedDays when lessonDate changes if user hasn't explicitly customized yet
  const handleDateChange = (newDate: string) => {
    setLessonDate(newDate);
    const day = getRRuleDayFromDate(newDate);
    if (!selectedDays.includes(day)) {
      setSelectedDays([day]);
    }
  };

  const handleToggleDay = (day: string) => {
    setSelectedDays((prev) =>
      prev.includes(day) ? prev.filter((d) => d !== day) : [...prev, day]
    );
  };

  const allStartTimeOptions = useMemo(() => {
    if (lessonStartTime && !SCHEDULE_TIME_OPTIONS.includes(lessonStartTime)) {
      return [...SCHEDULE_TIME_OPTIONS, lessonStartTime].sort();
    }
    return SCHEDULE_TIME_OPTIONS;
  }, [lessonStartTime]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedClientId || !lessonTitle.trim()) return;

    if (isRecurring && selectedDays.length === 0) {
      alert('Пожалуйста, выберите хотя бы один день недели для повторения');
      return;
    }

    setIsSubmitting(true);
    try {
      if (isRecurring) {
        // Create repeating series
        const rrule = `FREQ=WEEKLY;BYDAY=${selectedDays.join(',')}`;
        await createSeries({
          client_id: selectedClientId,
          classroom_id: locationType === 'offline' ? selectedClassroomId || null : null,
          title: lessonTitle.trim(),
          rrule,
          start_time_of_day: lessonStartTime,
          duration_minutes: Number(lessonDuration),
          format: selectedFormat,
          location_or_url: locationType === 'online' ? (onlineLink.trim() || 'online') : null,
          notes: lessonComment.trim() || null,
          start_date: lessonDate,
          until_date: untilMode === 'date' && untilDate ? untilDate : null,
        });
      } else {
        // Single lesson
        const [hoursVal, minsVal] = lessonStartTime.split(':').map(Number);
        const start = new Date(lessonDate);
        start.setHours(hoursVal, minsVal, 0, 0);
        const end = new Date(start.getTime() + Number(lessonDuration) * 60 * 1000);

        await createLesson({
          client_id: selectedClientId,
          title: lessonTitle.trim(),
          format: selectedFormat,
          classroom_id: locationType === 'offline' ? selectedClassroomId || null : null,
          online_link: locationType === 'online' ? (onlineLink.trim() || 'online') : undefined,
          comment: lessonComment.trim() || undefined,
          start_time: start.toISOString(),
          end_time: end.toISOString(),
        });
      }

      await onCreated();
      onClose();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка создания занятия');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title={isRecurring ? 'Назначить регулярную серию' : 'Назначить занятие'}
      description="Выберите ученика, дату, формат и параметры проведения"
      maxWidth="lg"
    >
      <form onSubmit={handleCreate} className="space-y-4">
        <CreateLessonForm
          clients={clients}
          classrooms={classrooms}
          selectedClientId={selectedClientId}
          onSelectClient={(id) => {
            setSelectedClientId(id);
            const found = clients.find((c) => c.id === id);
            if (found && (!lessonTitle || lessonTitle.startsWith('Урок'))) {
              setLessonTitle(`Урок: ${found.name}`);
            }
          }}
          lessonTitle={lessonTitle}
          onChangeTitle={setLessonTitle}
          lessonDate={lessonDate}
          onChangeDate={handleDateChange}
          lessonStartTime={lessonStartTime}
          onChangeStartTime={setLessonStartTime}
          lessonDuration={lessonDuration}
          onChangeDuration={setLessonDuration}
          allStartTimeOptions={allStartTimeOptions}
          selectedFormat={selectedFormat}
          onChangeFormat={setSelectedFormat}
          locationType={locationType}
          onChangeLocationType={setLocationType}
          selectedClassroomId={selectedClassroomId}
          onChangeClassroomId={setSelectedClassroomId}
          onlineLink={onlineLink}
          onChangeOnlineLink={setOnlineLink}
          lessonComment={lessonComment}
          onChangeComment={setLessonComment}
          isRecurring={isRecurring}
          onToggleRecurring={setIsRecurring}
          selectedDays={selectedDays}
          onToggleDay={handleToggleDay}
          untilMode={untilMode}
          onChangeUntilMode={setUntilMode}
          untilDate={untilDate}
          onChangeUntilDate={setUntilDate}
        />

        <div className="pt-3 flex justify-end gap-3 border-t border-black/[0.05]">
          <GlassButton
            type="button"
            variant="secondary"
            onClick={onClose}
          >
            Отмена
          </GlassButton>
          <GlassButton
            type="submit"
            variant="primary"
            isLoading={isSubmitting}
          >
            {isRecurring ? 'Создать серию уроков' : 'Запланировать урок'}
          </GlassButton>
        </div>
      </form>
    </GlassModal>
  );
};
