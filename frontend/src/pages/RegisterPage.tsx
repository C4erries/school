import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Mail, Lock, User as UserIcon, Phone, GraduationCap, BookOpen, ArrowRight, AlertCircle } from 'lucide-react';
import { useAuth } from '../features/auth/useAuth';
import { Role } from '../types/auth';
import { GlassCard } from '../shared/components/GlassCard';
import { GlassInput } from '../shared/components/GlassInput';
import { GlassButton } from '../shared/components/GlassButton';
import { LiquidBackground } from '../shared/components/LiquidBackground';

export const RegisterPage: React.FC = () => {
  const navigate = useNavigate();
  const { register } = useAuth();

  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState<Role>('student');
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!fullName.trim() || !email.trim() || !password) {
      setError('Заполните все обязательные поля');
      return;
    }

    if (password.length < 6) {
      setError('Пароль должен содержать не менее 6 символов');
      return;
    }

    setIsLoading(true);
    try {
      await register({
        full_name: fullName.trim(),
        email: email.trim(),
        phone: phone.trim() || undefined,
        password,
        role,
      });
      navigate('/dashboard');
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Произошла непредвиденная ошибка регистрации');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen w-full flex items-center justify-center p-4 relative overflow-hidden">
      <LiquidBackground />

      <div className="w-full max-w-md relative z-10 py-8">
        <div className="text-center mb-6">
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl liquid-glass mb-4 shadow-sm border border-white/80">
            <GraduationCap className="w-7 h-7 text-indigo-600" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-900">
            Создать аккаунт
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Присоединяйтесь к платформе школы
          </p>
        </div>

        <GlassCard className="shadow-2xl">
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="p-3.5 rounded-2xl bg-rose-500/10 border border-rose-500/20 flex items-start gap-2.5 text-rose-600 text-sm">
                <AlertCircle className="w-5 h-5 flex-shrink-0 mt-0.5" />
                <span>{error}</span>
              </div>
            )}

            {/* Выбор роли */}
            <div className="space-y-1.5 text-left">
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500 ml-1">
                Выберите роль
              </label>
              <div className="grid grid-cols-2 gap-3">
                <button
                  type="button"
                  onClick={() => setRole('student')}
                  className={`p-3.5 rounded-2xl border text-left transition-all duration-200 flex items-center gap-3 ${
                    role === 'student'
                      ? 'bg-indigo-600/10 border-indigo-500 text-indigo-900 shadow-sm'
                      : 'liquid-glass border-slate-200 text-slate-600 hover:border-slate-300'
                  }`}
                >
                  <BookOpen className={`w-5 h-5 ${role === 'student' ? 'text-indigo-600' : 'text-slate-400'}`} />
                  <div>
                    <div className="text-sm font-semibold">Ученик</div>
                    <div className="text-[11px] text-slate-500">Занятия и курсы</div>
                  </div>
                </button>

                <button
                  type="button"
                  onClick={() => setRole('teacher')}
                  className={`p-3.5 rounded-2xl border text-left transition-all duration-200 flex items-center gap-3 ${
                    role === 'teacher'
                      ? 'bg-indigo-600/10 border-indigo-500 text-indigo-900 shadow-sm'
                      : 'liquid-glass border-slate-200 text-slate-600 hover:border-slate-300'
                  }`}
                >
                  <GraduationCap className={`w-5 h-5 ${role === 'teacher' ? 'text-indigo-600' : 'text-slate-400'}`} />
                  <div>
                    <div className="text-sm font-semibold">Преподаватель</div>
                    <div className="text-[11px] text-slate-500">Расписание и финансы</div>
                  </div>
                </button>
              </div>
            </div>

            <GlassInput
              label="ФИО / Имя"
              type="text"
              placeholder="Иван Иванов"
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              icon={<UserIcon className="w-4 h-4" />}
              required
            />

            <GlassInput
              label="Электронная почта"
              type="email"
              placeholder="user@school.ru"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              icon={<Mail className="w-4 h-4" />}
              required
            />

            <GlassInput
              label="Телефон (необязательно)"
              type="tel"
              placeholder="+7 (999) 000-00-00"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              icon={<Phone className="w-4 h-4" />}
            />

            <GlassInput
              label="Пароль"
              type="password"
              placeholder="Минимум 6 символов"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              icon={<Lock className="w-4 h-4" />}
              required
            />

            <div className="pt-2">
              <GlassButton
                type="submit"
                variant="primary"
                size="lg"
                className="w-full"
                isLoading={isLoading}
                icon={<ArrowRight className="w-4 h-4" />}
              >
                Зарегистрироваться
              </GlassButton>
            </div>
          </form>

          <div className="mt-6 pt-6 border-t border-slate-200/50 text-center">
            <p className="text-sm text-slate-500">
              Уже зарегистрированы?{' '}
              <Link
                to="/login"
                className="font-semibold text-indigo-600 hover:text-indigo-500 transition-colors ml-1"
              >
                Войти
              </Link>
            </p>
          </div>
        </GlassCard>
      </div>
    </div>
  );
};
