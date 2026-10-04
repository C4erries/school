import { AuthResponse, LoginRequest, RegisterRequest, User, ApiError } from '../types/auth';

const BASE_URL = '/api/v1/auth';

class AuthApiError extends Error {
  code: string;
  details?: Record<string, unknown>[];

  constructor(error: ApiError) {
    super(error.message);
    this.name = 'AuthApiError';
    this.code = error.code;
    this.details = error.details;
  }
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let errorData: { error?: ApiError };
    try {
      errorData = await res.json();
    } catch {
      throw new Error(`HTTP error ${res.status}: ${res.statusText}`);
    }

    if (errorData.error) {
      throw new AuthApiError(errorData.error);
    }
    throw new Error(`HTTP error ${res.status}`);
  }
  return res.json();
}

export async function loginApi(data: LoginRequest): Promise<AuthResponse> {
  const res = await fetch(`${BASE_URL}/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  return handleResponse<AuthResponse>(res);
}

export async function registerApi(data: RegisterRequest): Promise<AuthResponse> {
  const res = await fetch(`${BASE_URL}/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  return handleResponse<AuthResponse>(res);
}

export async function refreshApi(refreshToken: string): Promise<AuthResponse> {
  const res = await fetch(`${BASE_URL}/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
  return handleResponse<AuthResponse>(res);
}

export async function getMeApi(accessToken: string): Promise<User> {
  const res = await fetch(`${BASE_URL}/me`, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
  });
  return handleResponse<User>(res);
}

export { AuthApiError };
