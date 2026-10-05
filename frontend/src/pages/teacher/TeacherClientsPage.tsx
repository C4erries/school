import React, { useEffect, useState } from 'react';
import { Plus, UserPlus, Phone, Clock, Coins } from 'lucide-react';
import { Client } from '../../types/schedule';
import { getClients, createClient, addSubscription } from '../../api/schedule';
import { GlassCard } from '../../shared/components/GlassCard';
import { GlassButton } from '../../shared/components/GlassButton';
import { GlassModal } from '../../shared/components/GlassModal';
import { GlassInput } from '../../shared/components/GlassInput';
import { Badge } from '../../shared/components/Badge';

export const TeacherClientsPage: React.FC = () => {
  const [clients, setClients] = useState<Client[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [newClientName, setNewClientName] = useState('');
  const [newClientPhone, setNewClientPhone] = useState('');
  const [newClientRate, setNewClientRate] = useState('');
  const [newClientTag, setNewClientTag] = useState('');

  const [isSubModalOpen, setIsSubModalOpen] = useState(false);
  const [selectedClientId, setSelectedClientId] = useState<string | null>(null);
  const [subAmount, setSubAmount] = useState('');

  const loadClients = async () => {
    setIsLoading(true);
    try {
      const data = await getClients();
      setClients(data);
    } catch (e) {
      console.error(e);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadClients();
  }, []);

  const handleCreateClient = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newClientName || !newClientRate) return;
    try {
      await createClient({
        name: newClientName,
        phone: newClientPhone,
        base_rate: Number(newClientRate),
        tag: newClientTag
      });
      setIsCreateModalOpen(false);
      setNewClientName('');
      setNewClientPhone('');
      setNewClientRate('');
      setNewClientTag('');
      loadClients();
    } catch (error) {
      console.error(error);
    }
  };

  const handleAddSubscription = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedClientId || !subAmount) return;
    try {
      await addSubscription({
        client_id: selectedClientId,
        amount: Number(subAmount)
      });
      setIsSubModalOpen(false);
      setSubAmount('');
      loadClients();
    } catch (error) {
      console.error(error);
    }
  };

  const openSubModal = (clientId: string) => {
    setSelectedClientId(clientId);
    setIsSubModalOpen(true);
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">
            CRM Репетитора
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Управление учениками, ставками и абонементами
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

      {isLoading ? (
        <div className="text-center p-12 text-slate-500">Загрузка учеников...</div>
      ) : clients.length === 0 ? (
        <GlassCard className="text-center p-12 flex flex-col items-center">
          <div className="w-16 h-16 rounded-3xl bg-indigo-50 flex items-center justify-center text-indigo-400 mb-4">
            <UserPlus className="w-8 h-8" />
          </div>
          <h3 className="text-lg font-bold text-slate-900 mb-2">Нет учеников</h3>
          <p className="text-slate-500 text-sm max-w-sm mb-6">
            Добавьте своего первого ученика, чтобы начать вести расписание и учет.
          </p>
          <GlassButton variant="primary" onClick={() => setIsCreateModalOpen(true)}>
            Добавить первого ученика
          </GlassButton>
        </GlassCard>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
          {clients.map(client => (
            <GlassCard key={client.id} className="flex flex-col p-5">
              <div className="flex justify-between items-start mb-4">
                <div>
                  <h3 className="text-lg font-bold text-slate-900">{client.name}</h3>
                  {client.phone && (
                    <div className="flex items-center text-xs text-slate-500 mt-1 gap-1">
                      <Phone className="w-3 h-3" />
                      {client.phone}
                    </div>
                  )}
                </div>
                {client.tag && (
                  <Badge variant="amber" className="text-[10px] uppercase font-bold tracking-wider">
                    {client.tag}
                  </Badge>
                )}
              </div>

              <div className="flex-1 space-y-3 mb-6">
                <div className="flex justify-between items-center bg-slate-50 p-3 rounded-2xl border border-slate-100">
                  <div className="flex items-center gap-2 text-sm font-medium text-slate-700">
                    <Coins className="w-4 h-4 text-indigo-500" />
                    Ставка:
                  </div>
                  <span className="font-bold text-slate-900">{client.base_rate} ₽/ч</span>
                </div>
                <div className="flex justify-between items-center bg-slate-50 p-3 rounded-2xl border border-slate-100">
                  <div className="flex items-center gap-2 text-sm font-medium text-slate-700">
                    <Clock className="w-4 h-4 text-mint-500" />
                    Баланс:
                  </div>
                  <span className={`font-bold ${client.balance > 0 ? 'text-mint-600' : 'text-rose-500'}`}>
                    {client.balance} занятий
                  </span>
                </div>
              </div>

              <GlassButton
                variant="secondary"
                className="w-full"
                onClick={() => openSubModal(client.id)}
                icon={<Plus className="w-4 h-4" />}
              >
                Пополнить баланс
              </GlassButton>
            </GlassCard>
          ))}
        </div>
      )}

      {/* Модалка создания клиента */}
      <GlassModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="Новый ученик"
        description="Добавьте ученика для назначения уроков и учета абонементов."
      >
        <form onSubmit={handleCreateClient} className="space-y-4">
          <GlassInput
            label="Имя ученика"
            value={newClientName}
            onChange={e => setNewClientName(e.target.value)}
            required
          />
          <GlassInput
            label="Телефон (необязательно)"
            value={newClientPhone}
            onChange={e => setNewClientPhone(e.target.value)}
          />
          <GlassInput
            label="Базовая ставка (₽/ч)"
            type="number"
            value={newClientRate}
            onChange={e => setNewClientRate(e.target.value)}
            required
          />
          <GlassInput
            label="Тег / Комиссия школы (например, '5% школе')"
            value={newClientTag}
            onChange={e => setNewClientTag(e.target.value)}
          />
          <div className="pt-2 flex justify-end gap-3">
            <GlassButton type="button" variant="secondary" onClick={() => setIsCreateModalOpen(false)}>
              Отмена
            </GlassButton>
            <GlassButton type="submit" variant="primary">
              Сохранить
            </GlassButton>
          </div>
        </form>
      </GlassModal>

      {/* Модалка пополнения баланса */}
      <GlassModal
        isOpen={isSubModalOpen}
        onClose={() => setIsSubModalOpen(false)}
        title="Пополнить абонемент"
        description="Укажите количество занятий (часов) для добавления к балансу."
      >
        <form onSubmit={handleAddSubscription} className="space-y-4">
          <GlassInput
            label="Количество занятий"
            type="number"
            value={subAmount}
            onChange={e => setSubAmount(e.target.value)}
            required
            min="1"
          />
          <div className="pt-2 flex justify-end gap-3">
            <GlassButton type="button" variant="secondary" onClick={() => setIsSubModalOpen(false)}>
              Отмена
            </GlassButton>
            <GlassButton type="submit" variant="primary">
              Начислить
            </GlassButton>
          </div>
        </form>
      </GlassModal>
    </div>
  );
};
