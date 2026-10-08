import React from 'react';
import { GlassCard } from '../../../shared/components/GlassCard';
import { GlassButton } from '../../../shared/components/GlassButton';
import { Badge } from '../../../shared/components/Badge';
import { Lesson, LessonFormat } from '../../../types/schedule';
import { Calendar as CalendarIcon, Check, Edit2, Clock, MapPin, Video, RotateCw, BookOpen } from 'lucide-react';

interface ScheduleListViewProps {
  dayLessons: Lesson[];
  isLoading: boolean;
  onQuickComplete: (e: React.MouseEvent, lessonId: string) => void;
  onOpenJournal: (lesson: Lesson) => void;
  onOpenEdit: (lesson: Lesson) => void;
  getClientDisplayName: (lesson: Lesson) => string;
}

export const ScheduleListView: React.FC<ScheduleListViewProps> = ({
  dayLessons,
  isLoading,
  onQuickComplete,
  onOpenJournal,
  onOpenEdit,
  getClientDisplayName,
}) => {
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
      case 'cancelled_by_teacher':
      case 'cancelled_by_student':
        return <Badge variant="neutral">Отменён</Badge>;
      default:
        return <Badge variant="neutral">{status}</Badge>;
    }
  };

  const getFormatBadge = (format: LessonFormat) => {
    switch (format) {
      case 'pair':
        return (
          <span className="inline-flex items-center px-1.5 py-0.5 rounded-md text-[9px] font-bold bg-purple-500/15 text-purple-700 border border-purple-400/30">
            Пара
          </span>
        );
      case 'group':
        return (
          <span className="inline-flex items-center px-1.5 py-0.5 rounded-md text-[9px] font-bold bg-amber-500/15 text-amber-700 border border-amber-400/30">
            Группа
          </span>
        );
      case 'individual':
      default:
        return (
          <span className="inline-flex items-center px-1.5 py-0.5 rounded-md text-[9px] font-bold bg-indigo-500/15 text-indigo-700 border border-indigo-400/30">
            Индив.
          </span>
        );
    }
  };

  return (
    <div className="space-y-3">
      {dayLessons.map((lesson) => {
        const studentName = getClientDisplayName(lesson);
        const isCancelled =
          lesson.status === 'cancelled' ||
          lesson.status.startsWith('cancelled') ||
          lesson.status === 'declined';

        return (
          <GlassCard
            key={lesson.id}
            onClick={() => onOpenEdit(lesson)}
            className={`p-5 flex flex-col md:flex-row items-start md:items-center justify-between gap-4 cursor-pointer hover:shadow-md transition-all ${
              isCancelled ? 'opacity-60 bg-rose-50/20 border-rose-300/40 text-slate-400' : ''
            }`}
          >
            <div className="flex items-start gap-4">
              <div
                className="w-12 h-12 rounded-2xl flex items-center justify-center text-white shrink-0 shadow-sm"
                style={{
                  backgroundColor: isCancelled
                    ? '#F43F5E'
                    : lesson.classroom_color ||
                      (lesson.online_link || lesson.format === 'online' ? '#10B981' : '#4F46E5'),
                }}
              >
                {(lesson.online_link || lesson.format === 'online') ? (
                  <Video className="w-6 h-6" />
                ) : (
                  <MapPin className="w-6 h-6" />
                )}
              </div>
              <div>
                <div className="flex items-center gap-2 flex-wrap">
                  <h3 className={`font-bold text-base ${isCancelled ? 'line-through text-slate-400' : 'text-slate-900'}`}>{studentName}</h3>
                  {(lesson.is_recurring || lesson.series_id) && (
                    <span title="Регулярная серия занятий" className="text-indigo-600 shrink-0">
                      <RotateCw className="w-4 h-4" />
                    </span>
                  )}
                  {getFormatBadge(lesson.format)}
                  {isCancelled ? (
                    <Badge variant="danger" className="text-[9px]">Отменено</Badge>
                  ) : (
                    getStatusBadge(lesson.status)
                  )}
                  {isCancelled && lesson.cancel_reason && (
                    <span className="text-xs text-rose-600/90 font-medium">({lesson.cancel_reason})</span>
                  )}
                </div>
                <div className={`flex items-center gap-3 text-xs mt-1 flex-wrap ${isCancelled ? 'line-through text-slate-400' : 'text-slate-500'}`}>
                  <span className={`font-semibold ${isCancelled ? 'line-through text-slate-400' : 'text-slate-800'}`}>{lesson.title}</span>
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
                {lesson.classroom_name ? (
                  <div className="text-xs text-indigo-600 font-medium mt-1 flex items-center gap-1">
                    <MapPin className="w-3.5 h-3.5 text-slate-400" />
                    <span>Аудитория: {lesson.classroom_name}</span>
                  </div>
                ) : (lesson.online_link || lesson.format === 'online') ? (
                  <div className="text-xs text-emerald-600 font-medium mt-1 flex items-center gap-1">
                    <Video className="w-3.5 h-3.5" />
                    {lesson.online_link && lesson.online_link.startsWith('http') ? (
                      <a
                        href={lesson.online_link}
                        target="_blank"
                        rel="noopener noreferrer"
                        onClick={(e) => e.stopPropagation()}
                        className="underline hover:text-emerald-700"
                      >
                        Ссылка на звонок
                      </a>
                    ) : (
                      <span>Онлайн занятие</span>
                    )}
                  </div>
                ) : null}
              </div>
            </div>

            <div className="flex items-center gap-2 self-end md:self-auto flex-wrap">
              <GlassButton
                variant="secondary"
                size="sm"
                onClick={(e) => {
                  e.stopPropagation();
                  onOpenJournal(lesson);
                }}
                icon={<BookOpen className="w-3.5 h-3.5 text-indigo-600" />}
              >
                Дневник
              </GlassButton>
              {!isCancelled && lesson.status === 'scheduled' && (
                <GlassButton
                  variant="mint"
                  size="sm"
                  onClick={(e) => onQuickComplete(e, lesson.id)}
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
                  onOpenEdit(lesson);
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
  );
};
