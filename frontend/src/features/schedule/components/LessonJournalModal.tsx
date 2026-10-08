import React, { useState, useEffect, useCallback } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassButton } from '../../../shared/components/GlassButton';
import { GlassInput } from '../../../shared/components/GlassInput';
import { Lesson } from '../../../types/schedule';
import { HomeworkAssignment, HomeworkStatus } from '../../../types/journal';
import {
  getLessonJournalBundle,
  upsertLessonJournal,
  createHomework,
  updateHomeworkStatus,
} from '../../../api/journal';
import { LessonScoreChips } from './journal/LessonScoreChips';
import { LessonDueHomeworks } from './journal/LessonDueHomeworks';
import { LessonAssignHomework } from './journal/LessonAssignHomework';
import { BookOpen, Calendar, Clock, Loader2, AlertCircle, Save } from 'lucide-react';

interface LessonJournalModalProps {
  isOpen: boolean;
  onClose: () => void;
  lesson: Lesson | null;
  clientDisplayName: string;
  onSaved?: () => void;
}

export const LessonJournalModal: React.FC<LessonJournalModalProps> = ({
  isOpen,
  onClose,
  lesson,
  clientDisplayName,
  onSaved,
}) => {
  const [topic, setTopic] = useState('');
  const [notes, setNotes] = useState('');
  const [performanceScore, setPerformanceScore] = useState<number | null>(null);
  const [assignedHomeworks, setAssignedHomeworks] = useState<HomeworkAssignment[]>([]);
  const [dueHomeworks, setDueHomeworks] = useState<HomeworkAssignment[]>([]);

  // Новое ДЗ
  const [newHwTitle, setNewHwTitle] = useState('');
  const [newHwDescription, setNewHwDescription] = useState('');
  const [newHwDueDate, setNewHwDueDate] = useState('');

  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [updatingDueId, setUpdatingDueId] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const loadJournal = useCallback(async (lessonId: string, currentTitle: string) => {
    setIsLoading(true);
    setErrorMessage(null);
    try {
      const bundle = await getLessonJournalBundle(lessonId);
      if (bundle.journal) {
        setTopic(bundle.journal.topic || '');
        setNotes(bundle.journal.notes || '');
        setPerformanceScore(bundle.journal.performance_score ?? null);
      } else {
        setTopic(currentTitle && currentTitle !== 'Урок' ? currentTitle : '');
        setNotes('');
        setPerformanceScore(null);
      }
      setAssignedHomeworks(bundle.assigned_homeworks || []);
      setDueHomeworks(bundle.due_homeworks || []);
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Ошибка загрузки дневника');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (isOpen && lesson) {
      setNewHwTitle('');
      setNewHwDescription('');
      setNewHwDueDate('');
      loadJournal(lesson.id, lesson.title);
    }
  }, [isOpen, lesson, loadJournal]);

  const handleUpdateDueStatus = async (homeworkId: string, status: HomeworkStatus) => {
    setUpdatingDueId(homeworkId);
    try {
      const updated = await updateHomeworkStatus(homeworkId, { status });
      setDueHomeworks((prev) =>
        prev.map((item) => (item.id === homeworkId ? updated : item))
      );
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка смены статуса ДЗ');
    } finally {
      setUpdatingDueId(null);
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!lesson) return;

    if (!topic.trim()) {
      setErrorMessage('Пожалуйста, укажите тему проведенного занятия');
      return;
    }

    setIsSaving(true);
    setErrorMessage(null);

    try {
      await upsertLessonJournal(lesson.id, {
        topic: topic.trim(),
        notes: notes.trim() || null,
        performance_score: performanceScore,
      });

      // Если заполнено новое ДЗ, сохраняем его
      const clientId = lesson.client_id || lesson.student_id;
      if (newHwTitle.trim() && clientId) {
        await createHomework(clientId, {
          title: newHwTitle.trim(),
          description: newHwDescription.trim() || null,
          due_date: newHwDueDate || null,
          assigned_lesson_id: lesson.id,
        });
      }

      onSaved?.();
      onClose();
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Ошибка сохранения отчета');
    } finally {
      setIsSaving(false);
    }
  };

  if (!lesson) return null;

  const lessonDateStr = new Date(lesson.start_time).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'long',
    weekday: 'short',
  });
  const lessonTimeStr = `${new Date(lesson.start_time).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  })} – ${new Date(lesson.end_time).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  })}`;

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title="Дневник занятия"
      description={`${clientDisplayName} • ${lessonDateStr}`}
      maxWidth="lg"
    >
      {isLoading ? (
        <div className="py-16 flex flex-col items-center justify-center space-y-3">
          <Loader2 className="w-8 h-8 text-indigo-600 animate-spin" />
          <p className="text-sm font-medium text-slate-500">Загрузка данных дневника...</p>
        </div>
      ) : (
        <form onSubmit={handleSave} className="space-y-4">
          {/* Информационная шапка занятия */}
          <div className="flex items-center gap-3 p-3 rounded-2xl bg-white/35 backdrop-blur-md border border-white/60 text-xs text-slate-600 flex-wrap">
            <div className="flex items-center gap-1.5 font-semibold text-slate-800">
              <BookOpen className="w-4 h-4 text-indigo-600" />
              <span>{clientDisplayName}</span>
            </div>
            <span className="text-slate-300">•</span>
            <div className="flex items-center gap-1 font-mono text-slate-600">
              <Clock className="w-3.5 h-3.5 text-slate-400" />
              <span>{lessonTimeStr}</span>
            </div>
            <span className="text-slate-300">•</span>
            <div className="flex items-center gap-1 text-slate-600">
              <Calendar className="w-3.5 h-3.5 text-slate-400" />
              <span>{lessonDateStr}</span>
            </div>
          </div>

          {errorMessage && (
            <div className="p-3 rounded-2xl bg-rose-500/15 border border-rose-400/30 backdrop-blur-md flex items-center gap-2.5 text-rose-800 text-xs font-medium animate-in fade-in">
              <AlertCircle className="w-4 h-4 text-rose-600 shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          {/* Экспресс-проверка ДЗ, выданного к этому уроку */}
          <LessonDueHomeworks
            homeworks={dueHomeworks}
            onUpdateStatus={handleUpdateDueStatus}
            isUpdatingId={updatingDueId}
          />

          {/* Быстрый ввод темы занятия */}
          <GlassInput
            label="Тема занятия *"
            placeholder="Например: Квадратные уравнения и теорема Виета"
            value={topic}
            onChange={(e) => setTopic(e.target.value)}
            disabled={isSaving}
            autoFocus
          />

          {/* Легкая шкала оценки понимания (1..5) */}
          <LessonScoreChips
            score={performanceScore}
            onChange={setPerformanceScore}
            disabled={isSaving}
          />

          {/* Заметки преподавателя */}
          <div className="space-y-1.5">
            <label className="block text-xs font-semibold text-slate-700 ml-1">
              Заметки преподавателя
            </label>
            <textarea
              disabled={isSaving}
              rows={3}
              placeholder="Пробелы, успехи, особенности понимания, рекомендации к следующему уроку..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              className="w-full rounded-2xl p-3 text-xs sm:text-sm text-slate-800 placeholder-slate-400 liquid-glass-input resize-none"
            />
          </div>

          {/* Блок назначения ДЗ */}
          <LessonAssignHomework
            assignedHomeworks={assignedHomeworks}
            title={newHwTitle}
            onTitleChange={setNewHwTitle}
            description={newHwDescription}
            onDescriptionChange={setNewHwDescription}
            dueDate={newHwDueDate}
            onDueDateChange={setNewHwDueDate}
            disabled={isSaving}
          />

          {/* Кнопки действий */}
          <div className="flex items-center justify-end gap-2.5 pt-2">
            <GlassButton
              type="button"
              variant="secondary"
              onClick={onClose}
              disabled={isSaving}
            >
              Отмена
            </GlassButton>
            <GlassButton
              type="submit"
              variant="primary"
              disabled={isSaving}
              icon={isSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
            >
              {isSaving ? 'Сохранение...' : 'Сохранить отчет'}
            </GlassButton>
          </div>
        </form>
      )}
    </GlassModal>
  );
};
