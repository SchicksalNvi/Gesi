import axios from 'axios';

// Create axios instance with default config
const api = axios.create({
  baseURL: 'http://localhost:8081/api',
  timeout: 10000,
  withCredentials: true, // Enable cookies
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor (cookies are automatically included with withCredentials: true)
api.interceptors.request.use(
  (config) => {
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor to handle errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // 只有当前不在登录页面时才重定向，避免无限循环
      if (!window.location.pathname.includes('/login')) {
        // Cookie will be cleared by server, just redirect to login
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

// Auth API
export const authAPI = {
  login: (credentials) => api.post('/auth/login', credentials),
  logout: () => api.post('/auth/logout'),
  getCurrentUser: () => api.get('/auth/user'),
};

// Nodes API
export const nodesAPI = {
  getNodes: () => api.get('/nodes'),
  getNode: (nodeName) => api.get(`/nodes/${nodeName}`),
  getNodeProcesses: (nodeName) => api.get(`/nodes/${nodeName}/processes`),
  startProcess: (nodeName, processName) => api.post(`/nodes/${nodeName}/processes/${processName}/start`),
  stopProcess: (nodeName, processName) => api.post(`/nodes/${nodeName}/processes/${processName}/stop`),
  restartProcess: (nodeName, processName) => api.post(`/nodes/${nodeName}/processes/${processName}/restart`),
  getProcessLogs: (nodeName, processName) => api.get(`/nodes/${nodeName}/processes/${processName}/logs`),
  startAllProcesses: (nodeName) => api.post(`/nodes/${nodeName}/processes/start-all`),
  stopAllProcesses: (nodeName) => api.post(`/nodes/${nodeName}/processes/stop-all`),
  restartAllProcesses: (nodeName) => api.post(`/nodes/${nodeName}/processes/restart-all`),
};

// Users API
export const usersAPI = {
  getUsers: () => api.get('/users'),
  createUser: (userData) => api.post('/users', userData),
  deleteUser: (username) => api.delete(`/users/${username}`),
  changePassword: (username, passwordData) => api.put(`/users/${username}/password`, passwordData),
};

// Profile API
export const profileAPI = {
  getProfile: () => api.get('/profile'),
  updateProfile: (profileData) => api.put('/profile', profileData),
};

// Environments API
export const environmentsAPI = {
  getEnvironments: () => api.get('/environments'),
  getEnvironmentDetails: (environmentName) => api.get(`/environments/${environmentName}`),
};

// Groups API
export const groupsAPI = {
  getGroups: () => api.get('/groups'),
  getGroupDetails: (groupName) => api.get(`/groups/${groupName}`),
  startGroupProcesses: (groupName) => api.post(`/groups/${groupName}/start`),
  stopGroupProcesses: (groupName) => api.post(`/groups/${groupName}/stop`),
  restartGroupProcesses: (groupName) => api.post(`/groups/${groupName}/restart`),
};

// Activity Logs API
export const activityLogsAPI = {
  getActivityLogs: (params) => api.get('/activity-logs', { params }),
  getRecentLogs: () => api.get('/activity-logs/recent'),
  getLogStatistics: () => api.get('/activity-logs/statistics'),
  cleanOldLogs: () => api.delete('/activity-logs/clean'),
};

// Health API
export const healthAPI = {
  getHealth: () => api.get('/health'),
};

// Data Management API
export const dataManagementAPI = {
  // Export functionality
  exportData: (exportData) => api.post('/data-management/export', exportData),
  getExportRecords: () => api.get('/data-management/exports'),
  downloadExportFile: (recordId) => api.get(`/data-management/exports/${recordId}/download`, { responseType: 'blob' }),
  deleteExportRecord: (recordId) => api.delete(`/data-management/exports/${recordId}`),
  
  // Backup functionality
  createBackup: (backupData) => api.post('/data-management/backup', backupData),
  getBackupRecords: () => api.get('/data-management/backups'),
  downloadBackupFile: (recordId) => api.get(`/data-management/backups/${recordId}/download`, { responseType: 'blob' }),
  deleteBackupRecord: (recordId) => api.delete(`/data-management/backups/${recordId}`),
  
  // Import functionality
  importData: (formData) => api.post('/data-management/import', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  }),
};

// System Settings API
export const systemSettingsAPI = {
  getSystemSettings: () => api.get('/system-settings'),
  getSystemSetting: (key) => api.get(`/system-settings/${key}`),
  updateSystemSetting: (key, value) => api.put(`/system-settings/${key}`, { value }),
  updateMultipleSettings: (settings) => api.put('/system-settings/batch', settings),
  deleteSystemSetting: (key) => api.delete(`/system-settings/${key}`),
  getUserPreferences: () => api.get('/system-settings/user-preferences'),
  updateUserPreferences: (preferences) => api.put('/system-settings/user-preferences', preferences),
  testEmailConfiguration: (testData) => api.post('/system-settings/test-email', testData),
  resetToDefaults: (category) => api.post('/system-settings/reset', { category }),
};

// Developer Tools API
export const developerToolsAPI = {
  // API Documentation
  getApiEndpoints: () => api.get('/developer/api-docs'),
  testApiEndpoint: (endpoint, requestData) => api.post('/developer/test-api', {
    endpoint,
    ...requestData
  }),
  
  // Debug Tools
  getDebugLogs: (params) => api.get('/developer/debug-logs', { params }),
  clearDebugLogs: () => api.delete('/developer/debug-logs'),
  setLogLevel: (level) => api.put('/developer/log-level', { level }),
  
  // Performance Monitoring
  getPerformanceMetrics: () => api.get('/developer/performance'),
  getSystemMetrics: () => api.get('/developer/system-metrics'),
  getApiMetrics: () => api.get('/developer/api-metrics'),
  getDatabaseMetrics: () => api.get('/developer/database-metrics'),
  getWebSocketMetrics: () => api.get('/developer/websocket-metrics'),
};

export default api;