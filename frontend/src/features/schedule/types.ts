import { Lesson } from '../../types/schedule';

export interface PositionedLesson extends Lesson {
  column: number;
  totalColumns: number;
  startMinutes: number;
  durationMinutes: number;
  endMinutes: number;
}

// 24-часовая сетка временных слотов (с шагом 15 минут от 07:00 до 23:00)
export const SCHEDULE_TIME_OPTIONS: string[] = (() => {
  const options: string[] = [];
  for (let h = 7; h <= 23; h++) {
    for (let m = 0; m < 60; m += 15) {
      options.push(`${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`);
    }
  }
  return options;
})();
