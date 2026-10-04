import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { GlassCard } from '../../shared/components/GlassCard';
import { GlassButton } from '../../shared/components/GlassButton';
import { GlassModal } from '../../shared/components/GlassModal';
import { Badge } from '../../shared/components/Badge';
import { AppNavbar } from '../../shared/components/AppNavbar';
import { Lesson } from '../../types/schedule';
import { getLessons, acceptLesson, declineLesson } from '../../api/schedule';
import {
  BookOpen,
  Calendar,
  Clock,
  Video,
  MapPin,
  CheckCircle,
  XCircle,
  ExternalLink,
  Sparkles,
  AlertCircle,
  Building2,
} from 'lucide-react';

export const StudentLessonsPage: React.FC = () => {
  const [lessons, setLessons] = useState<Lesson[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Decline Modal State
  const [declineLessonId, setDeclineLessonId] = useState<string | null>(null);
  const [declineReason, setDeclineReason] = useState('');
  const [isSubmittingDecline, setIsSubmittingDecline] = useState(false);

  // Filter tabs
  const [activeTab, setActiveTab] = useState<'pending' | 'upcoming' | 'history'>('pending');

  const loadLessons = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await getLessons();
      setLessons(data);
    } catch (err) {
      console.error('Failed to load student lessons', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadLessons();
  }, [loadLessons]);

  // Groups
  const pendingLessons = useMemo(
    () => lessons.filter((l) => l.status === 'pending_confirmation'),
    [lessons]
  );

  const upcomingLessons = useMemo(
    () => lessons.filter((l) => l.status === 'confirmed'),
    [lessons]
  );

  const historyLessons = useMemo(
    () => lessons.filter((l) => ['completed', 'no_show', 'declined', 'cancelled'].includes(l.status)),
    [lessons]
  );

  // Accept Lesson in 1 click
  const handleAccept = async (lessonId: string) => {
    try {
      await acceptLesson(lessonId);
      await loadLessons();
    } catch (err) {
      console.error('Failed to accept lesson', err);
    }
  };

  // Open Decline Modal
  const openDeclineModal = (lessonId: string) => {
    setDeclineLessonId(lessonId);
    setDeclineReason('');
  };

  // Submit Decline with reason
  const handleDeclineSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!declineLessonId) return;

    setIsSubmittingDecline(true);
    try {
      await declineLesson(declineLessonId, declineReason.trim() || 'Не смогу присутствовать');
      setDeclineLessonId(null);
      await loadLessons();
    } catch (err) {
      console.error('Failed to decline lesson', err);
    } finally {
      setIsSubmittingDecline(false);
    }
  };

  const getStatusBadge = (status: Lesson['status']) => {
    switch (status) {
      case 'confirmed':
        return <Badge variant="mint">Подтверждено</Badge>;
      case 'pending_confirmation':
        return <Badge variant="amber">Требуется подтверждение</Badge>;
      case 'completed':
        return <Badge variant="mint">Занятие завершено</Badge>;
      case 'no_show':
        return <Badge variant="coral">Неявка</Badge>;
      case 'declined':
        return <Badge variant="coral">Отклонено</Badge>;
      default:
        return <Badge variant="neutral">{status}</Badge>;
    }
  };

  return (
    <div className="min-h-screen bg-[#F5F5F7] text-slate-800 p-4 sm:p-8 relative overflow-hidden">
      {/* Liquid Glass Background Orbs */}
      <div className="absolute top-0 right-1/4 w-[500px] h-[500px] bg-emerald-500/10 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute bottom-10 left-10 w-[450px] h-[450px] bg-indigo-500/10 rounded-full blur-3xl pointer-events-none" />

      <div className="max-w-5xl mx-auto space-y-6 relative z-10">
        <AppNavbar />

        {/* Title */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">
              Мои Занятия
            </h1>
            <p className="text-sm text-slate-500 mt-1">
              Подтверждайте предложенные уроки репетиторов и подключайтесь к созвонам
            </p>
          </div>

          {/* Tab selector */}
          <div className="flex items-center gap-1.5 p-1 rounded-2xl liquid-glass border border-white/60 self-start sm:self-auto shadow-sm">
            <button
              onClick={() => setActiveTab('pending')}
              className={`flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-semibold transition-all ${
                activeTab === 'pending'
                  ? 'bg-amber-500 text-white shadow-md shadow-amber-500/25'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
              }`}
            >
              <AlertCircle className="w-3.5 h-3.5" />
              <span>Ожидают ({pendingLessons.length})</span>
            </button>
            <button
              onClick={() => setActiveTab('upcoming')}
              className={`flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-semibold transition-all ${
                activeTab === 'upcoming'
                  ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
              }`}
            >
              <Calendar className="w-3.5 h-3.5" />
              <span>Предстоящие ({upcomingLessons.length})</span>
            </button>
            <button
              onClick={() => setActiveTab('history')}
              className={`flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-semibold transition-all ${
                activeTab === 'history'
                  ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-white/40'
              }`}
            >
              <BookOpen className="w-3.5 h-3.5" />
              <span>История ({historyLessons.length})</span>
            </button>
          </div>
        </div>

        {/* ------------------------------------------------------------- */}
        {/* TAB 1: ТРЕБУЮТ ПОДТВЕРЖДЕНИЯ (Яркие интерактивные карточки) */}
        {/* ------------------------------------------------------------- */}
        {activeTab === 'pending' && (
          <div className="space-y-4">
            {pendingLessons.length > 0 && (
              <div className="flex items-center gap-2 px-4 py-2.5 rounded-2xl bg-amber-500/10 border border-amber-500/20 text-amber-800 text-xs font-semibold">
                <Sparkles className="w-4 h-4 text-amber-600 shrink-0" />
                <span>
                  Преподаватель предложил вам время занятий. Подтвердите его в 1 клик или укажите причину переноса.
                </span>
              </div>
            )}

            <div className="space-y-4">
              {pendingLessons.map((lesson) => (
                <GlassCard
                  key={lesson.id}
                  className="p-6 relative overflow-hidden border-2 border-amber-300/40 shadow-xl bg-gradient-to-br from-white/90 via-amber-50/20 to-white/70"
                >
                  {/* Decorative Amber Glow */}
                  <div className="absolute top-0 left-0 bottom-0 w-2.5 bg-gradient-to-b from-amber-400 to-amber-500" />

                  <div className="pl-2">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
                      <div>
                        <div className="flex items-center gap-2.5">
                          <h2 className="text-lg font-bold text-slate-900">
                            {lesson.title}
                          </h2>
                          {getStatusBadge(lesson.status)}
                        </div>
                        <p className="text-xs text-slate-500 mt-1">
                          Преподаватель: <span className="font-semibold text-slate-800">{lesson.teacher_name || 'Александр Верников'}</span>
                        </p>
                      </div>

                      <div className="flex items-center gap-2 text-xs font-medium text-slate-600 bg-white/70 px-3.5 py-1.5 rounded-xl border border-white/80 self-start sm:self-auto">
                        <Clock className="w-4 h-4 text-amber-500" />
                        <span className="font-bold">
                          {new Date(lesson.start_time).toLocaleDateString('ru-RU', {
                            weekday: 'short',
                            day: 'numeric',
                            month: 'long',
                          })}
                        </span>
                        <span>•</span>
                        <span>
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
                    </div>

                    {/* Location or Online link details */}
                    <div className="p-3.5 rounded-2xl liquid-glass border border-white/60 mb-5 flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div className="w-9 h-9 rounded-xl bg-indigo-50 border border-indigo-200/50 flex items-center justify-center text-indigo-600">
                          {lesson.format === 'online' ? (
                            <Video className="w-4 h-4" />
                          ) : (
                            <MapPin className="w-4 h-4" />
                          )}
                        </div>
                        <div>
                          <span className="text-xs font-bold text-slate-900 block">
                            {lesson.format === 'online'
                              ? 'Онлайн видеоконференция'
                              : lesson.classroom_name || 'Кабинет в школе'}
                          </span>
                          <span className="text-[11px] text-slate-500">
                            {lesson.format === 'online'
                              ? 'Ссылка станет активна сразу после подтверждения урока'
                              : 'Адрес школы: ул. Ломоносова, д. 12, главный корпус'}
                          </span>
                        </div>
                      </div>
                    </div>

                    {lesson.comment && (
                      <p className="text-xs text-slate-600 bg-white/40 p-3 rounded-xl mb-4 italic">
                        «{lesson.comment}»
                      </p>
                    )}

                    {/* Action Buttons: Принять (в 1 клик) / Отклонить */}
                    <div className="flex items-center gap-3 pt-2">
                      <GlassButton
                        variant="mint"
                        onClick={() => handleAccept(lesson.id)}
                        icon={<CheckCircle className="w-4 h-4" />}
                        className="shadow-emerald-500/25"
                      >
                        Принять занятие
                      </GlassButton>

                      <GlassButton
                        variant="coral"
                        onClick={() => openDeclineModal(lesson.id)}
                        icon={<XCircle className="w-4 h-4" />}
                      >
                        Отклонить
                      </GlassButton>
                    </div>
                  </div>
                </GlassCard>
              ))}

              {pendingLessons.length === 0 && !isLoading && (
                <div className="p-12 text-center liquid-glass rounded-3xl">
                  <CheckCircle className="w-12 h-12 text-emerald-400 mx-auto mb-3" />
                  <p className="text-base font-medium text-slate-800">
                    Нет ожидающих подтверждения уроков
                  </p>
                  <p className="text-xs text-slate-400 mt-1">
                    Все назначенные занятия согласованы и готовы к проведению
                  </p>
                </div>
              )}
            </div>
          </div>
        )}

        {/* ------------------------------------------------------------- */}
        {/* TAB 2: ПРЕДСТОЯЩИЕ ПОДТВЕРЖДЕННЫЕ ЗАНЯТИЯ */}
        {/* ------------------------------------------------------------- */}
        {activeTab === 'upcoming' && (
          <div className="space-y-4">
            {upcomingLessons.map((lesson) => (
              <GlassCard key={lesson.id} className="p-6 relative overflow-hidden group">
                <div
                  className="absolute top-0 left-0 right-0 h-1.5"
                  style={{
                    backgroundColor:
                      lesson.classroom_color ||
                      (lesson.format === 'online' ? '#10B981' : '#4F46E5'),
                  }}
                />

                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pt-1">
                  <div>
                    <div className="flex items-center gap-2 flex-wrap">
                      <h3 className="text-base font-bold text-slate-900">
                        {lesson.title}
                      </h3>
                      {getStatusBadge(lesson.status)}
                    </div>

                    <div className="flex items-center gap-3 text-xs text-slate-500 mt-1.5 flex-wrap">
                      <span className="font-semibold text-slate-800">
                        {lesson.teacher_name || 'Александр Верников'}
                      </span>
                      <span>•</span>
                      <span className="font-mono text-slate-700 font-medium">
                        {new Date(lesson.start_time).toLocaleDateString('ru-RU', {
                          day: 'numeric',
                          month: 'long',
                          weekday: 'short',
                        })}
                        ,{' '}
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
                      <div className="mt-2.5 flex items-center gap-2 text-xs font-semibold text-slate-700 bg-white/70 px-3 py-1.5 rounded-xl border border-white/80 w-fit">
                        <Building2 className="w-4 h-4 text-indigo-600" />
                        <span>{lesson.classroom_name}</span>
                        <span className="text-slate-400 font-normal">| Главный корпус</span>
                      </div>
                    )}
                  </div>

                  {/* Онлайн созвон */}
                  {lesson.format === 'online' && (
                    <div className="self-start sm:self-auto">
                      <a
                        href={lesson.online_link || 'https://telemost.yandex.ru/'}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex items-center gap-2 px-5 py-2.5 rounded-2xl bg-emerald-500 hover:bg-emerald-600 text-white font-semibold text-xs shadow-lg shadow-emerald-500/25 border border-emerald-400/30 transition-all hover:scale-[1.02] active:scale-[0.98]"
                      >
                        <Video className="w-4 h-4" />
                        <span>Подключиться к созвону</span>
                        <ExternalLink className="w-3.5 h-3.5 opacity-80" />
                      </a>
                    </div>
                  )}
                </div>
              </GlassCard>
            ))}

            {upcomingLessons.length === 0 && !isLoading && (
              <div className="p-12 text-center liquid-glass rounded-3xl">
                <Calendar className="w-12 h-12 text-slate-300 mx-auto mb-3" />
                <p className="text-base font-medium text-slate-800">
                  Нет запланированных подтвержденных уроков
                </p>
                <p className="text-xs text-slate-400 mt-1">
                  Новые уроки отобразятся во вкладке «Ожидают» после назначения преподавателем
                </p>
              </div>
            )}
          </div>
        )}

        {/* ------------------------------------------------------------- */}
        {/* TAB 3: ИСТОРИЯ ЗАНЯТИЙ */}
        {/* ------------------------------------------------------------- */}
        {activeTab === 'history' && (
          <div className="space-y-3">
            {historyLessons.map((lesson) => (
              <GlassCard key={lesson.id} className="p-4 sm:p-5 flex items-center justify-between gap-4">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="font-bold text-slate-900 text-sm">{lesson.title}</span>
                    {getStatusBadge(lesson.status)}
                  </div>
                  <div className="text-xs text-slate-500">
                    {new Date(lesson.start_time).toLocaleDateString('ru-RU', {
                      day: 'numeric',
                      month: 'long',
                      year: 'numeric',
                    })}{' '}
                    • {lesson.teacher_name || 'Преподаватель'}
                  </div>
                  {lesson.decline_reason && (
                    <div className="text-xs text-rose-600 bg-rose-50 px-2.5 py-1 rounded-lg border border-rose-200/50 mt-1">
                      Причина отказа: {lesson.decline_reason}
                    </div>
                  )}
                </div>

                <div className="text-xs text-slate-400 font-mono">
                  {lesson.format === 'online' ? 'Онлайн' : 'Оффлайн'}
                </div>
              </GlassCard>
            ))}

            {historyLessons.length === 0 && !isLoading && (
              <div className="p-12 text-center liquid-glass rounded-3xl text-slate-500 text-sm">
                История занятий пока пуста
              </div>
            )}
          </div>
        )}
      </div>

      {/* ------------------------------------------------------------- */}
      {/* MODAL: ОТКЛОНЕНИЕ ЗАНЯТИЯ */}
      {/* ------------------------------------------------------------- */}
      <GlassModal
        isOpen={!!declineLessonId}
        onClose={() => setDeclineLessonId(null)}
        title="Отклонить занятие"
        description="Укажите причину для преподавателя, чтобы подобрать другое удобное время"
      >
        <form onSubmit={handleDeclineSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
              Причина отказа / комментарий
            </label>
            <textarea
              rows={3}
              placeholder="например: В это время у меня тренировка / контрольная, прошу перенести на вечер"
              value={declineReason}
              onChange={(e) => setDeclineReason(e.target.value)}
              className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 liquid-glass-input focus:border-rose-500 focus:outline-none resize-none"
              required
            />
          </div>

          <div className="flex items-center justify-end gap-3 pt-3 border-t border-black/[0.05]">
            <GlassButton
              type="button"
              variant="secondary"
              onClick={() => setDeclineLessonId(null)}
            >
              Отмена
            </GlassButton>
            <GlassButton
              type="submit"
              variant="coral"
              isLoading={isSubmittingDecline}
            >
              Отклонить занятие
            </GlassButton>
          </div>
        </form>
      </GlassModal>
    </div>
  );
};
