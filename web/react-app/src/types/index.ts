// API Response Types
export interface ApiResponse<T = any> {
  status: string;
  data?: T;
  message?: string;
  error?: string;
}

// Node Types
export interface Node {
  name: string;
  environment: string;
  host: string;
  port: number;
  username?: string;
  is_connected: boolean;
  process_count?: number;
  last_ping?: string;
}

// Process Types
export interface Process {
  name: string;
  group: string;
  state: number;
  state_string: string;
  description: string;
  start_time: string;
  stop_time: string;
  pid: number;
  exit_status: number;
  uptime: number;
  uptime_human: string;
}

// User Types
export interface User {
  id: string;
  username: string;
  email: string;
  is_admin: boolean;
  role?: string;
  created_at: string;
  updated_at: string;
}

// Alert Types
export interface AlertRule {
  id: number;
  name: string;
  description: string;
  metric: string;
  condition: string;
  threshold: number;
  duration: number;
  severity: 'low' | 'medium' | 'high' | 'critical';
  enabled: boolean;
  node_id?: number;
  process_name?: string;
  tags?: string;
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface Alert {
  id: number;
  rule_id: number;
  node_id?: number;
  process_name?: string;
  message: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  status: 'active' | 'acknowledged' | 'resolved';
  value: number;
  start_time: string;
  end_time?: string;
  acked_by?: number;
  acked_at?: string;
  resolved_by?: number;
  resolved_at?: string;
  created_at: string;
}

// Activity Log Types
export interface ActivityLog {
  id: number;
  user_id: string;
  username: string;
  action: string;
  resource_type: string;
  resource_id: string;
  details: string;
  ip_address: string;
  user_agent: string;
  created_at: string;
}

// System Stats Types
export interface SystemStats {
  total_nodes: number;
  connected_nodes: number;
  running_processes: number;
  stopped_processes: number;
  active_alerts: number;
  timestamp: string;
}

// WebSocket Message Types
export interface WebSocketMessage {
  type: string;
  data?: any;
  payload?: any;
  timestamp?: number;
}

// Chart Data Types
export interface ChartDataPoint {
  timestamp: string;
  value: number;
  label?: string;
}

// Role and Permission Types
export interface Role {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface Permission {
  id: string;
  name: string;
  description: string;
  resource: string;
  action: string;
}

// Configuration Types
export interface Configuration {
  id: number;
  key: string;
  value: string;
  category: string;
  description: string;
  is_sensitive: boolean;
  created_at: string;
  updated_at: string;
}

// Log Entry Types
export interface LogEntry {
  id: number;
  node_id?: number;
  process_name?: string;
  level: 'debug' | 'info' | 'warning' | 'error' | 'critical';
  message: string;
  source: string;
  timestamp: string;
  metadata?: string;
}

// Process Group Types
export interface ProcessGroup {
  id: number;
  name: string;
  description: string;
  environment: string;
  startup_order: number;
  created_at: string;
  updated_at: string;
}

// Scheduled Task Types
export interface ScheduledTask {
  id: number;
  name: string;
  description: string;
  node_id: number;
  process_name: string;
  schedule: string;
  action: 'start' | 'stop' | 'restart';
  enabled: boolean;
  last_run?: string;
  next_run?: string;
  created_at: string;
  updated_at: string;
}
