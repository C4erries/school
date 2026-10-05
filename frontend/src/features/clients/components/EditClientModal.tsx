import React, { useState, useEffect } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassInput } from '../../../shared/components/GlassInput';
import { GlassButton } from '../../../shared/components/GlassButton';
import { Client, Tag } from '../../../types/schedule';
import { updateClient, createTag } from '../../../api/schedule';
import { Plus, Tag as TagIcon, Check, Percent } from 'lucide-react';

interface EditClientModalProps {
  isOpen: boolean;
  onClose: () => void;
  client: Client | null;
  tags: Tag[];
  onUpdated: () => Promise<void>;
  onTagCreated: (tag: Tag) => void;
}

export const EditClientModal: React.FC<EditClientModalProps> = ({
  isOpen,
  onClose,
  client,
  tags,
  onUpdated,
  onTagCreated,
}) => {
  const [editName, setEditName] = useState('');
  const [editPhone, setEditPhone] = useState('');
  const [editRateIndividual, setEditRateIndividual] = useState('');
  const [editRatePair, setEditRatePair] = useState('');
  const [editRateGroup, setEditRateGroup] = useState('');
  const [editSelectedTagIds, setEditSelectedTagIds] = useState<string[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Форма создания нового тега на лету
  const [isTagFormOpen, setIsTagFormOpen] = useState(false);
  const [newTagName, setNewTagName] = useState('');
  const [newTagPercent, setNewTagPercent] = useState('0');
  const [isCreatingTag, setIsCreatingTag] = useState(false);

  useEffect(() => {
    if (client && isOpen) {
      setEditName(client.name);
      setEditPhone(client.phone || '');
      setEditRateIndividual(String(client.rate_individual ?? client.base_rate ?? ''));
      setEditRatePair(client.rate_pair != null ? String(client.rate_pair) : '');
      setEditRateGroup(client.rate_group != null ? String(client.rate_group) : '');
      const clientTagIds =
        client.tags && client.tags.length > 0
          ? client.tags.map((t) => t.id)
          : client.tag_ids || [];
      setEditSelectedTagIds(clientTagIds);
      setIsTagFormOpen(false);
    }
  }, [client, isOpen]);

  const toggleTagSelection = (tagId: string) => {
    setEditSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((id) => id !== tagId) : [...prev, tagId]
    );
  };

  const handleCreateNewTag = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = newTagName.trim();
    if (!trimmed) return;

    setIsCreatingTag(true);
    try {
      const created = await createTag({
        name: trimmed,
        school_percent: parseInt(newTagPercent, 10) || 0,
      });
      onTagCreated(created);
      setEditSelectedTagIds((prev) => [...prev, created.id]);
      setNewTagName('');
      setNewTagPercent('0');
      setIsTagFormOpen(false);
    } catch (err) {
      console.error('Failed to create tag', err);
    } finally {
      setIsCreatingTag(false);
    }
  };

  const handleUpdateClient = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!client || !editName.trim() || !editRateIndividual) return;

    setIsSubmitting(true);
    try {
      const indRate = Number(editRateIndividual);
      const pRate = editRatePair.trim() ? Number(editRatePair) : null;
      const gRate = editRateGroup.trim() ? Number(editRateGroup) : null;

      await updateClient(client.id, {
        name: editName.trim(),
        phone: editPhone.trim() || null,
        rate_individual: indRate,
        rate_pair: pRate,
        rate_group: gRate,
        tag_ids: editSelectedTagIds,
      });

      await onUpdated();
      onClose();
    } catch (error: unknown) {
      alert(error instanceof Error ? error.message : 'Ошибка обновления ученика');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!client) return null;

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title="Редактировать ученика"
      description="Обновите контакты, тарифную сетку ставок и прикрепленные теги."
      maxWidth="lg"
    >
      <form onSubmit={handleUpdateClient} className="space-y-4">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <GlassInput
            label="ФИО ученика"
            placeholder="например: Михаил Светлов"
            value={editName}
            onChange={(e) => setEditName(e.target.value)}
            required
          />
          <GlassInput
            label="Телефон (необязательно)"
            placeholder="+7 (999) 000-00-00"
            value={editPhone}
            onChange={(e) => setEditPhone(e.target.value)}
          />
        </div>

        {/* Тарифная сетка */}
        <div className="p-4 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 space-y-3 shadow-sm">
          <label className="block text-xs font-bold uppercase tracking-wider text-slate-700">
            Тарифная сетка ставок (₽ за 1 час)
          </label>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <GlassInput
              label={
                <span>
                  Индивидуально <span className="text-rose-500">*</span>
                </span>
              }
              labelClassName="text-[11px] font-semibold whitespace-nowrap"
              type="number"
              placeholder="1500"
              value={editRateIndividual}
              onChange={(e) => setEditRateIndividual(e.target.value)}
              required
              min="100"
            />
            <GlassInput
              label="В паре (опция)"
              labelClassName="text-[11px] font-semibold whitespace-nowrap"
              type="number"
              placeholder="1000"
              value={editRatePair}
              onChange={(e) => setEditRatePair(e.target.value)}
              min="100"
            />
            <GlassInput
              label="В группе (опция)"
              labelClassName="text-[11px] font-semibold whitespace-nowrap"
              type="number"
              placeholder="700"
              value={editRateGroup}
              onChange={(e) => setEditRateGroup(e.target.value)}
              min="100"
            />
          </div>
        </div>

        {/* Мультиселект динамических тегов */}
        <div className="p-4 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 space-y-3 shadow-sm">
          <div className="flex items-center justify-between">
            <label className="text-xs font-bold uppercase tracking-wider text-slate-700 flex items-center gap-1.5">
              <TagIcon className="w-3.5 h-3.5 text-indigo-500" />
              Теги и процент школы
            </label>
            <button
              type="button"
              onClick={() => setIsTagFormOpen(!isTagFormOpen)}
              className="text-xs font-semibold text-indigo-600 hover:text-indigo-700 flex items-center gap-1"
            >
              <Plus className="w-3.5 h-3.5" />
              {isTagFormOpen ? 'Скрыть форму' : 'Создать новый тег'}
            </button>
          </div>

          {isTagFormOpen && (
            <div className="p-3 rounded-2xl bg-white/60 border border-indigo-200/50 space-y-3 animate-in fade-in">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                <input
                  type="text"
                  placeholder="Название (например: Школа Фоксфорд)"
                  value={newTagName}
                  onChange={(e) => setNewTagName(e.target.value)}
                  className="rounded-xl px-3 py-2 text-xs text-slate-800 bg-white/80 border border-slate-200 focus:outline-none focus:border-indigo-500"
                />
                <div className="flex items-center gap-2">
                  <div className="relative flex-1">
                    <input
                      type="number"
                      placeholder="% школы"
                      value={newTagPercent}
                      onChange={(e) => setNewTagPercent(e.target.value)}
                      min="0"
                      max="100"
                      className="w-full rounded-xl px-3 py-2 text-xs text-slate-800 bg-white/80 border border-slate-200 focus:outline-none focus:border-indigo-500"
                    />
                    <Percent className="w-3 h-3 text-slate-400 absolute right-3 top-2.5 pointer-events-none" />
                  </div>
                  <GlassButton
                    type="button"
                    size="sm"
                    variant="primary"
                    isLoading={isCreatingTag}
                    onClick={handleCreateNewTag}
                  >
                    Добавить
                  </GlassButton>
                </div>
              </div>
            </div>
          )}

          <div className="flex flex-wrap gap-2 pt-1">
            {tags.map((t) => {
              const isSelected = editSelectedTagIds.includes(t.id);
              return (
                <button
                  key={t.id}
                  type="button"
                  onClick={() => toggleTagSelection(t.id)}
                  className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-semibold transition-all ${
                    isSelected
                      ? 'bg-indigo-600 text-white shadow-md shadow-indigo-500/20'
                      : 'bg-white/50 text-slate-700 hover:bg-white/80 border border-white/80'
                  }`}
                >
                  {isSelected && <Check className="w-3.5 h-3.5" />}
                  <span>{t.name}</span>
                  {t.school_percent > 0 && (
                    <span
                      className={`text-[10px] px-1.5 py-0.5 rounded-full ${
                        isSelected ? 'bg-indigo-500 text-white' : 'bg-black/5 text-slate-500'
                      }`}
                    >
                      {t.school_percent}%
                    </span>
                  )}
                </button>
              );
            })}
            {tags.length === 0 && !isTagFormOpen && (
              <p className="text-xs text-slate-400">
                Теги не созданы. Нажмите «Создать новый тег», чтобы добавить категорию.
              </p>
            )}
          </div>
        </div>

        <div className="pt-3 flex justify-end gap-3 border-t border-black/[0.05]">
          <GlassButton
            type="button"
            variant="secondary"
            onClick={onClose}
          >
            Отмена
          </GlassButton>
          <GlassButton
            type="submit"
            variant="primary"
            isLoading={isSubmitting}
          >
            Сохранить изменения
          </GlassButton>
        </div>
      </form>
    </GlassModal>
  );
};
