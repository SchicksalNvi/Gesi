import apiClient from './client';
import { ApiResponse, Node, Process } from '@/types';

export const nodesApi = {
  // Get all nodes
  getNodes: () => apiClient.get<ApiResponse<{ nodes: Node[] }>>('/nodes'),

  // Get single node
  getNode: (nodeName: string) => apiClient.get<ApiResponse<{ node: Node }>>(`/nodes/${nodeName}`),

  // Get node processes
  getNodeProcesses: (nodeName: string) =>
    apiClient.get<ApiResponse<{ processes: Process[] }>>(`/nodes/${nodeName}/processes`),

  // Process control
  startProcess: (nodeName: string, processName: string) =>
    apiClient.post(`/nodes/${nodeName}/processes/${processName}/start`),

  stopProcess: (nodeName: string, processName: string) =>
    apiClient.post(`/nodes/${nodeName}/processes/${processName}/stop`),

  restartProcess: (nodeName: string, processName: string) =>
    apiClient.post(`/nodes/${nodeName}/processes/${processName}/restart`),

  // Batch operations
  startAllProcesses: (nodeName: string) =>
    apiClient.post(`/nodes/${nodeName}/processes/start-all`),

  stopAllProcesses: (nodeName: string) =>
    apiClient.post(`/nodes/${nodeName}/processes/stop-all`),

  restartAllProcesses: (nodeName: string) =>
    apiClient.post(`/nodes/${nodeName}/processes/restart-all`),

  // Get process logs
  getProcessLogs: (nodeName: string, processName: string) =>
    apiClient.get(`/nodes/${nodeName}/processes/${processName}/logs`),

  // Get process log stream (structured logs with pagination)
  // offset: undefined/-1 = read from end, >= 0 = read from specific offset
  getProcessLogStream: (nodeName: string, processName: string, offset?: number, maxLines?: number) =>
    apiClient.get(`/nodes/${nodeName}/processes/${processName}/logs/stream`, {
      params: { 
        offset: offset === undefined ? -1 : offset, 
        max_lines: maxLines || 50 
      }
    }),
};
