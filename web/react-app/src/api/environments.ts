import apiClient from './client';
import { Environment, EnvironmentDetail } from '@/types';

// 后端实际返回的响应格式
interface EnvironmentsResponse {
  status: string;
  environments: Environment[];
}

interface EnvironmentDetailResponse {
  status: string;
  environment: EnvironmentDetail;
}

export const environmentsApi = {
  // Get all environments
  getEnvironments: () =>
    apiClient.get<EnvironmentsResponse>('/environments'),

  // Get environment detail
  getEnvironmentDetail: (name: string) =>
    apiClient.get<EnvironmentDetailResponse>(`/environments/${name}`),
};
