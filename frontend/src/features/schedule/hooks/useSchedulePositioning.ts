import { useMemo, useCallback } from 'react';
import { Lesson } from '../../../types/schedule';
import { PositionedLesson } from '../types';

export const START_HOUR = 8;
export const END_HOUR = 22;
export const HOUR_HEIGHT = 64; // px per hour slot

export function useSchedulePositioning() {
  const hours = useMemo(() => {
    const list: number[] = [];
    for (let h = START_HOUR; h <= END_HOUR; h++) {
      list.push(h);
    }
    return list;
  }, []);

  const positionLessons = useCallback((dayItems: Lesson[]): PositionedLesson[] => {
    if (dayItems.length === 0) return [];

    const sorted = [...dayItems].sort(
      (a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
    );

    const clusters: PositionedLesson[][] = [];
    let currentCluster: PositionedLesson[] = [];
    let clusterEnd = 0;

    for (const lesson of sorted) {
      const start = new Date(lesson.start_time);
      const end = new Date(lesson.end_time);
      const startMinutes = start.getHours() * 60 + start.getMinutes();
      const endMinutes = end.getHours() * 60 + end.getMinutes();
      const durationMinutes = Math.max(30, endMinutes - startMinutes);

      const item: PositionedLesson = {
        ...lesson,
        column: 0,
        totalColumns: 1,
        startMinutes,
        durationMinutes,
        endMinutes: startMinutes + durationMinutes,
      };

      if (currentCluster.length === 0) {
        currentCluster.push(item);
        clusterEnd = item.endMinutes;
      } else {
        if (startMinutes < clusterEnd) {
          currentCluster.push(item);
          clusterEnd = Math.max(clusterEnd, item.endMinutes);
        } else {
          clusters.push(currentCluster);
          currentCluster = [item];
          clusterEnd = item.endMinutes;
        }
      }
    }
    if (currentCluster.length > 0) {
      clusters.push(currentCluster);
    }

    const result: PositionedLesson[] = [];
    for (const cluster of clusters) {
      const columns: number[] = [];

      for (const item of cluster) {
        let placedCol = -1;
        for (let col = 0; col < columns.length; col++) {
          if (columns[col] <= item.startMinutes) {
            placedCol = col;
            columns[col] = item.endMinutes;
            break;
          }
        }
        if (placedCol === -1) {
          placedCol = columns.length;
          columns.push(item.endMinutes);
        }
        item.column = placedCol;
      }

      const totalCols = Math.max(1, columns.length);
      for (const item of cluster) {
        item.totalColumns = totalCols;
        result.push(item);
      }
    }

    return result;
  }, []);

  return {
    hours,
    positionLessons,
    START_HOUR,
    END_HOUR,
    HOUR_HEIGHT,
  };
}
