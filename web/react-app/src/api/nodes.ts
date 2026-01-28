import apiClient from './client';
import { Node, Process } from '@/types';

// 后端实际返回的响应格式
interface NodesResponse {
  status: string;
  nodes: Node[];
}

interface NodeResponse {
  name: string;
  environment: string;
  host: string;
  port: number;
  is_connected: boolean;
  process_count: number;
}

interface ProcessesResponse {
  status: string;
  processes: Process[];
}

interface LogStreamResponse {
  status: string;
  data: {
    process_name: string;
    node_name: string;
    log_type: string;
    entries: Array<{
      timestamp: string;
      level: string;
      message: string;
      source: string;
      process_name: string;
      node_name: string;
    }>;
    last_offset: number;
    overflow: boolean;
  };
}

export const nodesApi = {
  // Get all nodes
  getNodes: () => apiClient.get<NodesResponse>('/nodes'),

  // Get single node
  getNode: (nodeName: string) => apiClient.get<NodeResponse>(`/nodes/${nodeName}`),

  // Get node processes
  getNodeProcesses: (nodeName: string) =>
    apiClient.get<ProcessesResponse>(`/nodes/${nodeName}/processes`),

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
    apiClient.get<LogStreamResponse>(`/nodes/${nodeName}/processes/${processName}/logs/stream`, {
      params: { 
        offset: offset === undefined ? -1 : offset, 
        max_lines: maxLines || 50 
      }
    }),
};
