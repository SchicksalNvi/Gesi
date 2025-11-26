import apiClient from './client';
import { ApiResponse, User } from '@/types';

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export const authApi = {
  // Login
  login: (data: LoginRequest) =>
    apiClient.post<ApiResponse<LoginResponse>>('/auth/login', data),

  // Logout
  logout: () => apiClient.post('/auth/logout'),

  // Get current user
  getCurrentUser: () => apiClient.get<ApiResponse<{ user: User }>>('/auth/user'),

  // Update profile
  updateProfile: (data: Partial<User>) => apiClient.put('/profile', data),
};
