const BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:5000/api/v1';

export class ApiError extends Error {
  status: number;
  data: any;

  constructor(status: number, message: string, data?: any) {
    super(message);
    this.status = status;
    this.data = data;
    this.name = 'ApiError';
  }
}

interface FetchOptions extends RequestInit {
  data?: any;
}

export const api = {
  get: (endpoint: string, options?: FetchOptions) => request(endpoint, { ...options, method: 'GET' }),
  post: (endpoint: string, options?: FetchOptions) => request(endpoint, { ...options, method: 'POST' }),
  patch: (endpoint: string, options?: FetchOptions) => request(endpoint, { ...options, method: 'PATCH' }),
  put: (endpoint: string, options?: FetchOptions) => request(endpoint, { ...options, method: 'PUT' }),
  delete: (endpoint: string, options?: FetchOptions) => request(endpoint, { ...options, method: 'DELETE' }),
};

async function request(endpoint: string, customConfig: FetchOptions = {}) {
  const token = localStorage.getItem('token');
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  if (customConfig.headers) {
    Object.assign(headers, customConfig.headers);
  }

  const config: RequestInit = {
    ...customConfig,
    headers,
  };

  if (customConfig.data) {
    config.body = JSON.stringify(customConfig.data);
  }

  try {
    const response = await fetch(`${BASE_URL}${endpoint}`, config);
    
    // Attempt to parse JSON response
    let data;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      data = await response.json();
    } else {
      data = await response.text();
    }

    if (response.ok) {
      return data;
    }

    // Unauthenticated -> maybe clear token and redirect in the future
    if (response.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.dispatchEvent(new Event('unauthorized'));
    }

    const errorMessage = data?.message || data?.error || response.statusText;
    throw new ApiError(response.status, errorMessage, data);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }
    // Network errors or JSON parsing errors
    throw new ApiError(500, error instanceof Error ? error.message : 'Unknown network error');
  }
}
