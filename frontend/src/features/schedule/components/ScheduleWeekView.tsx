import React, { useState, useRef, useEffect } from 'react';
import { GlassCard } from '../../../shared/components/GlassCard';
import { Badge } from '../../../shared/components/Badge';
import { Lesson, LessonFormat } from '../../../types/schedule';
import { PositionedLesson } from '../types';
import { START_HOUR, HOUR_HEIGHT } from '../hooks/useSchedulePositioning';
import { Check, MapPin, Video } from 'lucide-react';

interface DragState {
  isDragging: boolean;
  day: Date;
  dayKey: string;
  startHour: number;
  currentHour: number;
}

interface ScheduleWeekViewProps {
  currentDate: Date;
  setCurrentDate: (date: Date) => void;
  weekDays: Date[];
  hours: number[];
  lessons: Lesson[];
  positionLessons: (items: Lesson[]) => PositionedLesson[];
  onSlotClick: (date: Date, hour: number) => void;
  onSlotDragSelect?: (date: Date, startHour: number, durationMinutes: number) => void;
  onQuickComplete: (e: React.MouseEvent, lessonId: string) => void;
  onOpenEdit: (lesson: Lesson) => void;
  getClientDisplayName: (lesson: Lesson) => string;
}

export const ScheduleWeekView: React.FC<ScheduleWeekViewProps> = ({
  currentDate,
  setCurrentDate,
  weekDays,
  hours,
  lessons,
  positionLessons,
  onSlotClick,
  onSlotDragSelect,
  onQuickComplete,
  onOpenEdit,
  getClientDisplayName,
}) => {
  const [dragState, setDragState] = useState<DragState | null>(null);
  const dragStateRef = useRef<DragState | null>(null);
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
            onSlotDragSelect(current.day, minHour, durationMinutes);
          } else {
            onSlotClick(current.day, minHour);
          }
        } else {
          onSlotClick(current.day, minHour);
        }
      }
      dragStateRef.current = null;
      setDragState(null);
    };

    window.addEventListener('mouseup', handleMouseUp);
    return () => {
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [onSlotClick, onSlotDragSelect]);

  const handleSlotMouseDown = (e: React.MouseEvent, day: Date, hour: number) => {
    if (e.button !== 0) return;
    e.preventDefault();
    const dayKey = day.toISOString().split('T')[0];
    const state: DragState = {
      isDragging: true,
      day,
      dayKey,
      startHour: hour,
      currentHour: hour,
    };
    dragStateRef.current = state;
    setDragState(state);
  };

  const handleSlotMouseEnter = (day: Date, hour: number) => {
    const current = dragStateRef.current;
    if (!current || !current.isDragging) return;
    const dayKey = day.toISOString().split('T')[0];
    if (current.dayKey !== dayKey) return;
    if (current.currentHour === hour) return;

    const next = { ...current, currentHour: hour };
    dragStateRef.current = next;
    setDragState(next);
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
    <GlassCard className="p-3 sm:p-5 overflow-x-auto">
      <div className="min-w-[760px] select-none">
        {/* Шапка 7 колонок дней недели */}
        <div className="grid grid-cols-[56px_repeat(7,1fr)] gap-0 border-b border-slate-200/60 pb-2 mb-2">
          <div />
          {weekDays.map((day) => {
            const isToday = day.toDateString() === new Date().toDateString();
            const isSelected = day.toDateString() === currentDate.toDateString();

            return (
              <div
                key={day.toISOString()}
                onClick={() => setCurrentDate(day)}
                className={`text-center py-2 px-1 rounded-2xl cursor-pointer transition-all ${
                  isSelected
                    ? 'bg-indigo-600/10 border border-indigo-500/20'
                    : 'hover:bg-white/40'
                }`}
              >
                <span className="block text-[11px] font-semibold text-slate-500 uppercase tracking-wider">
                  {day.toLocaleDateString('ru-RU', { weekday: 'short' })}
                </span>
                <span
                  className={`inline-flex items-center justify-center w-7 h-7 mt-0.5 rounded-full text-sm font-bold ${
                    isToday
                      ? 'bg-indigo-600 text-white shadow-sm'
                      : 'text-slate-800'
                  }`}
                >
                  {day.getDate()}
                </span>
              </div>
            );
          })}
        </div>

        {/* Сетка времени и колонок */}
        <div className="relative grid grid-cols-[56px_repeat(7,1fr)] gap-0">
          {/* Левая шкала времени */}
          <div className="relative">
            {hours.map((hour) => (
              <div
                key={hour}
                className="border-b border-transparent text-right pr-3 text-xs font-mono text-slate-400 -mt-2.5"
                style={{ height: `${HOUR_HEIGHT}px` }}
              >
                {String(hour).padStart(2, '0')}:00
              </div>
            ))}
          </div>

          {/* 7 колонок дней недели */}
          {weekDays.map((day) => {
            const dayStr = day.toISOString().split('T')[0];
            const dayItems = lessons.filter((l) => l.start_time.startsWith(dayStr));
            const positionedItems = positionLessons(dayItems);

            return (
              <div
                key={day.toISOString()}
                className="relative border-l border-slate-200/60"
                style={{ height: `${hours.length * HOUR_HEIGHT}px` }}
              >
                {/* Фоновые горизонтальные линии и кликабельные слоты */}
                {hours.map((hour) => (
                  <div
                    key={hour}
                    onMouseDown={(e) => handleSlotMouseDown(e, day, hour)}
                    onMouseEnter={() => handleSlotMouseEnter(day, hour)}
                    className="border-b border-slate-200/50 hover:bg-indigo-50/30 cursor-pointer transition-colors"
                    style={{ height: `${HOUR_HEIGHT}px` }}
                    title={`Выделите или нажмите на ${day.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })} в ${hour}:00`}
                  />
                ))}

                {/* Интерактивный Liquid Glass оверлей выделения диапазона при перетаскивании мышью */}
                {dragState && dragState.dayKey === dayStr && (
                  <div
                    className="absolute left-1 right-1 pointer-events-none z-20 rounded-2xl bg-indigo-500/25 border-2 border-indigo-400/60 backdrop-blur-[2px] p-2 flex flex-col justify-between shadow-sm animate-in fade-in transition-all duration-75"
                    style={{
                      top: `${(Math.min(dragState.startHour, dragState.currentHour) - START_HOUR) * HOUR_HEIGHT}px`,
                      height: `${(Math.abs(dragState.currentHour - dragState.startHour) + 1) * HOUR_HEIGHT}px`,
                    }}
                  >
                    <div className="flex items-center gap-1.5 text-xs font-bold text-indigo-950 bg-white/90 backdrop-blur-md rounded-xl px-2 py-0.5 w-fit shadow-xs">
                      <span>
                        {String(Math.min(dragState.startHour, dragState.currentHour)).padStart(2, '0')}:00 –{' '}
                        {String(Math.max(dragState.startHour, dragState.currentHour) + 1).padStart(2, '0')}:00
                      </span>
                      <span className="text-indigo-600 font-mono text-[10px]">
                        ({(Math.abs(dragState.currentHour - dragState.startHour) + 1).toFixed(1)} ч)
                      </span>
                    </div>
                  </div>
                )}

                {/* Размещение уроков в колонке дня */}
                {positionedItems.map((lesson) => {
                  const topOffset =
                    ((lesson.startMinutes - START_HOUR * 60) / 60) * HOUR_HEIGHT;
                  const cardHeight = (lesson.durationMinutes / 60) * HOUR_HEIGHT;
                  const widthPercent = 100 / lesson.totalColumns;
                  const leftPercent = lesson.column * widthPercent;
                  const studentName = getClientDisplayName(lesson);
                  const isCompleted = lesson.status === 'completed';
                  const isCancelled =
                    lesson.status === 'cancelled' ||
                    lesson.status.startsWith('cancelled') ||
                    lesson.status === 'declined';

                  return (
                    <div
                      key={lesson.id}
                      onClick={() => onOpenEdit(lesson)}
                      onMouseDown={(e) => e.stopPropagation()}
                      className="absolute p-0.5 pointer-events-auto transition-all group z-10 cursor-pointer"
                      style={{
                        top: `${Math.max(0, topOffset)}px`,
                        height: `${Math.max(48, cardHeight)}px`,
                        left: `${leftPercent}%`,
                        width: `${widthPercent}%`,
                      }}
                    >
                      <div
                        className={`h-full w-full rounded-2xl p-2 sm:p-2.5 flex flex-col justify-between shadow-sm relative overflow-hidden transition-all group-hover:scale-[1.02] group-hover:shadow-md border ${
                          isCancelled
                            ? 'opacity-60 bg-rose-50/20 border-rose-300/40 text-slate-400'
                            : isCompleted
                            ? 'bg-emerald-500/20 border-emerald-400/50 text-emerald-950'
                            : 'bg-white/85 backdrop-blur-md border-white/80'
                        }`}
                        style={{
                          borderLeftWidth: '4px',
                          borderLeftColor: isCancelled
                            ? '#F43F5E'
                            : lesson.classroom_color ||
                              (lesson.online_link || lesson.format === 'online' ? '#10B981' : '#4F46E5'),
                        }}
                        title={isCancelled && lesson.cancel_reason ? `Отменено: ${lesson.cancel_reason}` : undefined}
                      >
                        <div className="min-w-0">
                          {/* 1-я строка: Имя ученика и бейдж формата */}
                          <div className="flex items-start justify-between gap-1">
                            <div className="flex items-center gap-1 min-w-0 flex-wrap">
                              <span className={`font-bold text-xs sm:text-sm leading-tight truncate ${isCancelled ? 'line-through text-slate-400' : 'text-slate-900'}`}>
                                {studentName}
                              </span>
                              {isCancelled ? (
                                <Badge variant="danger" className="text-[9px] px-1.5 py-0">Отменено</Badge>
                              ) : (
                                getFormatBadge(lesson.format)
                              )}
                            </div>

                            {/* Кнопка быстрого подтверждения проведения ✓ */}
                            {!isCancelled && lesson.status === 'scheduled' && (
                              <button
                                type="button"
                                onClick={(e) => onQuickComplete(e, lesson.id)}
                                title="Отметить проведённым"
                                className="w-5 h-5 rounded-md bg-emerald-500 hover:bg-emerald-600 text-white flex items-center justify-center shrink-0 transition-all shadow-sm active:scale-95"
                              >
                                <Check className="w-3 h-3 stroke-[3]" />
                              </button>
                            )}
                          </div>

                          {/* 2-я строка: Время */}
                          <div className={`text-[10px] font-mono mt-0.5 truncate ${isCancelled ? 'line-through text-slate-400' : 'text-slate-600'}`}>
                            {new Date(lesson.start_time).toLocaleTimeString('ru-RU', {
                              hour: '2-digit',
                              minute: '2-digit',
                            })}{' '}
                            –{' '}
                            {new Date(lesson.end_time).toLocaleTimeString('ru-RU', {
                              hour: '2-digit',
                              minute: '2-digit',
                            })}
                          </div>

                          {/* Причина отмены */}
                          {isCancelled && lesson.cancel_reason && (
                            <div className="text-[9px] text-rose-500/90 truncate mt-0.5 font-medium" title={lesson.cancel_reason}>
                              {lesson.cancel_reason}
                            </div>
                          )}

                          {/* 3-я строка: Кабинет или онлайн */}
                          {lesson.classroom_name ? (
                            <div className="text-[10px] text-slate-500 truncate flex items-center gap-1 mt-0.5">
                              <MapPin className="w-2.5 h-2.5 text-slate-400 shrink-0" />
                              <span className="truncate">{lesson.classroom_name}</span>
                            </div>
                          ) : (lesson.online_link || lesson.format === 'online') ? (
                            <div className="text-[10px] text-emerald-600 font-medium truncate flex items-center gap-1 mt-0.5">
                              <Video className="w-2.5 h-2.5 shrink-0" />
                              <span>Онлайн</span>
                            </div>
                          ) : null}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            );
          })}
        </div>
      </div>
    </GlassCard>
  );
};
