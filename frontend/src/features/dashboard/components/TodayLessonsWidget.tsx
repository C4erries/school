import React from 'react';
import { TodayLesson } from '../../../types/dashboard';
import { GlassCard } from '../../../shared/components/GlassCard';
import { GlassButton } from '../../../shared/components/GlassButton';
import { Calendar, Video, MapPin, CheckCircle2, ArrowRight } from 'lucide-react';

interface TodayLessonsWidgetProps {
  lessons: TodayLesson[];
  isLoading?: boolean;
  onNavigateToSchedule: () => void;
}

export const TodayLessonsWidget: React.FC<TodayLessonsWidgetProps> = ({
  lessons,
  isLoading = false,
  onNavigateToSchedule,
}) => {
  const getFormatBadge = (format: string) => {
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
            Индивидуально
          </span>
        );
    }
  };

  return (
    <GlassCard className="p-5 sm:p-6 space-y-4">
      <div className="flex items-center justify-between border-b border-black/[0.05] pb-3">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-2xl bg-indigo-500/15 text-indigo-600 flex items-center justify-center border border-indigo-400/20">
            <Calendar className="w-5 h-5" />
          </div>
          <div>
            <h2 className="font-bold text-lg text-slate-900">Уроки на сегодня</h2>
            <p className="text-xs text-slate-500">
              {lessons.length > 0
                ? `Запланировано занятий: ${lessons.length}`
                : 'На сегодня занятий нет'}
            </p>
          </div>
        </div>

        <GlassButton
          variant="secondary"
          size="sm"
          onClick={onNavigateToSchedule}
          icon={<ArrowRight className="w-3.5 h-3.5" />}
        >
          В календарь
        </GlassButton>
      </div>

      {isLoading ? (
        <div className="p-8 text-center text-slate-400">
          <div className="w-6 h-6 border-2 border-indigo-600 border-t-transparent rounded-full animate-spin mx-auto mb-2" />
          <span>Загрузка расписания на сегодня...</span>
        </div>
      ) : lessons.length === 0 ? (
        <div className="p-8 text-center text-slate-400 space-y-2">
          <CheckCircle2 className="w-8 h-8 text-emerald-400 mx-auto" />
          <p className="text-sm font-semibold text-slate-700">Все уроки завершены или свободный день!</p>
          <p className="text-xs text-slate-400">
            Вы можете запланировать новое занятие в расписании или проверить балансы учеников
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {lessons.map((l) => {
            const startTime = new Date(l.start_at).toLocaleTimeString('ru-RU', {
              hour: '2-digit',
              minute: '2-digit',
            });
            const endTime = new Date(l.end_at).toLocaleTimeString('ru-RU', {
              hour: '2-digit',
              minute: '2-digit',
            });

            return (
              <div
                key={l.id}
                className="p-3.5 sm:p-4 rounded-2xl bg-black/[0.02] border border-black/[0.04] flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-black/[0.04] transition-colors"
                style={{
                  borderLeftWidth: '4px',
                  borderLeftColor: l.classroom_color || '#6366F1',
                }}
              >
                <div className="flex items-start sm:items-center gap-3">
                  <div className="px-2.5 py-1.5 rounded-xl bg-white text-indigo-700 font-mono font-bold text-xs border border-indigo-100 shadow-xs shrink-0">
                    {startTime} – {endTime}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-bold text-sm text-slate-900">{l.client_name}</span>
                      {getFormatBadge(l.format)}
                    </div>
                    <div className="flex items-center gap-3 text-xs text-slate-500 mt-0.5">
                      {l.classroom_name ? (
                        <span className="flex items-center gap-1">
                          <MapPin className="w-3 h-3 text-slate-400" />
                          {l.classroom_name}
                        </span>
                      ) : l.online_link || l.location_type === 'online' ? (
                        <span className="flex items-center gap-1 text-emerald-600 font-medium">
                          <Video className="w-3 h-3" />
                          Онлайн
                        </span>
                      ) : null}
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2 self-end sm:self-auto">
                  {l.online_link && l.online_link.startsWith('http') && (
                    <a
                      href={l.online_link}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="px-3 py-1.5 rounded-xl bg-emerald-500 hover:bg-emerald-600 text-white font-medium text-xs flex items-center gap-1.5 shadow-sm transition-all"
                    >
                      <Video className="w-3.5 h-3.5" />
                      <span>Войти</span>
                    </a>
                  )}
                  <GlassButton
                    variant="secondary"
                    size="sm"
                    onClick={onNavigateToSchedule}
                  >
                    К уроку
                  </GlassButton>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </GlassCard>
  );
};
