import React from 'react';
import { GlassButton } from '../../../shared/components/GlassButton';
import { GlassInput } from '../../../shared/components/GlassInput';
import { Badge } from '../../../shared/components/Badge';
import { CalendarSettings } from '../../../types/calendar';
import {
  Calendar,
  Copy,
  Check,
  Download,
  ExternalLink,
  Smartphone,
  AlertTriangle,
  RotateCcw,
} from 'lucide-react';

interface CalendarExportTabProps {
  settings: CalendarSettings | null;
  copied: boolean;
  onCopyLink: () => void;
  isDownloading: boolean;
  onDownload: () => void;
  showRotateConfirm: boolean;
  onSetShowRotateConfirm: (val: boolean) => void;
  isRotating: boolean;
  onRotate: () => void;
}

export const CalendarExportTab: React.FC<CalendarExportTabProps> = ({
  settings,
  copied,
  onCopyLink,
  isDownloading,
  onDownload,
  showRotateConfirm,
  onSetShowRotateConfirm,
  isRotating,
  onRotate,
}) => {
  const googleCalendarUrl = settings?.webcal_url
    ? `https://calendar.google.com/calendar/render?cid=${encodeURIComponent(settings.webcal_url)}`
    : '#';

  return (
    <div className="space-y-5">
      {/* 1-Click кнопки добавления */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <a
          href={googleCalendarUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center justify-center gap-2.5 p-3.5 rounded-2xl bg-indigo-600/10 hover:bg-indigo-600/20 border border-indigo-500/30 text-indigo-700 text-sm font-semibold transition-all hover:scale-[1.01] active:scale-[0.99] text-center"
        >
          <Calendar className="w-4 h-4 text-indigo-600 shrink-0" />
          <span>В Google Календарь</span>
          <ExternalLink className="w-3.5 h-3.5 ml-auto text-indigo-400" />
        </a>

        <a
          href={settings?.webcal_url || '#'}
          className="flex items-center justify-center gap-2.5 p-3.5 rounded-2xl bg-slate-900/10 hover:bg-slate-900/15 border border-slate-700/20 text-slate-900 text-sm font-semibold transition-all hover:scale-[1.01] active:scale-[0.99] text-center"
        >
          <Smartphone className="w-4 h-4 text-slate-800 shrink-0" />
          <span>В Apple Календарь</span>
          <ExternalLink className="w-3.5 h-3.5 ml-auto text-slate-500" />
        </a>
      </div>

      {/* Персональная ссылка */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
            Персональная ссылка подписки (iCal / Webcal)
          </span>
          {copied && (
            <Badge variant="mint" className="text-[11px] py-0.5">
              Скопировано в буфер
            </Badge>
          )}
        </div>
        <div className="flex gap-2">
          <div className="flex-1">
            <GlassInput
              readOnly
              value={settings?.webcal_url || ''}
              className="font-mono text-xs select-all bg-white/40"
              onFocus={(e) => e.target.select()}
            />
          </div>
          <GlassButton
            variant="primary"
            onClick={onCopyLink}
            icon={copied ? <Check className="w-4 h-4 text-white" /> : <Copy className="w-4 h-4" />}
            className="shrink-0"
          >
            {copied ? 'Готово' : 'Скопировать'}
          </GlassButton>
        </div>
        <p className="text-[11px] text-slate-500 leading-normal">
          Ссылка обновляется в реальном времени. Все изменения и переносы уроков автоматически отобразятся в вашем календаре.
        </p>
      </div>

      {/* Скачивание и перевыпуск */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 pt-2 border-t border-black/[0.06]">
        <GlassButton
          variant="secondary"
          size="sm"
          onClick={onDownload}
          isLoading={isDownloading}
          icon={<Download className="w-4 h-4 text-slate-600" />}
        >
          Скачать .ics файл
        </GlassButton>

        {!showRotateConfirm ? (
          <GlassButton
            variant="ghost"
            size="sm"
            onClick={() => onSetShowRotateConfirm(true)}
            icon={<RotateCcw className="w-3.5 h-3.5 text-slate-500" />}
            className="text-slate-500 hover:text-slate-800"
          >
            Сменить секретную ссылку
          </GlassButton>
        ) : (
          <div className="flex items-center gap-2 p-2 rounded-xl bg-amber-500/10 border border-amber-500/20 text-xs">
            <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0" />
            <span className="text-amber-900 font-medium text-[11px]">Старая ссылка перестанет работать!</span>
            <button
              type="button"
              onClick={onRotate}
              disabled={isRotating}
              className="px-2 py-1 bg-amber-600 text-white rounded-lg font-semibold hover:bg-amber-700 disabled:opacity-50"
            >
              {isRotating ? '...' : 'Подтвердить'}
            </button>
            <button
              type="button"
              onClick={() => onSetShowRotateConfirm(false)}
              className="text-slate-500 hover:text-slate-800"
            >
              Отмена
            </button>
          </div>
        )}
      </div>

      {/* Памятка по настройке */}
      <div className="p-4 rounded-2xl bg-black/[0.02] border border-black/[0.04] space-y-2.5">
        <span className="text-xs font-bold text-slate-800 flex items-center gap-1.5">
          <Smartphone className="w-3.5 h-3.5 text-indigo-500" />
          Инструкция по настройке
        </span>
        <ul className="text-xs text-slate-600 space-y-1.5 list-disc list-inside leading-relaxed">
          <li>
            <strong className="text-slate-700">iPhone / iPad / Mac:</strong> Нажмите «В Apple Календарь» или откройте{' '}
            <span className="italic">Настройки → Календарь → Учетные записи → Добавить учетную запись → Подписные календари</span> и вставьте ссылку.
          </li>
          <li>
            <strong className="text-slate-700">Google Календарь / Android:</strong> Нажмите «В Google Календарь» или в веб-версии Google Календаря нажмите{' '}
            <span className="italic">«Другие календари» (+) → «С URL-адреса»</span> и укажите ссылку.
          </li>
          <li>
            <strong className="text-slate-700">Напоминания:</strong> Смартфон будет присылать push-уведомления за 15–30 минут до начала каждого урока.
          </li>
        </ul>
      </div>
    </div>
  );
};

