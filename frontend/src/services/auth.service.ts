import { api } from './api';

export interface LoginCredentials {
  email: string;
  password?: string;
}

export interface User {
  id: number;
  email: string;
  role: 'admin' | 'driver';
}

// Shape returned inside the 'data' field of the Go backend response
interface BackendLoginResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  role: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

// Decode JWT payload (no verification — just extract claims for UI use)
function decodeJwt(token: string): Record<string, any> {
  try {
    const payload = token.split('.')[1];
    return JSON.parse(atob(payload));
  } catch {
    return {};
  }
}

export const authService = {
  login: async (credentials: LoginCredentials): Promise<LoginResponse> => {
    // api.post returns the full response object { status, data, message }
    const res = await api.post('/auth/login', { data: credentials }) as { data: BackendLoginResponse };
    const backendData = res.data;

    const claims = decodeJwt(backendData.access_token);

    return {
      token: backendData.access_token,
      user: {
        id: claims.sub ?? 0,
        email: claims.email ?? credentials.email,
        role: (backendData.role as 'admin' | 'driver') ?? 'driver',
      },
    };
  },
};
