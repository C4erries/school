import React from 'react';
import { GlassCard } from '../../../shared/components/GlassCard';
import { TagStat } from '../../../types/analytics';
import { Tags, Users, Clock, Wallet } from 'lucide-react';

interface AnalyticsTagsDistributionProps {
  tags: TagStat[];
  isLoading?: boolean;
}

export const AnalyticsTagsDistribution: React.FC<AnalyticsTagsDistributionProps> = ({
  tags,
  isLoading = false,
}) => {
  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      maximumFractionDigits: 0,
    }).format(Math.round(val));
  };

  const formatHours = (val: number) => {
    return Number.isInteger(val) ? String(val) : val.toFixed(1);
  };

  const totalGross = tags.reduce((acc, t) => acc + (t.gross_revenue || 0), 0);
  const totalNet = tags.reduce((acc, t) => acc + (t.net_income || 0), 0);
  const totalHours = tags.reduce((acc, t) => acc + (t.completed_hours || 0), 0);

  return (
    <GlassCard className="p-5 sm:p-6 space-y-4">
      {/* Заголовок */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 border-b border-black/[0.05] pb-3">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-2xl bg-indigo-500/15 text-indigo-600 flex items-center justify-center border border-indigo-400/20">
            <Tags className="w-5 h-5" />
          </div>
          <div>
            <h3 className="font-bold text-lg text-slate-900">Доходность и нагрузка по тегам</h3>
            <p className="text-xs text-slate-500">
              Группировка учеников по категориям (ОГЭ, ЕГЭ, партнерские школы и др.)
            </p>
          </div>
        </div>

        {tags.length > 0 && (
          <div className="flex items-center gap-3 text-xs text-slate-500">
            <span>
              Всего по тегам:{' '}
              <strong className="text-slate-800 font-semibold">{formatHours(totalHours)} ч</strong>
            </span>
            <span className="text-slate-300">|</span>
            <span>
              Чистый доход:{' '}
              <strong className="text-emerald-600 font-semibold">{formatCurrency(totalNet)}</strong>
            </span>
          </div>
        )}
      </div>

      {/* Содержимое */}
      {isLoading ? (
        <div className="p-8 text-center text-slate-400">
          <div className="w-6 h-6 border-2 border-indigo-600 border-t-transparent rounded-full animate-spin mx-auto mb-2" />
          <span>Загрузка статистики по тегам...</span>
        </div>
      ) : tags.length === 0 ? (
        <div className="p-8 text-center text-slate-400 space-y-1.5">
          <Tags className="w-8 h-8 text-slate-300 mx-auto" />
          <p className="text-sm font-semibold text-slate-700">Нет данных по тегам за выбранный период</p>
          <p className="text-xs text-slate-400">
            Привяжите теги к ученикам в CRM, чтобы отслеживать категории и направления обучения
          </p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs border-collapse">
            <thead>
              <tr className="border-b border-black/[0.05] text-slate-400 uppercase font-semibold text-[10px] tracking-wider">
                <th className="py-2.5 px-3">Категория (Тег)</th>
                <th className="py-2.5 px-3 text-center">Учеников</th>
                <th className="py-2.5 px-3 text-center">Отработано часов</th>
                <th className="py-2.5 px-3 text-right">Выручка (Gross)</th>
                <th className="py-2.5 px-3 text-right">Чистый доход</th>
                <th className="py-2.5 px-3 text-right">Доля в доходе</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-black/[0.03]">
              {tags.map((tag) => {
                const sharePercent = totalNet > 0 ? ((tag.net_income || 0) / totalNet) * 100 : 0;

                return (
                  <tr key={tag.tag_id} className="hover:bg-black/[0.02] transition-colors">
                    <td className="py-3 px-3">
                      <div className="flex items-center gap-2">
                        <span
                          className="w-2.5 h-2.5 rounded-full shrink-0"
                          style={{ backgroundColor: tag.tag_color || '#6366F1' }}
                        />
                        <span className="font-bold text-slate-800 text-sm">{tag.tag_name}</span>
                      </div>
                    </td>
                    <td className="py-3 px-3 text-center">
                      <span className="inline-flex items-center gap-1 font-semibold text-slate-700">
                        <Users className="w-3 h-3 text-slate-400" />
                        {tag.students_count}
                      </span>
                    </td>
                    <td className="py-3 px-3 text-center">
                      <span className="inline-flex items-center gap-1 font-semibold text-slate-700">
                        <Clock className="w-3 h-3 text-slate-400" />
                        {formatHours(tag.completed_hours)} ч
                      </span>
                    </td>
                    <td className="py-3 px-3 text-right font-medium text-slate-600">
                      {formatCurrency(tag.gross_revenue)}
                    </td>
                    <td className="py-3 px-3 text-right font-bold text-emerald-600">
                      <span className="inline-flex items-center gap-1 justify-end">
                        <Wallet className="w-3 h-3 text-emerald-500" />
                        {formatCurrency(tag.net_income)}
                      </span>
                    </td>
                    <td className="py-3 px-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <div className="w-16 h-1.5 rounded-full bg-black/[0.06] overflow-hidden">
                          <div
                            className="h-full rounded-full bg-indigo-500"
                            style={{ width: `${Math.min(100, sharePercent)}%` }}
                          />
                        </div>
                        <span className="font-mono font-semibold text-slate-600 w-9 text-right">
                          {sharePercent.toFixed(0)}%
                        </span>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
            {tags.length > 1 && (
              <tfoot>
                <tr className="border-t-2 border-black/[0.08] font-bold text-slate-800">
                  <td className="py-3 px-3">Итого по тегам</td>
                  <td className="py-3 px-3 text-center">
                    {tags.reduce((acc, t) => acc + (t.students_count || 0), 0)}
                  </td>
                  <td className="py-3 px-3 text-center">{formatHours(totalHours)} ч</td>
                  <td className="py-3 px-3 text-right">{formatCurrency(totalGross)}</td>
                  <td className="py-3 px-3 text-right text-emerald-700">{formatCurrency(totalNet)}</td>
                  <td className="py-3 px-3 text-right font-mono">100%</td>
                </tr>
              </tfoot>
            )}
          </table>
        </div>
      )}
    </GlassCard>
  );
};
