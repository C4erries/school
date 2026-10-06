import React, { useState, useRef } from 'react';
import { GlassButton } from '../../../shared/components/GlassButton';
import { CalendarImportResponse } from '../../../types/calendar';
import { importCalendar } from '../../../api/calendar';
import { UploadCloud, CheckCircle2, AlertCircle, FileText } from 'lucide-react';

interface CalendarImportTabProps {
  onImportSuccess?: () => void;
}

export const CalendarImportTab: React.FC<CalendarImportTabProps> = ({ onImportSuccess }) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [result, setResult] = useState<CalendarImportResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      if (!file.name.endsWith('.ics')) {
        setError('Пожалуйста, выберите файл с расширением .ics');
        setSelectedFile(null);
        return;
      }
      setSelectedFile(file);
      setError(null);
      setResult(null);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const file = e.dataTransfer.files[0];
      if (!file.name.endsWith('.ics')) {
        setError('Поддерживаются только файлы формата iCalendar (.ics)');
        return;
      }
      setSelectedFile(file);
      setError(null);
      setResult(null);
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) return;

    setIsUploading(true);
    setError(null);
    setResult(null);

    try {
      const res = await importCalendar(selectedFile);
      setResult(res);
      setSelectedFile(null);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
      if (onImportSuccess) {
        onImportSuccess();
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Не удалось импортировать файл календаря');
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div className="space-y-4">
      {/* Drag & Drop Zone */}
      <div
        onDragOver={(e) => {
          e.preventDefault();
          setIsDragging(true);
        }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={handleDrop}
        onClick={() => fileInputRef.current?.click()}
        className={`p-6 sm:p-8 rounded-3xl border-2 border-dashed transition-all cursor-pointer flex flex-col items-center justify-center text-center ${
          isDragging
            ? 'border-indigo-500 bg-indigo-50/40'
            : selectedFile
            ? 'border-emerald-400 bg-emerald-50/20'
            : 'border-slate-300/80 bg-white/40 hover:bg-white/60'
        }`}
      >
        <input
          ref={fileInputRef}
          type="file"
          accept=".ics"
          onChange={handleFileChange}
          className="hidden"
        />

        {selectedFile ? (
          <div className="space-y-2">
            <div className="w-12 h-12 rounded-2xl bg-emerald-500/10 text-emerald-600 flex items-center justify-center mx-auto">
              <FileText className="w-6 h-6" />
            </div>
            <div>
              <p className="text-sm font-bold text-slate-800">{selectedFile.name}</p>
              <p className="text-xs text-slate-500">
                {(selectedFile.size / 1024).toFixed(1)} КБ • Нажмите «Начать импорт»
              </p>
            </div>
          </div>
        ) : (
          <div className="space-y-2">
            <div className="w-12 h-12 rounded-2xl bg-indigo-500/10 text-indigo-600 flex items-center justify-center mx-auto">
              <UploadCloud className="w-6 h-6" />
            </div>
            <div>
              <p className="text-sm font-semibold text-slate-800">
                Перетащите .ics файл сюда или <span className="text-indigo-600 underline">выберите на диске</span>
              </p>
              <p className="text-xs text-slate-500 mt-0.5">
                Экспортируйте файл календаря из Google Календаря или Яндекс Календаря
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Button */}
      {selectedFile && (
        <div className="flex justify-end">
          <GlassButton
            variant="primary"
            onClick={(e) => {
              e.stopPropagation();
              handleUpload();
            }}
            isLoading={isUploading}
            icon={<UploadCloud className="w-4 h-4" />}
          >
            Начать импорт уроков
          </GlassButton>
        </div>
      )}

      {/* Error display */}
      {error && (
        <div className="p-3.5 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-800 text-xs flex items-center gap-2">
          <AlertCircle className="w-4 h-4 text-rose-600 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Success notification */}
      {result && (
        <div className="p-4 rounded-2xl bg-emerald-500/15 border border-emerald-500/30 text-emerald-950 text-xs space-y-2 animate-in fade-in">
          <div className="flex items-center gap-2 font-bold text-sm text-emerald-900">
            <CheckCircle2 className="w-5 h-5 text-emerald-600 shrink-0" />
            <span>Календарь успешно импортирован!</span>
          </div>
          <div className="grid grid-cols-3 gap-2 pt-1 font-medium">
            <div className="p-2.5 rounded-xl bg-white/60">
              <div className="text-[10px] text-slate-500 uppercase">Разовых уроков</div>
              <div className="text-base font-bold text-emerald-700">{result.imported_lessons}</div>
            </div>
            <div className="p-2.5 rounded-xl bg-white/60">
              <div className="text-[10px] text-slate-500 uppercase">Серий занятий</div>
              <div className="text-base font-bold text-indigo-700">{result.imported_series}</div>
            </div>
            <div className="p-2.5 rounded-xl bg-white/60">
              <div className="text-[10px] text-slate-500 uppercase">Пропущено</div>
              <div className="text-base font-bold text-slate-600">{result.skipped_events}</div>
            </div>
          </div>
          {result.message && (
            <p className="text-[11px] text-emerald-800/80 mt-1">{result.message}</p>
          )}
        </div>
      )}

      {/* Help block */}
      <div className="p-3.5 rounded-2xl bg-black/[0.02] border border-black/[0.04] text-[11px] text-slate-500 space-y-1">
        <p className="font-semibold text-slate-700">Как экспортировать .ics из Google Календаря:</p>
        <ol className="list-decimal list-inside space-y-0.5">
          <li>Откройте Google Календарь на компьютере</li>
          <li>В верхнем правом углу нажмите значок шестеренки → Настройки</li>
          <li>В левом меню выберите «Импорт и экспорт» → «Экспорт»</li>
          <li>Распакуйте полученный архив и выберите нужный файл .ics</li>
        </ol>
      </div>
    </div>
  );
};

