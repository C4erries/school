export type Role = 'teacher' | 'assistant' | 'owner';

export interface User {
  id: string;
  email: string;
  full_name: string;
  phone?: string | null;
  role: Role;
  created_at: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

export interface AuthResponse {
  tokens: TokenPair;
  user: User;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  full_name: string;
  phone?: string;
  role?: Role;
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>[];
}
