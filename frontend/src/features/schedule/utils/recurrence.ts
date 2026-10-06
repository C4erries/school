export const WEEKDAYS = [
  { key: 'MO', label: 'Пн' },
  { key: 'TU', label: 'Вт' },
  { key: 'WE', label: 'Ср' },
  { key: 'TH', label: 'Чт' },
  { key: 'FR', label: 'Пт' },
  { key: 'SA', label: 'Сб' },
  { key: 'SU', label: 'Вс' },
];

export function getRRuleDayFromDate(dateStr: string): string {
  if (!dateStr) return 'MO';
  const d = new Date(dateStr);
  const day = d.getDay(); // 0 is SU, 1 is MO, ..., 6 is SA
  const map = ['SU', 'MO', 'TU', 'WE', 'TH', 'FR', 'SA'];
  return map[day] || 'MO';
}

