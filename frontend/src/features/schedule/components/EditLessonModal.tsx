import React, { useState, useEffect, useMemo } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassButton } from '../../../shared/components/GlassButton';
import { Classroom, Lesson, SubscriptionFormat, RecurrenceScope } from '../../../types/schedule';
import { updateLesson, cancelLesson, markNoShow } from '../../../api/schedule';
import { SCHEDULE_TIME_OPTIONS } from '../types';
import { EditLessonFormFields } from './EditLessonFormFields';
import { RecurrenceScopeModal } from './RecurrenceScopeModal';

interface EditLessonModalProps {
  isOpen: boolean;
  onClose: () => void;
  lesson: Lesson | null;
  classrooms: Classroom[];
  clientDisplayName: string;
  onUpdated: () => Promise<void>;
}

export const EditLessonModal: React.FC<EditLessonModalProps> = ({
  isOpen,
  onClose,
  lesson,
  classrooms,
  clientDisplayName,
  onUpdated,
}) => {
  const [editTitle, setEditTitle] = useState('');
  const [editDate, setEditDate] = useState('');
  const [editStartTime, setEditStartTime] = useState('');
  const [editEndTime, setEditEndTime] = useState('');
  const [editFormat, setEditFormat] = useState<SubscriptionFormat>('individual');
  const [editLocationType, setEditLocationType] = useState<'offline' | 'online'>('offline');
  const [editClassroomId, setEditClassroomId] = useState('');
  const [editOnlineLink, setEditOnlineLink] = useState('');
  const [editComment, setEditComment] = useState('');
  const [cancelReason, setCancelReason] = useState('');
  const [isCancelling, setIsCancelling] = useState(false);
  const [isSubmittingEdit, setIsSubmittingEdit] = useState(false);

  // Recurrence scope modal state
  const [scopeModalAction, setScopeModalAction] = useState<'edit' | 'cancel' | null>(null);

  useEffect(() => {
    if (lesson && isOpen) {
      setEditTitle(lesson.title);
      const start = new Date(lesson.start_time);
      const end = new Date(lesson.end_time);
      const y = start.getFullYear();
      const m = String(start.getMonth() + 1).padStart(2, '0');
      const d = String(start.getDate()).padStart(2, '0');
      setEditDate(`${y}-${m}-${d}`);
      setEditStartTime(
        `${String(start.getHours()).padStart(2, '0')}:${String(start.getMinutes()).padStart(2, '0')}`
      );
      setEditEndTime(
        `${String(end.getHours()).padStart(2, '0')}:${String(end.getMinutes()).padStart(2, '0')}`
      );
      const fmt: SubscriptionFormat =
        lesson.format === 'pair' || lesson.format === 'group' ? lesson.format : 'individual';
      setEditFormat(fmt);

      const isOnline =
        Boolean(lesson.location_or_url && (lesson.location_or_url.startsWith('http') || lesson.location_or_url === 'online')) ||
        Boolean(lesson.online_link && lesson.online_link !== 'offline') ||
        (lesson.format as string) === 'online';

      setEditLocationType(isOnline ? 'online' : 'offline');
      setEditClassroomId(lesson.classroom_id || '');
      const rawLink = lesson.online_link || lesson.location_or_url || '';
      setEditOnlineLink(rawLink.startsWith('http') ? rawLink : '');
      setEditComment(lesson.comment || '');
      setCancelReason('');
      setIsCancelling(false);
      setScopeModalAction(null);
    }
  }, [lesson, isOpen]);

  const allStartTimeOptions = useMemo(() => {
    if (editStartTime && !SCHEDULE_TIME_OPTIONS.includes(editStartTime)) {
      return [...SCHEDULE_TIME_OPTIONS, editStartTime].sort();
    }
    return SCHEDULE_TIME_OPTIONS;
  }, [editStartTime]);

  const allEndTimeOptions = useMemo(() => {
    if (editEndTime && !SCHEDULE_TIME_OPTIONS.includes(editEndTime)) {
      return [...SCHEDULE_TIME_OPTIONS, editEndTime].sort();
    }
    return SCHEDULE_TIME_OPTIONS;
  }, [editEndTime]);

  const executeSave = async (scope?: RecurrenceScope) => {
    if (!lesson) return;
    setIsSubmittingEdit(true);
    try {
      const [startH, startM] = editStartTime.split(':').map(Number);
      const [endH, endM] = editEndTime.split(':').map(Number);
      const start = new Date(editDate);
      start.setHours(startH, startM, 0, 0);
      const end = new Date(editDate);
      end.setHours(endH, endM, 0, 0);

      const isOnline = editLocationType === 'online';

      await updateLesson(lesson.id, {
        title: editTitle.trim(),
        start_time: start.toISOString(),
        end_time: end.toISOString(),
        format: editFormat,
        classroom_id: isOnline ? null : editClassroomId || null,
        online_link: isOnline ? (editOnlineLink.trim() || 'online') : '',
        comment: editComment.trim(),
        scope,
      });

      await onUpdated();
      onClose();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка сохранения занятия');
    } finally {
      setIsSubmittingEdit(false);
    }
  };

  const handleSaveEdit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!lesson) return;

    if (lesson.is_recurring || lesson.series_id) {
      setScopeModalAction('edit');
    } else {
      await executeSave();
    }
  };

  const executeCancel = async (scope?: RecurrenceScope) => {
    if (!lesson) return;
    try {
      await cancelLesson(lesson.id, cancelReason.trim() || undefined, scope);
      await onUpdated();
      onClose();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка отмены занятия');
    }
  };

  const handleCancelLesson = async () => {
    if (!lesson) return;

    if (lesson.is_recurring || lesson.series_id) {
      setScopeModalAction('cancel');
    } else {
      await executeCancel();
    }
  };

  const handleNoShow = async () => {
    if (!lesson) return;
    if (!window.confirm('Отметить неявку ученика на занятие?')) return;
    try {
      await markNoShow(lesson.id);
      await onUpdated();
      onClose();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка отметки неявки');
    }
  };

  if (!lesson) return null;

  const isCancelled =
    lesson.status === 'cancelled' ||
    lesson.status.startsWith('cancelled') ||
    lesson.status === 'declined';

  return (
    <>
      <GlassModal
        isOpen={isOpen}
        onClose={onClose}
        title={`Урок: ${clientDisplayName}`}
        description={isCancelled ? 'Занятие отменено' : 'Измените время, аудиторию или отмените занятие.'}
        maxWidth="lg"
      >
        <form onSubmit={handleSaveEdit} className="space-y-4">
          {isCancelled && (
            <div className="p-3 rounded-2xl bg-rose-500/10 border border-rose-400/40 text-rose-700 text-xs">
              <strong>Занятие отменено</strong>{lesson.cancel_reason ? `: ${lesson.cancel_reason}` : ''}
            </div>
          )}

          <EditLessonFormFields
            editTitle={editTitle}
            onChangeTitle={setEditTitle}
            editDate={editDate}
            onChangeDate={setEditDate}
            editStartTime={editStartTime}
            onChangeStartTime={setEditStartTime}
            editEndTime={editEndTime}
            onChangeEndTime={setEditEndTime}
            allStartTimeOptions={allStartTimeOptions}
            allEndTimeOptions={allEndTimeOptions}
            editFormat={editFormat}
            onChangeFormat={setEditFormat}
            editLocationType={editLocationType}
            onChangeLocationType={setEditLocationType}
            editClassroomId={editClassroomId}
            onChangeClassroomId={setEditClassroomId}
            editOnlineLink={editOnlineLink}
            onChangeOnlineLink={setEditOnlineLink}
            editComment={editComment}
            onChangeComment={setEditComment}
            classrooms={classrooms}
            isCancelled={isCancelled}
            lessonStatus={lesson.status}
            isCancelling={isCancelling}
            onSetIsCancelling={setIsCancelling}
            cancelReason={cancelReason}
            onChangeCancelReason={setCancelReason}
            onCancelClick={handleCancelLesson}
            onNoShowClick={handleNoShow}
          />

          <div className="pt-3 flex justify-end gap-3 border-t border-black/[0.05]">
            <GlassButton
              type="button"
              variant="secondary"
              onClick={onClose}
            >
              Закрыть
            </GlassButton>
            {!isCancelled && (
              <GlassButton
                type="submit"
                variant="primary"
                isLoading={isSubmittingEdit}
              >
                Сохранить
              </GlassButton>
            )}
          </div>
        </form>
      </GlassModal>

      {/* Recurrence Scope Confirmation Modal (Google Calendar Pattern) */}
      {scopeModalAction && (
        <RecurrenceScopeModal
          isOpen={Boolean(scopeModalAction)}
          onClose={() => setScopeModalAction(null)}
          actionType={scopeModalAction}
          lessonTitle={lesson.title}
          onConfirm={(scope) => {
            if (scopeModalAction === 'edit') {
              executeSave(scope);
            } else if (scopeModalAction === 'cancel') {
              executeCancel(scope);
            }
          }}
        />
      )}
    </>
  );
};
