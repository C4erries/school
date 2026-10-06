import React from 'react';
import { GlassInput } from '../../../shared/components/GlassInput';
import { Classroom, SubscriptionFormat } from '../../../types/schedule';
import { Trash2, AlertTriangle } from 'lucide-react';

interface EditLessonFormFieldsProps {
  editTitle: string;
  onChangeTitle: (val: string) => void;
  editDate: string;
  onChangeDate: (val: string) => void;
  editStartTime: string;
  onChangeStartTime: (val: string) => void;
  editEndTime: string;
  onChangeEndTime: (val: string) => void;
  allStartTimeOptions: string[];
  allEndTimeOptions: string[];
  editFormat: SubscriptionFormat;
  onChangeFormat: (fmt: SubscriptionFormat) => void;
  editLocationType: 'offline' | 'online';
  onChangeLocationType: (loc: 'offline' | 'online') => void;
  editClassroomId: string;
  onChangeClassroomId: (id: string) => void;
  editOnlineLink: string;
  onChangeOnlineLink: (link: string) => void;
  editComment: string;
  onChangeComment: (comment: string) => void;
  classrooms: Classroom[];
  isCancelled: boolean;
  lessonStatus: string;
  isCancelling: boolean;
  onSetIsCancelling: (val: boolean) => void;
  cancelReason: string;
  onChangeCancelReason: (reason: string) => void;
  onCancelClick: () => void;
  onNoShowClick: () => void;
}

export const EditLessonFormFields: React.FC<EditLessonFormFieldsProps> = ({
  editTitle,
  onChangeTitle,
  editDate,
  onChangeDate,
  editStartTime,
  onChangeStartTime,
  editEndTime,
  onChangeEndTime,
  allStartTimeOptions,
  allEndTimeOptions,
  editFormat,
  onChangeFormat,
  editLocationType,
  onChangeLocationType,
  editClassroomId,
  onChangeClassroomId,
  editOnlineLink,
  onChangeOnlineLink,
  editComment,
  onChangeComment,
  classrooms,
  isCancelled,
  lessonStatus,
  isCancelling,
  onSetIsCancelling,
  cancelReason,
  onChangeCancelReason,
  onCancelClick,
  onNoShowClick,
}) => {
  return (
    <div className="space-y-4">
      <GlassInput
        label="Тема / Название занятия"
        value={editTitle}
        onChange={(e) => onChangeTitle(e.target.value)}
        required
        disabled={isCancelled}
      />

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <GlassInput
          label="Дата"
          type="date"
          value={editDate}
          onChange={(e) => onChangeDate(e.target.value)}
          required
          disabled={isCancelled}
        />
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
            Начало *
          </label>
          <select
            value={editStartTime}
            onChange={(e) => onChangeStartTime(e.target.value)}
            disabled={isCancelled}
            className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 bg-white/70 border border-slate-200 focus:outline-none focus:border-indigo-500 disabled:opacity-50"
            required
          >
            {allStartTimeOptions.map((t) => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
            Конец *
          </label>
          <select
            value={editEndTime}
            onChange={(e) => onChangeEndTime(e.target.value)}
            disabled={isCancelled}
            className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 bg-white/70 border border-slate-200 focus:outline-none focus:border-indigo-500 disabled:opacity-50"
            required
          >
            {allEndTimeOptions.map((t) => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Формат занятия */}
      <div>
        <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
          Формат занятия
        </label>
        <div className="grid grid-cols-3 gap-2 p-1 rounded-2xl bg-black/[0.04]">
          <button
            type="button"
            disabled={isCancelled}
            onClick={() => onChangeFormat('individual')}
            className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
              editFormat === 'individual'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            Индивидуально
          </button>
          <button
            type="button"
            disabled={isCancelled}
            onClick={() => onChangeFormat('pair')}
            className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
              editFormat === 'pair'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            В паре
          </button>
          <button
            type="button"
            disabled={isCancelled}
            onClick={() => onChangeFormat('group')}
            className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
              editFormat === 'group'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            В группе
          </button>
        </div>
      </div>

      {/* Локация проведения */}
      <div>
        <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
          Локация проведения
        </label>
        <div className="grid grid-cols-2 gap-2 p-1 rounded-2xl bg-black/[0.04]">
          <button
            type="button"
            disabled={isCancelled}
            onClick={() => onChangeLocationType('offline')}
            className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
              editLocationType === 'offline'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            Оффлайн в кабинете
          </button>
          <button
            type="button"
            disabled={isCancelled}
            onClick={() => onChangeLocationType('online')}
            className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
              editLocationType === 'online'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            Онлайн урок
          </button>
        </div>
      </div>

      {editLocationType === 'offline' ? (
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">
            Кабинет
          </label>
          <select
            disabled={isCancelled}
            value={editClassroomId}
            onChange={(e) => onChangeClassroomId(e.target.value)}
            className="w-full rounded-2xl px-4 py-3 text-sm text-slate-800 bg-white/70 border border-slate-200 focus:outline-none focus:border-indigo-500 disabled:opacity-50"
          >
            <option value="">Без закрепления кабинета</option>
            {classrooms.map((room) => (
              <option key={room.id} value={room.id}>
                {room.name}
              </option>
            ))}
          </select>
        </div>
      ) : (
        <GlassInput
          label="Ссылка на созвон"
          placeholder="https://telemost.yandex.ru/j/..."
          value={editOnlineLink}
          onChange={(e) => onChangeOnlineLink(e.target.value)}
          disabled={isCancelled}
        />
      )}

      <GlassInput
        label="Заметка / ДЗ"
        value={editComment}
        onChange={(e) => onChangeComment(e.target.value)}
        disabled={isCancelled}
      />

      {/* Блок отмены урока */}
      {!isCancelled && lessonStatus === 'scheduled' && (
        <div className="p-3.5 rounded-2xl bg-rose-500/10 border border-rose-500/20 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-rose-800 flex items-center gap-1.5">
              <Trash2 className="w-3.5 h-3.5" /> Отмена занятия
            </span>
            {!isCancelling && (
              <div className="flex items-center gap-3">
                <button
                  type="button"
                  onClick={onNoShowClick}
                  className="text-xs font-semibold text-amber-700 hover:text-amber-800 flex items-center gap-1"
                >
                  <AlertTriangle className="w-3.5 h-3.5" />
                  Неявка
                </button>
                <button
                  type="button"
                  onClick={() => onSetIsCancelling(true)}
                  className="text-xs font-semibold text-rose-600 hover:text-rose-700 underline"
                >
                  Отменить занятие
                </button>
              </div>
            )}
          </div>

          {isCancelling && (
            <div className="space-y-2 pt-1 animate-in fade-in">
              <input
                type="text"
                placeholder="Укажите причину отмены (необязательно)"
                value={cancelReason}
                onChange={(e) => onChangeCancelReason(e.target.value)}
                className="w-full rounded-xl px-3 py-2 text-xs text-slate-800 bg-white/80 border border-rose-200 focus:outline-none"
              />
              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => onSetIsCancelling(false)}
                  className="px-3 py-1 rounded-lg text-xs font-medium text-slate-600 hover:bg-black/5"
                >
                  Назад
                </button>
                <button
                  type="button"
                  onClick={onCancelClick}
                  className="px-3 py-1 rounded-lg text-xs font-bold text-white bg-rose-500 hover:bg-rose-600 shadow-sm"
                >
                  Подтвердить отмену
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

