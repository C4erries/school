import React, { useState, useEffect, useCallback } from 'react';
import { User, LoginRequest, RegisterRequest } from '../../types/auth';
import { loginApi, registerApi, refreshApi, getMeApi } from '../../api/auth';
import { AuthContext } from './context';

const ACCESS_TOKEN_KEY = 'school_access_token';
const REFRESH_TOKEN_KEY = 'school_refresh_token';

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  const saveTokens = (access: string, refresh: string) => {
    localStorage.setItem(ACCESS_TOKEN_KEY, access);
    localStorage.setItem(REFRESH_TOKEN_KEY, refresh);
  };

  const clearTokens = () => {
    localStorage.removeItem(ACCESS_TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
  };

  const logout = useCallback(() => {
    clearTokens();
    setUser(null);
  }, []);

  // Первоначальная инициализация сессии
  useEffect(() => {
    const initAuth = async () => {
      const accessToken = localStorage.getItem(ACCESS_TOKEN_KEY);
      const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY);

      if (!accessToken && !refreshToken) {
        setIsLoading(false);
        return;
      }

      if (accessToken) {
        try {
          const profile = await getMeApi(accessToken);
          setUser(profile);
          setIsLoading(false);
          return;
        } catch {
          // Access token мог истечь, пробуем обновить через refresh
        }
      }

      if (refreshToken) {
        try {
          const authData = await refreshApi(refreshToken);
          saveTokens(authData.tokens.access_token, authData.tokens.refresh_token);
          setUser(authData.user);
        } catch {
          clearTokens();
          setUser(null);
        }
      } else {
        clearTokens();
      }

      setIsLoading(false);
    };

    initAuth();
  }, []);

  const login = async (credentials: LoginRequest) => {
    const data = await loginApi(credentials);
    saveTokens(data.tokens.access_token, data.tokens.refresh_token);
    setUser(data.user);
  };

  const register = async (data: RegisterRequest) => {
    const res = await registerApi(data);
    saveTokens(res.tokens.access_token, res.tokens.refresh_token);
    setUser(res.user);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        register,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

