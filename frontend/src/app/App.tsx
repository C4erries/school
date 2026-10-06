import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from '../features/auth/AuthContext';
import { ProtectedRoute } from './ProtectedRoute';
import { AppLayout } from '../shared/components/AppLayout';
import { LoginPage } from '../pages/LoginPage';
import { RegisterPage } from '../pages/RegisterPage';
import { DashboardPage } from '../pages/DashboardPage';
import { AdminDashboard } from '../pages/admin/AdminDashboard';
import { TeacherSchedulePage } from '../pages/teacher/TeacherSchedulePage';
import { TeacherClientsPage } from '../pages/teacher/TeacherClientsPage';
import { TeacherFinancePage } from '../pages/teacher/TeacherFinancePage';
import { TeacherAnalyticsPage } from '../pages/teacher/TeacherAnalyticsPage';

export const App: React.FC = () => {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />

          {/* Защищенные маршруты, обернутые в сквозной AppLayout согласно ADR-006 */}
          <Route
            element={
              <ProtectedRoute>
                <AppLayout />
              </ProtectedRoute>
            }
          >
            {/* Общий дашборд платформы */}
            <Route path="/dashboard" element={<DashboardPage />} />

            {/* Панель администратора: кабинеты, распределение учеников, сквозной обзор */}
            <Route
              path="/admin"
              element={
                <ProtectedRoute allowedRoles={['owner', 'assistant']}>
                  <AdminDashboard />
                </ProtectedRoute>
              }
            />

            {/* Расписание преподавателя: закрепленные ученики, назначение урока, Apple Calendar сетка */}
            <Route
              path="/teacher/schedule"
              element={
                <ProtectedRoute allowedRoles={['teacher', 'owner']}>
                  <TeacherSchedulePage />
                </ProtectedRoute>
              }
            />

            {/* CRM преподавателя: клиенты, тарифная сетка, балансы часов, абонементы */}
            <Route
              path="/teacher/clients"
              element={
                <ProtectedRoute allowedRoles={['teacher', 'owner']}>
                  <TeacherClientsPage />
                </ProtectedRoute>
              }
            />

            {/* Финансовая бухгалтерия: сводка, долги, журнал оплат, партнерские выплаты, экспорт */}
            <Route
              path="/teacher/finance"
              element={
                <ProtectedRoute allowedRoles={['teacher', 'owner']}>
                  <TeacherFinancePage />
                </ProtectedRoute>
              }
            />

            {/* Статистика и аналитика: KPI, динамика во времени, форматы, рейтинг учеников */}
            <Route
              path="/teacher/analytics"
              element={
                <ProtectedRoute allowedRoles={['teacher', 'owner']}>
                  <TeacherAnalyticsPage />
                </ProtectedRoute>
              }
            />
          </Route>

          {/* Все остальные маршруты перенаправляем на /dashboard */}
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
};

export default App;
