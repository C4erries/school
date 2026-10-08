import React, { useEffect, useRef } from 'react';
import { StudyStreamItem } from '../../types/journal';
import { LessonReportFeedItem } from './LessonReportFeedItem';
import { NoteFeedItem } from './NoteFeedItem';
import { MessageSquareDashed } from 'lucide-react';

interface StudyStreamFeedProps {
  items: StudyStreamItem[];
  onDeleteNote: (noteId: string) => Promise<void>;
  isLoading: boolean;
}

export const StudyStreamFeed: React.FC<StudyStreamFeedProps> = ({
  items,
  onDeleteNote,
  isLoading,
}) => {
  const feedEndRef = useRef<HTMLDivElement>(null);

  // Автоскролл к низу ленты при поступлении новых сообщений
  useEffect(() => {
    if (!isLoading && items.length > 0) {
      feedEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [items.length, isLoading]);

  if (items.length === 0) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-8 text-center text-slate-500 space-y-3">
        <div className="w-14 h-14 rounded-3xl bg-indigo-50/60 border border-indigo-200/50 flex items-center justify-center text-indigo-500 shadow-sm">
          <MessageSquareDashed className="w-7 h-7" />
        </div>
        <div className="space-y-1">
          <h3 className="text-sm font-bold text-slate-800">
            История обучения пока пуста
          </h3>
          <p className="text-xs text-slate-500 max-w-xs">
            Здесь появятся отчеты по проведенным урокам и ваши быстрые заметки по ученику.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto space-y-3.5 p-4 sm:p-5 custom-scrollbar">
      {items.map((item) => {
        if (item.type === 'note' && item.note) {
          return (
            <NoteFeedItem
              key={`note-${item.id}`}
              note={item.note}
              onDeleteNote={onDeleteNote}
            />
          );
        }
        if (item.type === 'lesson_report' && item.lesson_report) {
          return (
            <LessonReportFeedItem
              key={`lesson-${item.id}`}
              bundle={item.lesson_report}
            />
          );
        }
        return null;
      })}
      <div ref={feedEndRef} className="h-1" />
    </div>
  );
};
