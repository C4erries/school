import React, { useEffect, useState, useCallback } from 'react';
import {
  Plus,
  UserPlus,
  Phone,
  Clock,
  Coins,
  AlertCircle,
  RefreshCw,
  Tag as TagIcon,
  Check,
  Percent,
  Sparkles,
  Pencil,
} from 'lucide-react';
import { Client, Tag, SubscriptionFormat } from '../../types/schedule';
import {
  getClients,
  createClient,
  updateClient,
  addSubscription,
  getTags,
  createTag,
} from '../../api/schedule';
import { GlassCard } from '../../shared/components/GlassCard';
import { GlassButton } from '../../shared/components/GlassButton';
import { GlassModal } from '../../shared/components/GlassModal';
import { GlassInput } from '../../shared/components/GlassInput';
import { Badge } from '../../shared/components/Badge';

export const TeacherClientsPage: React.FC = () => {
  const [clients, setClients] = useState<Client[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Модалка создания клиента
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [newClientName, setNewClientName] = useState('');
  const [newClientPhone, setNewClientPhone] = useState('');
  const [rateIndividual, setRateIndividual] = useState('');
  const [ratePair, setRatePair] = useState('');
  const [rateGroup, setRateGroup] = useState('');
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([]);
  const [isSubmittingClient, setIsSubmittingClient] = useState(false);

  // Модалка редактирования клиента
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [editingClient, setEditingClient] = useState<Client | null>(null);
  const [editName, setEditName] = useState('');
  const [editPhone, setEditPhone] = useState('');
  const [editRateIndividual, setEditRateIndividual] = useState('');
  const [editRatePair, setEditRatePair] = useState('');
  const [editRateGroup, setEditRateGroup] = useState('');
  const [editSelectedTagIds, setEditSelectedTagIds] = useState<string[]>([]);
  const [isSubmittingEdit, setIsSubmittingEdit] = useState(false);

  // Форма добавления нового тега на лету прямо в модалке
  const [isTagFormOpen, setIsTagFormOpen] = useState(false);
  const [newTagName, setNewTagName] = useState('');
  const [newTagPercent, setNewTagPercent] = useState('0');
  const [isCreatingTag, setIsCreatingTag] = useState(false);

  // Модалка пополнения абонемента
  const [isSubModalOpen, setIsSubModalOpen] = useState(false);
  const [selectedClient, setSelectedClient] = useState<Client | null>(null);
  const [subFormat, setSubFormat] = useState<SubscriptionFormat>('individual');
  const [subHours, setSubHours] = useState('8');
  const [isSubmittingSub, setIsSubmittingSub] = useState(false);

  const loadData = useCallback(async () => {
    setIsLoading(true);
    setErrorMessage(null);
    try {
      const [clientsData, tagsData] = await Promise.all([
        getClients(),
        getTags().catch(() => []),
      ]);
      setClients(clientsData);
      setTags(tagsData);
    } catch (e: unknown) {
      console.error('Error loading CRM clients:', e);
      setErrorMessage(e instanceof Error ? e.message : 'Ошибка подключения к серверу');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Создание нового тега на лету
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
      setTags((prev) => [...prev, created]);
      if (isEditModalOpen) {
        setEditSelectedTagIds((prev) => [...prev, created.id]);
      } else {
        setSelectedTagIds((prev) => [...prev, created.id]);
      }
      setNewTagName('');
      setNewTagPercent('0');
      setIsTagFormOpen(false);
    } catch (err) {
      console.error('Failed to create tag', err);
    } finally {
      setIsCreatingTag(false);
    }
  };

  const toggleTagSelection = (tagId: string) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((id) => id !== tagId) : [...prev, tagId]
    );
  };

  const toggleEditTagSelection = (tagId: string) => {
    setEditSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((id) => id !== tagId) : [...prev, tagId]
    );
  };

  const openEditModal = (client: Client) => {
    setEditingClient(client);
    setEditName(client.name);
    setEditPhone(client.phone || '');
    setEditRateIndividual(String(client.rate_individual ?? client.base_rate ?? ''));
    setEditRatePair(client.rate_pair != null ? String(client.rate_pair) : '');
    setEditRateGroup(client.rate_group != null ? String(client.rate_group) : '');
    const clientTagIds = client.tags && client.tags.length > 0
      ? client.tags.map((t) => t.id)
      : (client.tag_ids || []);
    setEditSelectedTagIds(clientTagIds);
    setIsTagFormOpen(false);
    setIsEditModalOpen(true);
  };

  const handleUpdateClient = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingClient || !editName.trim() || !editRateIndividual) return;

    setIsSubmittingEdit(true);
    try {
      const indRate = Number(editRateIndividual);
      const pRate = editRatePair.trim() ? Number(editRatePair) : null;
      const gRate = editRateGroup.trim() ? Number(editRateGroup) : null;

      const updated = await updateClient(editingClient.id, {
        name: editName.trim(),
        phone: editPhone.trim() || null,
        rate_individual: indRate,
        rate_pair: pRate,
        rate_group: gRate,
        tag_ids: editSelectedTagIds,
      });

      setClients((prev) =>
        prev.map((c) => (c.id === updated.id ? { ...c, ...updated } : c))
      );
      setIsEditModalOpen(false);
      setEditingClient(null);
      await loadData();
    } catch (error: unknown) {
      alert(error instanceof Error ? error.message : 'Ошибка обновления ученика');
    } finally {
      setIsSubmittingEdit(false);
    }
  };

  const handleCreateClient = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newClientName.trim() || !rateIndividual) return;

    setIsSubmittingClient(true);
    try {
      const indRate = Number(rateIndividual);
      const pRate = ratePair.trim() ? Number(ratePair) : undefined;
      const gRate = rateGroup.trim() ? Number(rateGroup) : undefined;

      // Находим выбранные теги для вычисления макс. комиссии школы
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

      setIsCreateModalOpen(false);
      setNewClientName('');
      setNewClientPhone('');
      setRateIndividual('');
      setRatePair('');
      setRateGroup('');
      setSelectedTagIds([]);
      await loadData();
    } catch (error: unknown) {
      alert(error instanceof Error ? error.message : 'Ошибка сохранения ученика');
    } finally {
      setIsSubmittingClient(false);
    }
  };

  const handleAddSubscription = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedClient || !subHours) return;

    setIsSubmittingSub(true);
    try {
      await addSubscription({
        client_id: selectedClient.id,
        format: subFormat,
        hours: Number(subHours),
      });
      setIsSubModalOpen(false);
      setSubHours('8');
      await loadData();
    } catch (error: unknown) {
      alert(error instanceof Error ? error.message : 'Ошибка пополнения абонемента');
    } finally {
      setIsSubmittingSub(false);
    }
  };

  const openSubModal = (client: Client) => {
    setSelectedClient(client);
    setSubFormat('individual');
    setSubHours('8');
    setIsSubModalOpen(true);
  };

  // Рендер бейджа баланса по формату
  const renderBalanceBadge = (hours: number, label: string) => {
    if (hours > 0) {
      return (
        <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-xl text-xs font-bold bg-emerald-500/15 text-emerald-800 border border-emerald-400/30 shadow-sm">
          <Clock className="w-3.5 h-3.5 text-emerald-600" />
          {hours} ч {label}
        </span>
      );
    }
    if (hours < 0) {
      return (
        <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-xl text-xs font-bold bg-rose-500/15 text-rose-800 border border-rose-400/30 shadow-sm">
          <AlertCircle className="w-3.5 h-3.5 text-rose-600" />
          {hours} ч (долг)
        </span>
      );
    }
    return (
      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-xl text-xs font-semibold bg-white/40 text-slate-500 border border-white/60">
        0 ч {label}
      </span>
    );
  };

  return (
    <div className="space-y-6">
      {/* Верхний заголовок */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">
            CRM Репетитора
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Управление учениками, тарифная сетка ставок и баланс часов по форматам
          </p>
        </div>
        <GlassButton
          variant="primary"
          onClick={() => setIsCreateModalOpen(true)}
          icon={<UserPlus className="w-4 h-4" />}
        >
          Добавить ученика
        </GlassButton>
      </div>

      {/* Ошибка загрузки */}
      {errorMessage && (
        <div className="p-4 rounded-3xl bg-rose-500/10 border border-rose-500/20 backdrop-blur-md flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
            <p className="text-sm font-medium text-rose-900">{errorMessage}</p>
          </div>
          <GlassButton
            variant="secondary"
            size="sm"
            onClick={loadData}
            icon={<RefreshCw className="w-3.5 h-3.5" />}
          >
            Повторить запрос
          </GlassButton>
        </div>
      )}

      {/* Состояние загрузки */}
      {isLoading ? (
        <div className="p-16 flex flex-col items-center justify-center liquid-glass rounded-3xl">
          <RefreshCw className="w-8 h-8 animate-spin text-indigo-600 mb-3" />
          <p className="text-sm font-medium text-slate-600">Загрузка данных CRM...</p>
        </div>
      ) : clients.length === 0 ? (
        /* Пустой список */
        <GlassCard className="text-center p-12 flex flex-col items-center">
          <div className="w-16 h-16 rounded-3xl bg-indigo-50/70 border border-indigo-200/50 flex items-center justify-center text-indigo-500 mb-4 shadow-sm">
            <UserPlus className="w-8 h-8" />
          </div>
          <h3 className="text-lg font-bold text-slate-900 mb-2">Ученики пока не добавлены</h3>
          <p className="text-slate-500 text-sm max-w-sm mb-6">
            Добавьте первого ученика, настройте тарифную сетку и пополните баланс часов.
          </p>
          <GlassButton
            variant="primary"
            onClick={() => setIsCreateModalOpen(true)}
            icon={<Plus className="w-4 h-4" />}
          >
            Добавить первого ученика
          </GlassButton>
        </GlassCard>
      ) : (
        /* Сетка карточек учеников */
        <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-5">
          {clients.map((client) => {
            const indBal = client.balances?.individual_hours ?? client.balance ?? 0;
            const pairBal = client.balances?.pair_hours ?? 0;
            const grpBal = client.balances?.group_hours ?? 0;

            return (
              <GlassCard
                key={client.id}
                className="flex flex-col p-5 sm:p-6 transition-all duration-200 hover:shadow-lg hover:border-white/90"
              >
                {/* Шапка карточки */}
                <div className="flex justify-between items-start mb-4">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <h3 className="text-lg font-bold text-slate-900 tracking-tight">
                        {client.name}
                      </h3>
                      <button
                        type="button"
                        onClick={() => openEditModal(client)}
                        className="p-1.5 rounded-xl hover:bg-white/40 text-slate-400 hover:text-slate-700 transition-colors"
                        title="Редактировать ученика"
                      >
                        <Pencil className="w-3.5 h-3.5" />
                      </button>
                    </div>
                    {client.phone && (
                      <div className="flex items-center text-xs text-slate-500 gap-1.5 font-medium">
                        <Phone className="w-3.5 h-3.5 text-slate-400" />
                        {client.phone}
                      </div>
                    )}
                  </div>

                  {/* Теги ученика */}
                  <div className="flex flex-wrap gap-1 justify-end max-w-[150px]">
                    {client.tags && client.tags.length > 0 ? (
                      client.tags.map((t) => (
                        <Badge
                          key={t.id}
                          variant="amber"
                          className="text-[10px] font-bold tracking-tight"
                        >
                          {t.name} {t.school_percent > 0 ? `(${t.school_percent}%)` : ''}
                        </Badge>
                      ))
                    ) : client.tag ? (
                      <Badge variant="amber" className="text-[10px] font-bold tracking-tight">
                        {client.tag}
                      </Badge>
                    ) : null}
                  </div>
                </div>

                {/* Тарифная сетка ставок */}
                <div className="space-y-2 mb-4">
                  <div className="text-[11px] font-bold uppercase tracking-wider text-slate-500 flex items-center gap-1.5 ml-1">
                    <Coins className="w-3.5 h-3.5 text-indigo-500" />
                    Тарифная сетка
                  </div>

                  <div className="grid grid-cols-3 gap-2">
                    <div className="p-2.5 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 shadow-sm text-center">
                      <span className="block text-[10px] font-semibold text-slate-500 uppercase">
                        Индив.
                      </span>
                      <span className="text-xs sm:text-sm font-bold text-slate-900">
                        {client.rate_individual} ₽/ч
                      </span>
                    </div>

                    <div className="p-2.5 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 shadow-sm text-center">
                      <span className="block text-[10px] font-semibold text-slate-500 uppercase">
                        Пара
                      </span>
                      <span className="text-xs sm:text-sm font-bold text-slate-900">
                        {client.rate_pair ? `${client.rate_pair} ₽/ч` : '—'}
                      </span>
                    </div>

                    <div className="p-2.5 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 shadow-sm text-center">
                      <span className="block text-[10px] font-semibold text-slate-500 uppercase">
                        Группа
                      </span>
                      <span className="text-xs sm:text-sm font-bold text-slate-900">
                        {client.rate_group ? `${client.rate_group} ₽/ч` : '—'}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Баланс часов по форматам */}
                <div className="p-3.5 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 shadow-sm mb-5 space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-[11px] font-bold uppercase tracking-wider text-slate-500">
                      Остаток абонемента
                    </span>
                    <span className="text-xs font-semibold text-slate-600">
                      Всего: <strong className="text-slate-900">{client.balance ?? indBal} ч</strong>
                    </span>
                  </div>

                  <div className="flex flex-wrap items-center gap-2 pt-1">
                    {renderBalanceBadge(indBal, 'индив.')}
                    {(client.rate_pair || pairBal !== 0) && renderBalanceBadge(pairBal, 'пара')}
                    {(client.rate_group || grpBal !== 0) && renderBalanceBadge(grpBal, 'группа')}
                  </div>
                </div>

                {/* Кнопка пополнения абонемента */}
                <div className="mt-auto">
                  <GlassButton
                    variant="secondary"
                    className="w-full"
                    onClick={() => openSubModal(client)}
                    icon={<Plus className="w-4 h-4 text-indigo-600" />}
                  >
                    Пополнить абонемент
                  </GlassButton>
                </div>
              </GlassCard>
            );
          })}
        </div>
      )}

      {/* МОДАЛКА: Новый ученик */}
      <GlassModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
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
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-700">
              Тарифная сетка ставок (₽ за 1 час)
            </label>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <GlassInput
                label={<span>Индивидуально <span className="text-rose-500">*</span></span>}
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

            {/* Компактная форма создания тега прямо в модалке */}
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

            {/* Список чипсов тегов для выбора */}
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
              onClick={() => setIsCreateModalOpen(false)}
            >
              Отмена
            </GlassButton>
            <GlassButton
              type="submit"
              variant="primary"
              isLoading={isSubmittingClient}
            >
              Сохранить ученика
            </GlassButton>
          </div>
        </form>
      </GlassModal>

      {/* МОДАЛКА: Редактировать ученика */}
      <GlassModal
        isOpen={isEditModalOpen}
        onClose={() => {
          setIsEditModalOpen(false);
          setEditingClient(null);
        }}
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
                label={<span>Индивидуально <span className="text-rose-500">*</span></span>}
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

            {/* Компактная форма создания тега прямо в модалке */}
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

            {/* Список чипсов тегов для выбора */}
            <div className="flex flex-wrap gap-2 pt-1">
              {tags.map((t) => {
                const isSelected = editSelectedTagIds.includes(t.id);
                return (
                  <button
                    key={t.id}
                    type="button"
                    onClick={() => toggleEditTagSelection(t.id)}
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
              onClick={() => {
                setIsEditModalOpen(false);
                setEditingClient(null);
              }}
            >
              Отмена
            </GlassButton>
            <GlassButton
              type="submit"
              variant="primary"
              isLoading={isSubmittingEdit}
            >
              Сохранить изменения
            </GlassButton>
          </div>
        </form>
      </GlassModal>

      {/* МОДАЛКА: Пополнить абонемент */}
      <GlassModal
        isOpen={isSubModalOpen}
        onClose={() => setIsSubModalOpen(false)}
        title="Пополнить абонемент"
        description={
          selectedClient
            ? `Начисление оплаченных часов для ${selectedClient.name}`
            : 'Выберите формат и количество часов'
        }
      >
        <form onSubmit={handleAddSubscription} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 mb-2 ml-1">
              Формат занятия
            </label>
            <div className="grid grid-cols-3 gap-2 p-1 rounded-2xl bg-black/[0.04] border border-black/[0.05]">
              <button
                type="button"
                onClick={() => setSubFormat('individual')}
                className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                  subFormat === 'individual'
                    ? 'bg-white text-indigo-600 shadow-sm'
                    : 'text-slate-600 hover:text-slate-900'
                }`}
              >
                Индив.
              </button>
              <button
                type="button"
                onClick={() => setSubFormat('pair')}
                className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                  subFormat === 'pair'
                    ? 'bg-white text-indigo-600 shadow-sm'
                    : 'text-slate-600 hover:text-slate-900'
                }`}
              >
                В паре
              </button>
              <button
                type="button"
                onClick={() => setSubFormat('group')}
                className={`py-2 px-3 rounded-xl text-xs font-semibold transition-all ${
                  subFormat === 'group'
                    ? 'bg-white text-indigo-600 shadow-sm'
                    : 'text-slate-600 hover:text-slate-900'
                }`}
              >
                Группа
              </button>
            </div>
          </div>

          <GlassInput
            label="Количество часов"
            type="number"
            step="0.5"
            min="0.5"
            value={subHours}
            onChange={(e) => setSubHours(e.target.value)}
            required
            icon={<Clock className="w-4 h-4" />}
          />

          {/* Быстрые пресеты часов */}
          <div className="flex items-center gap-2 pt-1">
            <span className="text-xs text-slate-400 font-medium">Пресеты:</span>
            {['4', '8', '12', '16'].map((preset) => (
              <button
                key={preset}
                type="button"
                onClick={() => setSubHours(preset)}
                className={`px-2.5 py-1 rounded-lg text-xs font-semibold transition-all ${
                  subHours === preset
                    ? 'bg-indigo-600 text-white shadow-sm'
                    : 'bg-white/40 text-slate-700 hover:bg-white/70 border border-white/60'
                }`}
              >
                {preset} ч
              </button>
            ))}
          </div>

          {/* Расчет суммы к оплате */}
          {selectedClient && (
            <div className="p-3.5 rounded-2xl bg-white/30 backdrop-blur-md border border-white/40 flex items-center justify-between text-xs shadow-sm">
              <span className="text-slate-500 font-medium flex items-center gap-1.5">
                <Sparkles className="w-3.5 h-3.5 text-indigo-500" />
                Ориентировочная сумма:
              </span>
              <span className="font-bold text-slate-900 text-sm">
                {(
                  Number(subHours || 0) *
                  (subFormat === 'individual'
                    ? selectedClient.rate_individual
                    : subFormat === 'pair'
                    ? selectedClient.rate_pair || selectedClient.rate_individual
                    : selectedClient.rate_group || selectedClient.rate_individual)
                ).toLocaleString('ru-RU')}{' '}
                ₽
              </span>
            </div>
          )}

          <div className="pt-3 flex justify-end gap-3 border-t border-black/[0.05]">
            <GlassButton
              type="button"
              variant="secondary"
              onClick={() => setIsSubModalOpen(false)}
            >
              Отмена
            </GlassButton>
            <GlassButton
              type="submit"
              variant="mint"
              isLoading={isSubmittingSub}
            >
              Начислить часы
            </GlassButton>
          </div>
        </form>
      </GlassModal>
    </div>
  );
};

export default TeacherClientsPage;
