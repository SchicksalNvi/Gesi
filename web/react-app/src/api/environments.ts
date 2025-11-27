import apiClient from './client';
import { ApiResponse, Environment, EnvironmentDetail } from '@/types';

export const environmentsApi = {
  // Get all environments
  getEnvironments: () =>
    apiClient.get<ApiResponse<{ environments: Environment[] }>>('/environments'),

  // Get environment detail
  getEnvironmentDetail: (name: string) =>
    apiClient.get<ApiResponse<{ environment: EnvironmentDetail }>>(`/environments/${name}`),
};
