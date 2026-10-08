import { useState, useEffect, useRef, useCallback } from 'react';

export interface JournalDraft {
  topic: string;
  notes: string;
  performanceScore: number | null;
  homeworkTitle: string;
  homeworkDescription: string;
  homeworkDueDate: string;
}

const STORAGE_KEY_PREFIX = 'school_draft_journal_';

/**
 * Хук автосохранения полей отчета по уроку в localStorage.
 * Предоставляет стабильные методы сохранения (с дебаунсом 300мс),
 * чтения и очистки черновика, не вызывая паразитных ре-рендеров.
 */
export function useJournalDraft(lessonId: string | null | undefined) {
  const [hasDraft, setHasDraft] = useState<boolean>(false);
  const debounceTimerRef = useRef<NodeJS.Timeout | null>(null);

  const getStorageKey = useCallback((id: string) => `${STORAGE_KEY_PREFIX}${id}`, []);

  // Синхронизация флага hasDraft при смене lessonId
  useEffect(() => {
    if (!lessonId) {
      setHasDraft(false);
      return;
    }
    try {
      const saved = localStorage.getItem(getStorageKey(lessonId));
      if (saved) {
        const parsed = JSON.parse(saved) as JournalDraft;
        const hasContent = Boolean(
          parsed.topic?.trim() ||
          parsed.notes?.trim() ||
          parsed.performanceScore !== null ||
          parsed.homeworkTitle?.trim() ||
          parsed.homeworkDescription?.trim() ||
          parsed.homeworkDueDate?.trim()
        );
        setHasDraft(hasContent);
        return;
      }
    } catch {
      // ignore
    }
    setHasDraft(false);
  }, [lessonId, getStorageKey]);

  // Чтение черновика по требованию (например, при инициализации формы)
  const getDraft = useCallback((): JournalDraft | null => {
    if (!lessonId) return null;
    try {
      const saved = localStorage.getItem(getStorageKey(lessonId));
      if (saved) {
        const parsed = JSON.parse(saved) as JournalDraft;
        const hasContent = Boolean(
          parsed.topic?.trim() ||
          parsed.notes?.trim() ||
          parsed.performanceScore !== null ||
          parsed.homeworkTitle?.trim() ||
          parsed.homeworkDescription?.trim() ||
          parsed.homeworkDueDate?.trim()
        );
        if (hasContent) {
          return parsed;
        }
      }
    } catch {
      // ignore
    }
    return null;
  }, [lessonId, getStorageKey]);

  // Сохранить или обновить черновик с дебаунсом 300мс
  const setDraft = useCallback((nextDraft: JournalDraft) => {
    if (!lessonId) return;

    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }

    debounceTimerRef.current = setTimeout(() => {
      try {
        const hasContent = Boolean(
          nextDraft.topic?.trim() ||
          nextDraft.notes?.trim() ||
          nextDraft.performanceScore !== null ||
          nextDraft.homeworkTitle?.trim() ||
          nextDraft.homeworkDescription?.trim() ||
          nextDraft.homeworkDueDate?.trim()
        );

        if (hasContent) {
          localStorage.setItem(getStorageKey(lessonId), JSON.stringify(nextDraft));
          setHasDraft(true);
        } else {
          localStorage.removeItem(getStorageKey(lessonId));
          setHasDraft(false);
        }
      } catch {
        // Игнорируем ошибки квоты localStorage
      }
    }, 300);
  }, [lessonId, getStorageKey]);

  // Немедленная очистка черновика (например, при успешном сохранении отчета)
  const clearDraft = useCallback(() => {
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }
    if (lessonId) {
      try {
        localStorage.removeItem(getStorageKey(lessonId));
      } catch {
        // ignore
      }
    }
    setHasDraft(false);
  }, [lessonId, getStorageKey]);

  useEffect(() => {
    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
    };
  }, []);

  return {
    getDraft,
    setDraft,
    clearDraft,
    hasDraft,
  };
}
