export interface User {
  id: string;
  email: string;
  name: string;
  created_at?: string;
  updated_at?: string;
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