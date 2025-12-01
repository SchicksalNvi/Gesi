import apiClient from './client';

export interface ActivityLog {
  id: number;
  created_at: string;
  level: string;
  message: string;
  action: string;
  resource: string;
  target: string;
  user_id: string;
  username: string;
  ip_address: string;
  user_agent: string;
  details: string;
  status: string;
  duration: number;
}

export interface ActivityLogsFilters {
  level?: string;
  action?: string;
  resource?: string;
  username?: string;
  start_time?: string;
  end_time?: string;
  page?: number;
  page_size?: number;
}

export interface PaginationInfo {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface ActivityLogsResponse {
  status: string;
  data: {
    logs: ActivityLog[];
    pagination: PaginationInfo;
  };
}

export interface LogStatistics {
  total_logs: number;
  info_count: number;
  warning_count: number;
  error_count: number;
  debug_count: number;
  top_actions: Array<{ action: string; count: number }>;
  top_users: Array<{ username: string; count: number }>;
}

export interface LogStatisticsResponse {
  status: string;
  data: LogStatistics;
}

export interface RecentLogsResponse {
  status: string;
  logs: ActivityLog[];
}

class ActivityLogsAPI {
  async getActivityLogs(filters: ActivityLogsFilters = {}): Promise<ActivityLogsResponse> {
    const params = new URLSearchParams();
    
    if (filters.level) params.append('level', filters.level);
    if (filters.action) params.append('action', filters.action);
    if (filters.resource) params.append('resource', filters.resource);
    if (filters.username) params.append('username', filters.username);
    if (filters.start_time) params.append('start_time', filters.start_time);
    if (filters.end_time) params.append('end_time', filters.end_time);
    if (filters.page) params.append('page', filters.page.toString());
    if (filters.page_size) params.append('page_size', filters.page_size.toString());

    const queryString = params.toString();
    const url = `/activity-logs${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<ActivityLogsResponse>(url);
  }

  async getRecentLogs(limit: number = 20): Promise<ActivityLog[]> {
    const response = await apiClient.get<RecentLogsResponse>(
      `/activity-logs/recent?limit=${limit}`
    );
    return response.logs;
  }

  async getLogStatistics(days: number = 7): Promise<LogStatistics> {
    const response = await apiClient.get<LogStatisticsResponse>(
      `/activity-logs/statistics?days=${days}`
    );
    return response.data;
  }

  async exportLogs(filters: ActivityLogsFilters = {}): Promise<Blob> {
    const params = new URLSearchParams();
    
    if (filters.level) params.append('level', filters.level);
    if (filters.action) params.append('action', filters.action);
    if (filters.resource) params.append('resource', filters.resource);
    if (filters.username) params.append('username', filters.username);
    if (filters.start_time) params.append('start_time', filters.start_time);
    if (filters.end_time) params.append('end_time', filters.end_time);

    const queryString = params.toString();
    const url = `/activity-logs/export${queryString ? `?${queryString}` : ''}`;
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
      },
    });

    if (!response.ok) {
      throw new Error('Failed to export logs');
    }

    return response.blob();
  }
}

export const activityLogsAPI = new ActivityLogsAPI();
export default activityLogsAPI;
