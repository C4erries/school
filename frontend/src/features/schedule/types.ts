import { Lesson } from '../../types/schedule';

export interface PositionedLesson extends Lesson {
  column: number;
  totalColumns: number;
  startMinutes: number;
  durationMinutes: number;
  endMinutes: number;
}
