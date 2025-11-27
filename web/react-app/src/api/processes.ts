import apiClient from './client';
import { AggregatedProcess, BatchOperationResult } from '@/types';

export const processesApi = {
  // Get aggregated processes
  getAggregated: () =>
    apiClient.get<{ status: string; processes: AggregatedProcess[] }>('/processes/aggregated'),

  // Batch operations
  batchStart: (processName: string) =>
    apiClient.post<{ status: string; result: BatchOperationResult }>(
      `/processes/${processName}/start`
    ),

  batchStop: (processName: string) =>
    apiClient.post<{ status: string; result: BatchOperationResult }>(
      `/processes/${processName}/stop`
    ),

  batchRestart: (processName: string) =>
    apiClient.post<{ status: string; result: BatchOperationResult }>(
      `/processes/${processName}/restart`
    ),
};
