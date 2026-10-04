import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from '../features/auth/AuthContext';
import { ProtectedRoute } from './ProtectedRoute';
import { LoginPage } from '../pages/LoginPage';
import { RegisterPage } from '../pages/RegisterPage';
import { DashboardPage } from '../pages/DashboardPage';
import { AdminDashboard } from '../pages/admin/AdminDashboard';
import { TeacherSchedulePage } from '../pages/teacher/TeacherSchedulePage';
import { StudentLessonsPage } from '../pages/student/StudentLessonsPage';

export const App: React.FC = () => {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />

          {/* Общий дашборд платформы */}
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <DashboardPage />
              </ProtectedRoute>
            }
          />

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

          {/* Личный кабинет ученика: подтверждение в 1 клик, отклонение с причиной, созвоны */}
          <Route
            path="/student/lessons"
            element={
              <ProtectedRoute allowedRoles={['student', 'owner']}>
                <StudentLessonsPage />
              </ProtectedRoute>
            }
          />

          {/* Все остальные маршруты перенаправляем на /dashboard */}
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
};

export default App;
