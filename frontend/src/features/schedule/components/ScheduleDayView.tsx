import React, { useState, useRef, useEffect } from 'react';
import { GlassCard } from '../../../shared/components/GlassCard';
import { Badge } from '../../../shared/components/Badge';
import { Lesson, LessonFormat } from '../../../types/schedule';
import { PositionedLesson } from '../types';
import { START_HOUR, HOUR_HEIGHT } from '../hooks/useSchedulePositioning';
import { Check, MapPin, Video } from 'lucide-react';

interface DayDragState {
  isDragging: boolean;
  startHour: number;
  currentHour: number;
}

interface ScheduleDayViewProps {
  currentDate: Date;
  hours: number[];
  positionedDayLessons: PositionedLesson[];
  onSlotClick: (date: Date, hour: number) => void;
  onSlotDragSelect?: (date: Date, startHour: number, durationMinutes: number) => void;
  onQuickComplete: (e: React.MouseEvent, lessonId: string) => void;
  onOpenEdit: (lesson: Lesson) => void;
  getClientDisplayName: (lesson: Lesson) => string;
}

export const ScheduleDayView: React.FC<ScheduleDayViewProps> = ({
  currentDate,
  hours,
  positionedDayLessons,
  onSlotClick,
  onSlotDragSelect,
  onQuickComplete,
  onOpenEdit,
  getClientDisplayName,
}) => {
  const [dragState, setDragState] = useState<DayDragState | null>(null);
  const dragStateRef = useRef<DayDragState | null>(null);
  dragStateRef.current = dragState;

  useEffect(() => {
    const handleMouseUp = () => {
      const current = dragStateRef.current;
      if (current && current.isDragging) {
        const minHour = Math.min(current.startHour, current.currentHour);
        const maxHour = Math.max(current.startHour, current.currentHour);
        const durationHours = maxHour - minHour + 1;
        const durationMinutes = durationHours * 60;

        if (durationMinutes > 60 || current.startHour !== current.currentHour) {
          if (onSlotDragSelect) {
            onSlotDragSelect(currentDate, minHour, durationMinutes);
          } else {
            onSlotClick(currentDate, minHour);
          }
        } else {
          onSlotClick(currentDate, minHour);
        }
      }
      dragStateRef.current = null;
      setDragState(null);
    };

    window.addEventListener('mouseup', handleMouseUp);
    return () => {
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [currentDate, onSlotClick, onSlotDragSelect]);

  const handleSlotMouseDown = (e: React.MouseEvent, hour: number) => {
    if (e.button !== 0) return;
    e.preventDefault();
    const state: DayDragState = {
      isDragging: true,
      startHour: hour,
      currentHour: hour,
    };
    dragStateRef.current = state;
    setDragState(state);
  };

  const handleSlotMouseEnter = (hour: number) => {
    const current = dragStateRef.current;
    if (!current || !current.isDragging) return;
    if (current.currentHour === hour) return;

    const next = { ...current, currentHour: hour };
    dragStateRef.current = next;
    setDragState(next);
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
    <GlassCard className="p-4 sm:p-6 overflow-hidden">
      <div className="relative border-t border-slate-200/60 select-none">
        {/* Фоновые линии шкалы времени */}
        <div className="relative">
          {hours.map((hour) => (
            <div
              key={hour}
              onMouseDown={(e) => handleSlotMouseDown(e, hour)}
              onMouseEnter={() => handleSlotMouseEnter(hour)}
              className="flex items-start border-b border-slate-200/50 hover:bg-indigo-50/20 cursor-pointer transition-colors"
              style={{ height: `${HOUR_HEIGHT}px` }}
              title={`Выделите или нажмите для занятия в ${hour}:00`}
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
          {/* Интерактивный Liquid Glass оверлей выделения диапазона */}
          {dragState && (
            <div
              className="absolute left-2 right-2 pointer-events-none z-20 rounded-2xl bg-indigo-500/25 border-2 border-indigo-400/60 backdrop-blur-[2px] p-2 flex flex-col justify-between shadow-sm animate-in fade-in transition-all duration-75"
              style={{
                top: `${(Math.min(dragState.startHour, dragState.currentHour) - START_HOUR) * HOUR_HEIGHT}px`,
                height: `${(Math.abs(dragState.currentHour - dragState.startHour) + 1) * HOUR_HEIGHT}px`,
              }}
            >
              <div className="flex items-center gap-1.5 text-xs font-bold text-indigo-950 bg-white/90 backdrop-blur-md rounded-xl px-2.5 py-1 w-fit shadow-xs">
                <span>
                  {String(Math.min(dragState.startHour, dragState.currentHour)).padStart(2, '0')}:00 –{' '}
                  {String(Math.max(dragState.startHour, dragState.currentHour) + 1).padStart(2, '0')}:00
                </span>
                <span className="text-indigo-600 font-mono text-[11px]">
                  ({(Math.abs(dragState.currentHour - dragState.startHour) + 1).toFixed(1)} ч)
                </span>
              </div>
            </div>
          )}

          {positionedDayLessons.map((lesson) => {
            const topOffset =
              ((lesson.startMinutes - START_HOUR * 60) / 60) * HOUR_HEIGHT;
            const cardHeight = (lesson.durationMinutes / 60) * HOUR_HEIGHT;
            const widthPercent = 100 / lesson.totalColumns;
            const leftPercent = lesson.column * widthPercent;
            const isCancelled =
              lesson.status === 'cancelled' ||
              lesson.status.startsWith('cancelled') ||
              lesson.status === 'declined';
            const accentColor = isCancelled
              ? '#F43F5E'
              : lesson.classroom_color ||
                (lesson.online_link || lesson.format === 'online' ? '#10B981' : '#4F46E5');
            const studentName = getClientDisplayName(lesson);
            const isCompleted = lesson.status === 'completed';

            return (
              <div
                key={lesson.id}
                onClick={() => onOpenEdit(lesson)}
                onMouseDown={(e) => e.stopPropagation()}
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
                    isCancelled
                      ? 'opacity-60 bg-rose-50/20 border-rose-300/40 text-slate-400'
                      : isCompleted
                      ? 'bg-emerald-500/20 border-emerald-400/50'
                      : 'bg-white/85 backdrop-blur-md border-white/80'
                  }`}
                  style={{
                    borderLeftWidth: '5px',
                    borderLeftColor: accentColor,
                  }}
                  title={isCancelled && lesson.cancel_reason ? `Отменено: ${lesson.cancel_reason}` : undefined}
                >
                  <div>
                    {/* 1-я строка: Имя ученика и статус */}
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex items-center gap-1.5 min-w-0">
                        <span className={`font-bold text-sm sm:text-base truncate ${isCancelled ? 'line-through text-slate-400' : 'text-slate-900'}`}>
                          {studentName}
                        </span>
                        {getFormatBadge(lesson.format)}
                      </div>
                      <div className="flex items-center gap-1.5">
                        {isCancelled ? (
                          <div className="flex items-center gap-1">
                            <Badge variant="danger" className="text-[9px]">Отменено</Badge>
                            {lesson.cancel_reason && (
                              <span className="text-xs text-rose-600/90 font-medium truncate max-w-[120px]" title={lesson.cancel_reason}>
                                ({lesson.cancel_reason})
                              </span>
                            )}
                          </div>
                        ) : (
                          getStatusBadge(lesson.status)
                        )}
                        {!isCancelled && lesson.status === 'scheduled' && (
                          <button
                            type="button"
                            onClick={(e) => onQuickComplete(e, lesson.id)}
                            title="Отметить проведённым"
                            className="w-6 h-6 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-white flex items-center justify-center transition-all shadow-sm active:scale-95"
                          >
                            <Check className="w-3.5 h-3.5 stroke-[3]" />
                          </button>
                        )}
                      </div>
                    </div>

                    {/* 2-я строка: Время и тема */}
                    <div className={`flex items-center gap-2 text-xs mt-1 truncate ${isCancelled ? 'line-through text-slate-400' : 'text-slate-600'}`}>
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
                    {lesson.classroom_name ? (
                      <div className="flex items-center gap-1 text-xs font-medium text-slate-600 mt-1">
                        <MapPin className="w-3.5 h-3.5 text-slate-400" />
                        <span className="truncate">{lesson.classroom_name}</span>
                      </div>
                    ) : (lesson.online_link || lesson.format === 'online') ? (
                      <div className="flex items-center gap-1 text-xs font-medium text-emerald-600 mt-1">
                        <Video className="w-3.5 h-3.5" />
                        <span>Онлайн занятие</span>
                      </div>
                    ) : null}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </GlassCard>
  );
};
