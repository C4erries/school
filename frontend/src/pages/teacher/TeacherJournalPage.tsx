import React, { useState, useEffect, useCallback } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { Client, Lesson } from '../../types/schedule';
import { StudyStream, StudyStreamItem } from '../../types/journal';
import { getClients, getLessons } from '../../api/schedule';
import {
  getClientStudyStream,
  createClientNote,
  deleteClientNote,
} from '../../api/journal';
import { JournalSidebar } from '../../features/journal/JournalSidebar';
import { JournalChatHeader } from '../../features/journal/JournalChatHeader';
import { UpcomingLessonBanner } from '../../features/journal/UpcomingLessonBanner';
import { StudyStreamFeed } from '../../features/journal/StudyStreamFeed';
import { StudyStreamInputBar } from '../../features/journal/StudyStreamInputBar';
import { LessonJournalModal } from '../../features/schedule/components/LessonJournalModal';
import { BookOpen, AlertCircle, RefreshCw } from 'lucide-react';

export const TeacherJournalPage: React.FC = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();

  const queryClientId = searchParams.get('clientId');
  const queryLessonId = searchParams.get('lessonId');

  const [clients, setClients] = useState<Client[]>([]);
  const [selectedClientId, setSelectedClientId] = useState<string | null>(queryClientId);
  const [searchQuery, setSearchQuery] = useState('');
  const [isLoadingClients, setIsLoadingClients] = useState(true);

  // Состояние стрима выбранного ученика
  const [stream, setStream] = useState<StudyStream | null>(null);
  const [isLoadingStream, setIsLoadingStream] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Модалка отчета по уроку
  const [activeLessonForReport, setActiveLessonForReport] = useState<Lesson | null>(null);
  const [isReportModalOpen, setIsReportModalOpen] = useState(false);

  // Мобильный вид: показывать чат вместо сайдбара
  const [showMobileChat, setShowMobileChat] = useState<boolean>(Boolean(queryClientId));

  // 1. Загрузка списка учеников
  const loadClientsList = useCallback(async () => {
    setIsLoadingClients(true);
    try {
      const data = await getClients();
      // Оставляем только неархивных учеников
      const activeOnly = data.filter((c) => !c.is_archived);
      setClients(activeOnly);

      // Если в URL нет clientId, но есть ученики, по умолчанию выбираем первого (на десктопе)
      if (!queryClientId && activeOnly.length > 0 && window.innerWidth >= 1024) {
        setSelectedClientId(activeOnly[0].id);
        setSearchParams({ clientId: activeOnly[0].id }, { replace: true });
      }
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Ошибка загрузки клиентов');
    } finally {
      setIsLoadingClients(false);
    }
  }, [queryClientId, setSearchParams]);

  useEffect(() => {
    loadClientsList();
  }, [loadClientsList]);

  // Синхронизация при смене queryClientId в URL
  useEffect(() => {
    if (queryClientId && queryClientId !== selectedClientId) {
      setSelectedClientId(queryClientId);
      setShowMobileChat(true);
    }
  }, [queryClientId, selectedClientId]);

  // 2. Загрузка потока обучения при выборе ученика
  const loadStream = useCallback(async (clientId: string) => {
    setIsLoadingStream(true);
    setErrorMessage(null);
    try {
      const data = await getClientStudyStream(clientId);
      setStream(data);
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Ошибка загрузки потока обучения');
    } finally {
      setIsLoadingStream(false);
    }
  }, []);

  useEffect(() => {
    if (selectedClientId) {
      loadStream(selectedClientId);
    } else {
      setStream(null);
    }
  }, [selectedClientId, loadStream]);

  // Обработка queryLessonId (если перешли из расписания на конкретный урок)
  useEffect(() => {
    if (queryLessonId && selectedClientId) {
      getLessons({ client_id: selectedClientId }).then((lessons) => {
        const found = lessons.find((l) => l.id === queryLessonId);
        if (found) {
          setActiveLessonForReport(found);
          setIsReportModalOpen(true);
        }
      }).catch(() => {});
    }
  }, [queryLessonId, selectedClientId]);

  // Выбор ученика
  const handleSelectClient = (client: Client) => {
    setSelectedClientId(client.id);
    setShowMobileChat(true);
    setSearchParams({ clientId: client.id });
  };

  // Отправка свободной быстрой заметки
  const handleSendNote = async (content: string) => {
    if (!selectedClientId) return;
    const newNote = await createClientNote(selectedClientId, content);
    setStream((prev) => {
      if (!prev) return prev;
      const newItem: StudyStreamItem = {
        id: newNote.id,
        type: 'note',
        timestamp: newNote.created_at,
        note: newNote,
      };
      return {
        ...prev,
        items: [...prev.items, newItem],
      };
    });
  };

  // Удаление заметки
  const handleDeleteNote = async (noteId: string) => {
    await deleteClientNote(noteId);
    setStream((prev) => {
      if (!prev) return prev;
      return {
        ...prev,
        items: prev.items.filter((item) => item.id !== noteId),
      };
    });
  };

  // Открытие отчета по уроку
  const handleOpenLatestReport = async () => {
    if (!selectedClientId) return;
    try {
      const lessons = await getLessons({ client_id: selectedClientId });
      // Берем самый свежий урок
      if (lessons.length > 0) {
        setActiveLessonForReport(lessons[0]);
        setIsReportModalOpen(true);
      } else {
        alert('У этого ученика пока нет занятий в расписании');
      }
    } catch {
      alert('Не удалось загрузить уроки ученика');
    }
  };

  const handlePlanUpcomingLesson = (lessonId: string) => {
    getLessons({ client_id: selectedClientId || undefined }).then((lessons) => {
      const found = lessons.find((l) => l.id === lessonId);
      if (found) {
        setActiveLessonForReport(found);
        setIsReportModalOpen(true);
      }
    });
  };

  const selectedClient = clients.find((c) => c.id === selectedClientId);

  return (
    <div className="h-[calc(100vh-8.5rem)] flex flex-col space-y-4">
      {errorMessage && (
        <div className="p-3.5 rounded-2xl bg-rose-500/10 border border-rose-400/30 backdrop-blur-md flex items-center justify-between gap-3 text-rose-800 text-xs font-medium">
          <div className="flex items-center gap-2">
            <AlertCircle className="w-4 h-4 text-rose-600 shrink-0" />
            <span>{errorMessage}</span>
          </div>
          <button
            type="button"
            onClick={() => selectedClientId && loadStream(selectedClientId)}
            className="hover:underline font-semibold"
          >
            Повторить
          </button>
        </div>
      )}

      {/* Двухколоночный Master-Detail layout */}
      <div className="flex-1 flex gap-4 overflow-hidden relative">
        {/* Сайдбар учеников (скрывается на мобилках при открытом чате) */}
        <div className={`${showMobileChat ? 'hidden lg:flex' : 'flex'} w-full lg:w-80 shrink-0 h-full`}>
          <JournalSidebar
            clients={clients}
            selectedClientId={selectedClientId}
            onSelectClient={handleSelectClient}
            searchQuery={searchQuery}
            onSearchChange={setSearchQuery}
            isLoading={isLoadingClients}
          />
        </div>

        {/* Центральная рабочая область Study Stream */}
        <main
          className={`${
            !showMobileChat ? 'hidden lg:flex' : 'flex'
          } flex-1 flex flex-col h-full rounded-3xl liquid-glass border border-white/60 p-3 sm:p-4 shadow-sm overflow-hidden`}
        >
          {selectedClient ? (
            <div className="flex-1 flex flex-col h-full space-y-3 overflow-hidden">
              {/* Шапка чата */}
              <JournalChatHeader
                client={selectedClient}
                onBackMobile={() => setShowMobileChat(false)}
                onOpenCRM={(id) => navigate(`/teacher/clients?id=${id}`)}
              />

              {/* Компактный баннер ближайшего урока */}
              <UpcomingLessonBanner
                upcomingLesson={stream?.upcoming_lesson}
                onPlanLesson={handlePlanUpcomingLesson}
              />

              {/* Хронологический чат-стрим сообщений */}
              {isLoadingStream ? (
                <div className="flex-1 flex flex-col items-center justify-center space-y-2 text-slate-500">
                  <RefreshCw className="w-6 h-6 animate-spin text-indigo-600" />
                  <span className="text-xs font-medium">Загрузка потока обучения...</span>
                </div>
              ) : (
                <StudyStreamFeed
                  items={stream?.items || []}
                  onDeleteNote={handleDeleteNote}
                  isLoading={isLoadingStream}
                />
              )}

              {/* Нижняя панель быстрой отправки заметки */}
              <StudyStreamInputBar
                onSendNote={handleSendNote}
                onOpenReportModal={handleOpenLatestReport}
                disabled={isLoadingStream}
              />
            </div>
          ) : (
            <div className="flex-1 flex flex-col items-center justify-center p-8 text-center text-slate-500 space-y-3">
              <div className="w-16 h-16 rounded-3xl bg-indigo-50/60 border border-indigo-200/50 flex items-center justify-center text-indigo-500 shadow-sm">
                <BookOpen className="w-8 h-8" />
              </div>
              <div className="space-y-1">
                <h3 className="text-base font-bold text-slate-900">
                  Выберите ученика для просмотра дневника
                </h3>
                <p className="text-xs text-slate-500 max-w-sm">
                  В боковой панели отображаются все ученики, отсортированные по времени последних уроков.
                </p>
              </div>
            </div>
          )}
        </main>
      </div>

      {/* Модальное окно заполнения отчета по уроку */}
      {isReportModalOpen && activeLessonForReport && (
        <LessonJournalModal
          isOpen={isReportModalOpen}
          onClose={() => {
            setIsReportModalOpen(false);
            setActiveLessonForReport(null);
          }}
          lesson={activeLessonForReport}
          clientDisplayName={selectedClient?.name || 'Ученик'}
          onSaved={() => {
            if (selectedClientId) {
              loadStream(selectedClientId);
            }
          }}
        />
      )}
    </div>
  );
};

export default TeacherJournalPage;
