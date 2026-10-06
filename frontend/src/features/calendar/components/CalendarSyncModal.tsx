import React, { useState, useEffect, useCallback } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { CalendarSettings } from '../../../types/calendar';
import {
  getCalendarSettings,
  rotateCalendarToken,
  downloadCalendarExport,
} from '../../../api/calendar';
import { RefreshCw, Calendar, UploadCloud } from 'lucide-react';
import { CalendarExportTab } from './CalendarExportTab';
import { CalendarImportTab } from './CalendarImportTab';

interface CalendarSyncModalProps {
  isOpen: boolean;
  onClose: () => void;
  onImportSuccess?: () => void;
}

export const CalendarSyncModal: React.FC<CalendarSyncModalProps> = ({
  isOpen,
  onClose,
  onImportSuccess,
}) => {
  const [activeTab, setActiveTab] = useState<'sync' | 'import'>('sync');
  const [settings, setSettings] = useState<CalendarSettings | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isRotating, setIsRotating] = useState(false);
  const [isDownloading, setIsDownloading] = useState(false);
  const [copied, setCopied] = useState(false);
  const [showRotateConfirm, setShowRotateConfirm] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const loadSettings = useCallback(async () => {
    setIsLoading(true);
    setErrorMessage(null);
    try {
      const data = await getCalendarSettings();
      setSettings(data);
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Ошибка загрузки настроек');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (isOpen) {
      loadSettings();
      setShowRotateConfirm(false);
      setCopied(false);
    }
  }, [isOpen, loadSettings]);

  const handleCopyLink = async () => {
    if (!settings?.webcal_url) return;
    try {
      await navigator.clipboard.writeText(settings.webcal_url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2500);
    } catch {
      // fallback
    }
  };

  const handleRotate = async () => {
    setIsRotating(true);
    setErrorMessage(null);
    try {
      const updated = await rotateCalendarToken();
      setSettings(updated);
      setShowRotateConfirm(false);
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Ошибка ротации токена');
    } finally {
      setIsRotating(false);
    }
  };

  const handleDownload = async () => {
    setIsDownloading(true);
    setErrorMessage(null);
    try {
      await downloadCalendarExport();
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Ошибка скачивания файла');
    } finally {
      setIsDownloading(false);
    }
  };

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title="Синхронизация и импорт расписания"
      description="Подписка на живой iCal-фид или импорт расписания из Google Календаря"
      maxWidth="lg"
    >
      <div className="space-y-5">
        {/* Вкладки переключения: Синхронизация / Импорт */}
        <div className="grid grid-cols-2 gap-2 p-1 rounded-2xl bg-black/[0.04]">
          <button
            type="button"
            onClick={() => setActiveTab('sync')}
            className={`flex items-center justify-center gap-2 py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
              activeTab === 'sync'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <Calendar className="w-3.5 h-3.5" />
            <span>Синхронизация (Экспорт)</span>
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('import')}
            className={`flex items-center justify-center gap-2 py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
              activeTab === 'import'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <UploadCloud className="w-3.5 h-3.5" />
            <span>Импорт из Google / .ics</span>
          </button>
        </div>

        {errorMessage && activeTab === 'sync' && (
          <div className="p-3.5 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-800 text-xs flex items-center justify-between gap-2">
            <span>{errorMessage}</span>
            <button
              type="button"
              onClick={loadSettings}
              className="text-rose-700 font-semibold underline hover:no-underline"
            >
              Повторить
            </button>
          </div>
        )}

        {activeTab === 'sync' ? (
          isLoading ? (
            <div className="py-12 flex flex-col items-center justify-center gap-3 text-slate-500">
              <RefreshCw className="w-6 h-6 animate-spin text-indigo-500" />
              <p className="text-sm">Загрузка персональной ссылки...</p>
            </div>
          ) : (
            <CalendarExportTab
              settings={settings}
              copied={copied}
              onCopyLink={handleCopyLink}
              isDownloading={isDownloading}
              onDownload={handleDownload}
              showRotateConfirm={showRotateConfirm}
              onSetShowRotateConfirm={setShowRotateConfirm}
              isRotating={isRotating}
              onRotate={handleRotate}
            />
          )
        ) : (
          <CalendarImportTab onImportSuccess={onImportSuccess} />
        )}
      </div>
    </GlassModal>
  );
};
