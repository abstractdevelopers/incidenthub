const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface User {
  id: string;
  email: string;
  name: string;
}

export interface LoginResponse {
  token: string;
  id: string;
  email: string;
  name: string;
}

export type Severity = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
export type Status = 'OPEN' | 'INVESTIGATING' | 'MITIGATED' | 'RESOLVED';

export interface Incident {
  id: string;
  title: string;
  description: string;
  severity: Severity;
  status: Status;
  assignee_id?: string;
  assignee?: User;
  created_by: string;
  creator?: User;
  created_at: string;
  updated_at: string;
  resolved_at?: string;
}

export interface Comment {
  id: string;
  incident_id: string;
  user_id: string;
  author?: User;
  body: string;
  created_at: string;
  updated_at: string;
}

export interface DashboardStats {
  total: number;
  open: number;
  investigating: number;
  mitigated: number;
  resolved: number;
  critical: number;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

function getHeaders(extra?: Record<string, string>): HeadersInit {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  if (extra) {
    Object.assign(headers, extra);
  }
  return headers;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: getHeaders(options?.headers as Record<string, string>),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  if (res.status === 204) {
    return {} as T;
  }
  return res.json();
}

// Auth
export const authApi = {
  register: (data: { email: string; password: string; name: string }) =>
    request<User>('/api/auth/register', { method: 'POST', body: JSON.stringify(data) }),
  login: (data: { email: string; password: string }) =>
    request<LoginResponse>('/api/auth/login', { method: 'POST', body: JSON.stringify(data) }),
};

// Incidents
export const incidentsApi = {
  list: (params?: { status?: Status; severity?: Severity; search?: string; assignee?: string; page?: number; limit?: number }) => {
    const qs = new URLSearchParams();
    if (params?.status) qs.set('status', params.status);
    if (params?.severity) qs.set('severity', params.severity);
    if (params?.search) qs.set('search', params.search);
    if (params?.assignee) qs.set('assignee', params.assignee);
    if (params?.page) qs.set('page', String(params.page));
    if (params?.limit) qs.set('limit', String(params.limit));
    const query = qs.toString();
    return request<PaginatedResponse<Incident>>(`/api/incidents${query ? '?' + query : ''}`);
  },
  get: (id: string) => request<Incident>(`/api/incidents/${id}`),
  create: (data: { title: string; description: string; severity: Severity; status?: Status; assignee_id?: string }) =>
    request<Incident>('/api/incidents', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: Partial<Incident>) =>
    request<Incident>(`/api/incidents/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => request<void>(`/api/incidents/${id}`, { method: 'DELETE' }),
};

// Comments
export const commentsApi = {
  list: (incidentId: string) => request<Comment[]>(`/api/incidents/${incidentId}/comments`),
  create: (incidentId: string, data: { body: string }) =>
    request<Comment>(`/api/incidents/${incidentId}/comments`, { method: 'POST', body: JSON.stringify(data) }),
};

// Dashboard
export const dashboardApi = {
  stats: () =>
    request<{ stats: DashboardStats; recent: Incident[] }>('/api/dashboard/stats'),
};