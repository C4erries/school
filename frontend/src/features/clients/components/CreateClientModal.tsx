import React, { useState, useEffect } from 'react';
import { GlassModal } from '../../../shared/components/GlassModal';
import { GlassInput } from '../../../shared/components/GlassInput';
import { GlassButton } from '../../../shared/components/GlassButton';
import { Tag, UserDefaultRates } from '../../../types/schedule';
import { createClient, createTag } from '../../../api/schedule';
import { Plus, Tag as TagIcon, Check, Percent } from 'lucide-react';

interface CreateClientModalProps {
  isOpen: boolean;
  onClose: () => void;
  tags: Tag[];
  defaultRates?: UserDefaultRates | null;
  onCreated: () => Promise<void>;
  onTagCreated: (tag: Tag) => void;
}

export const CreateClientModal: React.FC<CreateClientModalProps> = ({
  isOpen,
  onClose,
  tags,
  defaultRates,
  onCreated,
  onTagCreated,
}) => {
  const [newClientName, setNewClientName] = useState('');
  const [newClientPhone, setNewClientPhone] = useState('');
  const [rateIndividual, setRateIndividual] = useState('1500');
  const [ratePair, setRatePair] = useState('1000');
  const [rateGroup, setRateGroup] = useState('700');
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Форма добавления нового тега на лету
  const [isTagFormOpen, setIsTagFormOpen] = useState(false);
  const [newTagName, setNewTagName] = useState('');
  const [newTagPercent, setNewTagPercent] = useState('0');
  const [isCreatingTag, setIsCreatingTag] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setNewClientName('');
      setNewClientPhone('');
      setSelectedTagIds([]);
      setIsTagFormOpen(false);

      if (defaultRates) {
        setRateIndividual(String(defaultRates.rate_individual || 1500));
        setRatePair(defaultRates.rate_pair ? String(defaultRates.rate_pair) : '1000');
        setRateGroup(defaultRates.rate_group ? String(defaultRates.rate_group) : '700');
      } else {
        setRateIndividual('1500');
        setRatePair('1000');
        setRateGroup('700');
      }
    }
  }, [isOpen, defaultRates]);

  const toggleTagSelection = (tagId: string) => {
    setSelectedTagIds((prev) =>
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
      setSelectedTagIds((prev) => [...prev, created.id]);
      setNewTagName('');
      setNewTagPercent('0');
      setIsTagFormOpen(false);
    } catch (err) {
      console.error('Failed to create tag', err);
    } finally {
      setIsCreatingTag(false);
    }
  };

  const handleCreateClient = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newClientName.trim() || !rateIndividual) return;

    setIsSubmitting(true);
    try {
      const indRate = Number(rateIndividual);
      const pRate = ratePair.trim() ? Number(ratePair) : undefined;
      const gRate = rateGroup.trim() ? Number(rateGroup) : undefined;

      const chosenTags = tags.filter((t) => selectedTagIds.includes(t.id));
      const maxSchoolPercent = chosenTags.reduce(
        (max, t) => Math.max(max, t.school_percent),
        0
      );

      await createClient({
        name: newClientName.trim(),
        phone: newClientPhone.trim() || undefined,
        rate_individual: indRate,
        rate_pair: pRate,
        rate_group: gRate,
        school_percent_tag: maxSchoolPercent,
        tag_ids: selectedTagIds,
      });

      await onCreated();
      onClose();
    } catch (error: unknown) {
      alert(error instanceof Error ? error.message : 'Ошибка сохранения ученика');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <GlassModal
      isOpen={isOpen}
      onClose={onClose}
      title="Новый ученик"
      description="Укажите контакты, тарифную сетку ставок и выберите теги школы."
      maxWidth="lg"
    >
      <form onSubmit={handleCreateClient} className="space-y-4">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <GlassInput
            label="ФИО ученика"
            placeholder="например: Михаил Светлов"
            value={newClientName}
            onChange={(e) => setNewClientName(e.target.value)}
            required
          />
          <GlassInput
            label="Телефон (необязательно)"
            placeholder="+7 (999) 000-00-00"
            value={newClientPhone}
            onChange={(e) => setNewClientPhone(e.target.value)}
          />
        </div>

        {/* Тарифная сетка */}
        <div className="p-4 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 space-y-3 shadow-sm">
          <div className="flex items-center justify-between">
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-700">
              Тарифная сетка ставок (₽ за 1 час)
            </label>
            {defaultRates && (
              <span className="text-[10px] text-indigo-600 font-semibold">
                Из базового прайса
              </span>
            )}
          </div>
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
              value={rateIndividual}
              onChange={(e) => setRateIndividual(e.target.value)}
              required
              min="100"
            />
            <GlassInput
              label="В паре (опция)"
              labelClassName="text-[11px] font-semibold whitespace-nowrap"
              type="number"
              placeholder="1000"
              value={ratePair}
              onChange={(e) => setRatePair(e.target.value)}
              min="100"
            />
            <GlassInput
              label="В группе (опция)"
              labelClassName="text-[11px] font-semibold whitespace-nowrap"
              type="number"
              placeholder="700"
              value={rateGroup}
              onChange={(e) => setRateGroup(e.target.value)}
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
              const isSelected = selectedTagIds.includes(t.id);
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
            Сохранить ученика
          </GlassButton>
        </div>
      </form>
    </GlassModal>
  );
};
