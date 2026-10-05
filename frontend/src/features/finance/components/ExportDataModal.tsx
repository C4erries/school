import React, { useState } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassButton } from '../../../shared/components/GlassButton';
import { GlassInput } from '../../../shared/components/GlassInput';
import { downloadExport } from '../../../api/finance';
import { ExportEntity } from '../../../types/finance';
import {
  Download,
  Users,
  Calendar,
  Receipt,
  FileSpreadsheet,
  CheckCircle2,
  AlertCircle,
} from 'lucide-react';

interface ExportDataModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const ExportDataModal: React.FC<ExportDataModalProps> = ({
  isOpen,
  onClose,
}) => {
  const [downloadingEntity, setDownloadingEntity] = useState<ExportEntity | null>(null);
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');
  const [successEntity, setSuccessEntity] = useState<ExportEntity | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const handleExport = async (entity: ExportEntity) => {
    setDownloadingEntity(entity);
    setSuccessEntity(null);
    setErrorMessage(null);
    try {
      const params: { from?: string; to?: string } = {};
      if (dateFrom) {
        params.from = new Date(dateFrom).toISOString();
      }
      if (dateTo) {
        // Конец выбранного дня
        const d = new Date(dateTo);
        d.setHours(23, 59, 59, 999);
        params.to = d.toISOString();
      }
      await downloadExport(entity, params);
      setSuccessEntity(entity);
      setTimeout(() => setSuccessEntity(null), 3500);
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Ошибка скачивания отчета');
    } finally {
      setDownloadingEntity(null);
    }
  };

  if (!isOpen) return null;

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title="Экспорт данных в Excel / CSV"
      description="Выгрузка данных в формате CSV с кодировкой UTF-8 BOM для корректного открытия в Excel на Windows и Mac"
      maxWidth="lg"
    >
      <div className="space-y-5">
        {errorMessage && (
          <div className="p-3.5 rounded-2xl bg-rose-500/10 border border-rose-500/20 flex items-center gap-2 text-xs font-semibold text-rose-700">
            <AlertCircle className="w-4 h-4 shrink-0 text-rose-600" />
            <span>{errorMessage}</span>
          </div>
        )}

        {/* Фильтр диапазона дат */}
        <div className="p-4 rounded-2xl bg-black/[0.03] border border-black/[0.05] space-y-2">
          <span className="block text-xs font-semibold uppercase tracking-wider text-slate-500">
            Период выгрузки (для уроков и оплат)
          </span>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <GlassInput
              type="date"
              label="С даты"
              value={dateFrom}
              onChange={(e) => setDateFrom(e.target.value)}
            />
            <GlassInput
              type="date"
              label="По дату"
              value={dateTo}
              onChange={(e) => setDateTo(e.target.value)}
            />
          </div>
          <p className="text-[11px] text-slate-400">
            Оставьте поля пустыми, чтобы выгрузить всю историю за все время
          </p>
        </div>

        {/* Варианты экспорта */}
        <div className="space-y-3">
          {/* Экспорт 1: База клиентов */}
          <div className="p-4 rounded-2xl bg-white/50 border border-black/[0.06] hover:bg-white/80 transition-all flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-2xl bg-indigo-500/15 text-indigo-600 flex items-center justify-center shrink-0 border border-indigo-400/20">
                <Users className="w-5 h-5" />
              </div>
              <div>
                <h4 className="font-bold text-sm text-slate-900">База учеников (Clients)</h4>
                <p className="text-xs text-slate-500">
                  Контакты, индивидуальные ставки по форматам, балансы часов, партнерские школы
                </p>
              </div>
            </div>

            <GlassButton
              variant={successEntity === 'clients' ? 'mint' : 'secondary'}
              size="sm"
              isLoading={downloadingEntity === 'clients'}
              onClick={() => handleExport('clients')}
              icon={
                successEntity === 'clients' ? (
                  <CheckCircle2 className="w-4 h-4 text-white" />
                ) : (
                  <Download className="w-4 h-4 text-indigo-600" />
                )
              }
              className="shrink-0 w-full sm:w-auto"
            >
              {successEntity === 'clients' ? 'Скачано' : 'Скачать CSV'}
            </GlassButton>
          </div>

          {/* Экспорт 2: Расписание уроков */}
          <div className="p-4 rounded-2xl bg-white/50 border border-black/[0.06] hover:bg-white/80 transition-all flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-2xl bg-amber-500/15 text-amber-600 flex items-center justify-center shrink-0 border border-amber-400/20">
                <Calendar className="w-5 h-5" />
              </div>
              <div>
                <h4 className="font-bold text-sm text-slate-900">Расписание уроков (Lessons)</h4>
                <p className="text-xs text-slate-500">
                  Время занятий, ученики, статусы, формат, кабинет или ссылка, темы и заметки
                </p>
              </div>
            </div>

            <GlassButton
              variant={successEntity === 'lessons' ? 'mint' : 'secondary'}
              size="sm"
              isLoading={downloadingEntity === 'lessons'}
              onClick={() => handleExport('lessons')}
              icon={
                successEntity === 'lessons' ? (
                  <CheckCircle2 className="w-4 h-4 text-white" />
                ) : (
                  <Download className="w-4 h-4 text-amber-600" />
                )
              }
              className="shrink-0 w-full sm:w-auto"
            >
              {successEntity === 'lessons' ? 'Скачано' : 'Скачать CSV'}
            </GlassButton>
          </div>

          {/* Экспорт 3: Журнал платежей */}
          <div className="p-4 rounded-2xl bg-white/50 border border-black/[0.06] hover:bg-white/80 transition-all flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-2xl bg-emerald-500/15 text-emerald-600 flex items-center justify-center shrink-0 border border-emerald-400/20">
                <Receipt className="w-5 h-5" />
              </div>
              <div>
                <h4 className="font-bold text-sm text-slate-900">Журнал платежей (Payments)</h4>
                <p className="text-xs text-slate-500">
                  Все поступившие оплаты, сумма в ₽, часы, формат, способ расчета (СБП/наличные), комментарии
                </p>
              </div>
            </div>

            <GlassButton
              variant={successEntity === 'payments' ? 'mint' : 'secondary'}
              size="sm"
              isLoading={downloadingEntity === 'payments'}
              onClick={() => handleExport('payments')}
              icon={
                successEntity === 'payments' ? (
                  <CheckCircle2 className="w-4 h-4 text-white" />
                ) : (
                  <Download className="w-4 h-4 text-emerald-600" />
                )
              }
              className="shrink-0 w-full sm:w-auto"
            >
              {successEntity === 'payments' ? 'Скачано' : 'Скачать CSV'}
            </GlassButton>
          </div>
        </div>

        {/* Подвал с подсказкой и закрытием */}
        <div className="pt-3 border-t border-black/[0.05] flex flex-col sm:flex-row items-center justify-between gap-3">
          <div className="flex items-center gap-1.5 text-xs text-slate-500">
            <FileSpreadsheet className="w-4 h-4 text-emerald-600 shrink-0" />
            <span>Файлы содержат BOM-метку и открываются в Excel без кракозябр</span>
          </div>

          <GlassButton type="button" variant="ghost" onClick={onClose}>
            Закрыть
          </GlassButton>
        </div>
      </div>
    </GlassModal>
  );
};
