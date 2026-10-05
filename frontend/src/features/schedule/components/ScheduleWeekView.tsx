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
  startSlot: number;
  currentSlot: number;
}

interface ScheduleWeekViewProps {
  currentDate: Date;
  setCurrentDate: (date: Date) => void;
  weekDays: Date[];
  hours: number[];
  lessons: Lesson[];
  positionLessons: (items: Lesson[]) => PositionedLesson[];
  onSlotClick: (date: Date, hour: number, minute?: number) => void;
  onSlotDragSelect?: (date: Date, startHour: number, startMinute: number, durationMinutes: number) => void;
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
        const minSlot = Math.min(current.startSlot, current.currentSlot);
        const maxSlot = Math.max(current.startSlot, current.currentSlot);
        const numSlots = maxSlot - minSlot + 1;
        const durationMinutes = numSlots * 30;
        const startTotalMinutes = START_HOUR * 60 + minSlot * 30;
        const startHour = Math.floor(startTotalMinutes / 60);
        const startMinute = startTotalMinutes % 60;

        if (numSlots > 1) {
          if (onSlotDragSelect) {
            onSlotDragSelect(current.day, startHour, startMinute, durationMinutes);
          } else {
            onSlotClick(current.day, startHour, startMinute);
          }
        } else {
          // Одиночный клик по слоту
          onSlotClick(current.day, startHour, startMinute);
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

  const handleSlotMouseDown = (e: React.MouseEvent, day: Date, hour: number, minute: number) => {
    if (e.button !== 0) return;
    e.preventDefault();
    const dayKey = day.toISOString().split('T')[0];
    const slot = (hour - START_HOUR) * 2 + (minute === 30 ? 1 : 0);
    const state: DragState = {
      isDragging: true,
      day,
      dayKey,
      startSlot: slot,
      currentSlot: slot,
    };
    dragStateRef.current = state;
    setDragState(state);
  };

  const handleSlotMouseEnter = (day: Date, hour: number, minute: number) => {
    const current = dragStateRef.current;
    if (!current || !current.isDragging) return;
    const dayKey = day.toISOString().split('T')[0];
    if (current.dayKey !== dayKey) return;
    const slot = (hour - START_HOUR) * 2 + (minute === 30 ? 1 : 0);
    if (current.currentSlot === slot) return;

    const next = { ...current, currentSlot: slot };
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
                className="border-b border-transparent text-right pr-2.5 text-xs font-mono text-slate-400 relative"
                style={{ height: `${HOUR_HEIGHT}px` }}
              >
                <span className="block -mt-2.5 font-semibold text-slate-600">{String(hour).padStart(2, '0')}:00</span>
                <span
                  className="block text-[10px] text-slate-300 absolute right-2.5"
                  style={{ top: `${HOUR_HEIGHT / 2 - 7}px` }}
                >
                  :30
                </span>
              </div>
            ))}
          </div>

          {/* 7 колонок дней недели */}
          {weekDays.map((day) => {
            const dayStr = day.toISOString().split('T')[0];
            const dayItems = lessons.filter((l) => l.start_time.startsWith(dayStr));
            const positionedItems = positionLessons(dayItems);
            const SLOT_HEIGHT = HOUR_HEIGHT / 2;

            return (
              <div
                key={day.toISOString()}
                className="relative border-l border-slate-200/60"
                style={{ height: `${hours.length * HOUR_HEIGHT}px` }}
              >
                {/* Фоновые горизонтальные линии и кликабельные слоты с шагом 30 минут */}
                {hours.map((hour) => (
                  <div key={hour} className="border-b border-slate-200/60" style={{ height: `${HOUR_HEIGHT}px` }}>
                    <div
                      onMouseDown={(e) => handleSlotMouseDown(e, day, hour, 0)}
                      onMouseEnter={() => handleSlotMouseEnter(day, hour, 0)}
                      className="border-b border-dashed border-slate-200/40 hover:bg-indigo-50/30 cursor-pointer transition-colors"
                      style={{ height: `${HOUR_HEIGHT / 2}px` }}
                      title={`${day.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })} в ${String(hour).padStart(2, '0')}:00`}
                    />
                    <div
                      onMouseDown={(e) => handleSlotMouseDown(e, day, hour, 30)}
                      onMouseEnter={() => handleSlotMouseEnter(day, hour, 30)}
                      className="hover:bg-indigo-50/30 cursor-pointer transition-colors"
                      style={{ height: `${HOUR_HEIGHT / 2}px` }}
                      title={`${day.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })} в ${String(hour).padStart(2, '0')}:30`}
                    />
                  </div>
                ))}

                {/* Интерактивный Liquid Glass оверлей выделения диапазона при перетаскивании мышью */}
                {dragState && dragState.dayKey === dayStr && (() => {
                  const minSlot = Math.min(dragState.startSlot, dragState.currentSlot);
                  const maxSlot = Math.max(dragState.startSlot, dragState.currentSlot);
                  const numSlots = maxSlot - minSlot + 1;
                  const durationMinutes = numSlots * 30;
                  const startTotalMinutes = START_HOUR * 60 + minSlot * 30;
                  const startH = Math.floor(startTotalMinutes / 60);
                  const startM = startTotalMinutes % 60;
                  const endTotalMinutes = startTotalMinutes + durationMinutes;
                  const endH = Math.floor(endTotalMinutes / 60);
                  const endM = endTotalMinutes % 60;

                  return (
                    <div
                      className="absolute left-1 right-1 pointer-events-none z-20 rounded-2xl bg-indigo-500/25 border-2 border-indigo-400/60 backdrop-blur-[2px] p-2 flex flex-col justify-between shadow-sm animate-in fade-in transition-all duration-75"
                      style={{
                        top: `${minSlot * SLOT_HEIGHT}px`,
                        height: `${numSlots * SLOT_HEIGHT}px`,
                      }}
                    >
                      <div className="flex items-center gap-1.5 text-xs font-bold text-indigo-950 bg-white/90 backdrop-blur-md rounded-xl px-2 py-0.5 w-fit shadow-xs">
                        <span>
                          {String(startH).padStart(2, '0')}:{String(startM).padStart(2, '0')} –{' '}
                          {String(endH).padStart(2, '0')}:{String(endM).padStart(2, '0')}
                        </span>
                        <span className="text-indigo-600 font-mono text-[10px]">
                          ({durationMinutes >= 60 ? `${(durationMinutes / 60).toFixed(1)} ч` : `${durationMinutes} мин`})
                        </span>
                      </div>
                    </div>
                  );
                })()}

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
                              {lesson.online_link && lesson.online_link.startsWith('http') ? (
                                <a
                                  href={lesson.online_link}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  onClick={(e) => e.stopPropagation()}
                                  className="underline hover:text-emerald-700 truncate"
                                >
                                  Звонок
                                </a>
                              ) : (
                                <span>Онлайн</span>
                              )}
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
