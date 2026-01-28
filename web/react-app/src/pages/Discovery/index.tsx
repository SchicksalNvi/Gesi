import { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Form,
  Input,
  InputNumber,
  Button,
  Table,
  Progress,
  Tag,
  Space,
  message,
  Modal,
  Tooltip,
  Empty,
  Spin,
  Tabs,
  Badge,
  Typography,
  Popconfirm,
  Alert,
} from 'antd';
import type { TabsProps } from 'antd';
import {
  SearchOutlined,
  StopOutlined,
  DeleteOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  EyeOutlined,
  RadarChartOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import {
  discoveryApi,
  DiscoveryTask,
  DiscoveryResult,
  StartDiscoveryRequest,
} from '@/api/discovery';
import { useWebSocket } from '@/hooks/useWebSocket';
import { WebSocketMessage } from '@/types';

const { Text, Title } = Typography;

// CIDR validation regex
const CIDR_REGEX = /^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/;

// Status tag colors
const statusColors: Record<string, string> = {
  pending: 'default',
  running: 'processing',
  completed: 'success',
  cancelled: 'warning',
  failed: 'error',
};

const resultStatusColors: Record<string, string> = {
  success: 'success',
  timeout: 'warning',
  connection_refused: 'default',
  auth_failed: 'error',
  error: 'error',
};

interface DiscoveryProgress {
  task_id: number;
  scanned_ips: number;
  total_ips: number;
  found_nodes: number;
  failed_ips: number;
  percent: number;
}

interface DiscoveredNode {
  task_id: number;
  ip: string;
  port: number;
  node_name: string;
  version: string;
}

// Helper to extract error message from various error formats
const getErrorMessage = (error: any, fallback: string): string => {
  if (!error) return fallback;
  const data = error.response?.data;
  if (!data) return fallback;
  if (typeof data === 'string') return data;
  if (typeof data.error === 'string') return data.error;
  if (typeof data.message === 'string') return data.message;
  if (data.error?.message) return data.error.message;
  return fallback;
};

export default function DiscoveryPage() {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [tasks, setTasks] = useState<DiscoveryTask[]>([]);
  const [tasksLoading, setTasksLoading] = useState(false);
  const [pagination, setPagination] = useState({ page: 1, limit: 10, total: 0 });
  const [activeTask, setActiveTask] = useState<DiscoveryTask | null>(null);
  const [activeResults, setActiveResults] = useState<DiscoveryResult[]>([]);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [cidrValidation, setCidrValidation] = useState<{ valid: boolean; count: number; error?: string } | null>(null);
  const [validatingCidr, setValidatingCidr] = useState(false);
  const [recentlyDiscovered, setRecentlyDiscovered] = useState<DiscoveredNode[]>([]);
  const [activeTab, setActiveTab] = useState('new');

  // Load tasks on mount
  useEffect(() => {
    loadTasks();
  }, [pagination.page, pagination.limit]);

  // WebSocket for real-time updates
  useWebSocket({
    onMessage: useCallback((msg: WebSocketMessage) => {
      if (msg.type === 'discovery_progress') {
        const progress = msg.data as DiscoveryProgress;
        setTasks(prev => prev.map(t => 
          t.id === progress.task_id 
            ? { ...t, scanned_ips: progress.scanned_ips, found_nodes: progress.found_nodes, failed_ips: progress.failed_ips }
            : t
        ));
        if (activeTask?.id === progress.task_id) {
          setActiveTask(prev => prev ? { ...prev, scanned_ips: progress.scanned_ips, found_nodes: progress.found_nodes, failed_ips: progress.failed_ips } : null);
        }
      } else if (msg.type === 'node_discovered') {
        const node = msg.data as DiscoveredNode;
        setRecentlyDiscovered(prev => [node, ...prev.slice(0, 9)]);
        message.success(`Discovered node: ${node.node_name}`);
      } else if (msg.type === 'discovery_completed') {
        const data = msg.data as { task_id: number; status: string };
        setTasks(prev => prev.map(t => 
          t.id === data.task_id ? { ...t, status: data.status as any } : t
        ));
        if (activeTask?.id === data.task_id) {
          loadTaskDetail(data.task_id);
        }
        message.info(`Discovery task ${data.task_id} completed`);
      }
    }, [activeTask]),
  });

  const loadTasks = async () => {
    setTasksLoading(true);
    try {
      const response = await discoveryApi.getTasks({ page: pagination.page, limit: pagination.limit });
      setTasks(response.tasks || []);
      setPagination(prev => ({ ...prev, total: response.total }));
    } catch (error) {
      console.error('Failed to load tasks:', error);
      message.error('Failed to load discovery tasks');
    } finally {
      setTasksLoading(false);
    }
  };

  const loadTaskDetail = async (taskId: number) => {
    setDetailLoading(true);
    try {
      const response = await discoveryApi.getTask(taskId);
      setActiveTask(response.task);
      setActiveResults(response.results || []);
    } catch (error) {
      console.error('Failed to load task detail:', error);
      message.error('Failed to load task details');
    } finally {
      setDetailLoading(false);
    }
  };

  // CIDR validation with debounce
  const validateCidr = async (cidr: string) => {
    if (!cidr) {
      setCidrValidation(null);
      return;
    }
    if (!CIDR_REGEX.test(cidr)) {
      setCidrValidation({ valid: false, count: 0, error: 'Invalid CIDR format (e.g., 192.168.1.0/24)' });
      return;
    }
    setValidatingCidr(true);
    try {
      const response = await discoveryApi.validateCIDR(cidr);
      setCidrValidation({ valid: response.valid, count: response.count });
    } catch (error: any) {
      const errorMsg = getErrorMessage(error, 'Failed to validate CIDR');
      setCidrValidation({ valid: false, count: 0, error: errorMsg });
    } finally {
      setValidatingCidr(false);
    }
  };

  // Start discovery
  const handleStartDiscovery = async (values: StartDiscoveryRequest) => {
    if (!cidrValidation?.valid) {
      message.error('Please enter a valid CIDR range');
      return;
    }
    setLoading(true);
    try {
      await discoveryApi.startDiscovery(values);
      message.success(`Discovery started for ${values.cidr}`);
      form.resetFields();
      setCidrValidation(null);
      setRecentlyDiscovered([]);
      setActiveTab('history');
      loadTasks();
    } catch (error: any) {
      const errorMsg = getErrorMessage(error, 'Failed to start discovery');
      message.error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  // Cancel task
  const handleCancelTask = async (taskId: number) => {
    try {
      await discoveryApi.cancelTask(taskId);
      message.success('Task cancelled');
      loadTasks();
    } catch (error) {
      message.error('Failed to cancel task');
    }
  };

  // Delete task
  const handleDeleteTask = async (taskId: number) => {
    try {
      await discoveryApi.deleteTask(taskId);
      message.success('Task deleted');
      loadTasks();
    } catch (error) {
      message.error('Failed to delete task');
    }
  };

  // View task details
  const handleViewTask = (task: DiscoveryTask) => {
    setActiveTask(task);
    setDetailModalVisible(true);
    loadTaskDetail(task.id);
  };

  // Task table columns
  const taskColumns: ColumnsType<DiscoveryTask> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: 'CIDR',
      dataIndex: 'cidr',
      key: 'cidr',
      render: (cidr: string) => <Text code>{cidr}</Text>,
    },
    {
      title: 'Port',
      dataIndex: 'port',
      key: 'port',
      width: 80,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string) => (
        <Tag color={statusColors[status] || 'default'}>
          {status.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: 'Progress',
      key: 'progress',
      width: 200,
      render: (_, record) => {
        const percent = record.total_ips > 0 ? Math.round((record.scanned_ips / record.total_ips) * 100) : 0;
        return (
          <Space direction="vertical" size={0} style={{ width: '100%' }}>
            <Progress percent={percent} size="small" status={record.status === 'running' ? 'active' : undefined} />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {record.scanned_ips}/{record.total_ips} IPs • {record.found_nodes} found
            </Text>
          </Space>
        );
      },
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 150,
      render: (_, record) => (
        <Space>
          <Tooltip title="View Details">
            <Button type="text" icon={<EyeOutlined />} onClick={() => handleViewTask(record)} />
          </Tooltip>
          {record.status === 'running' && (
            <Tooltip title="Cancel">
              <Button type="text" danger icon={<StopOutlined />} onClick={() => handleCancelTask(record.id)} />
            </Tooltip>
          )}
          {['completed', 'cancelled', 'failed'].includes(record.status) && (
            <Popconfirm title="Delete this task?" onConfirm={() => handleDeleteTask(record.id)}>
              <Tooltip title="Delete">
                <Button type="text" danger icon={<DeleteOutlined />} />
              </Tooltip>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  // Result table columns
  const resultColumns: ColumnsType<DiscoveryResult> = [
    {
      title: 'IP',
      dataIndex: 'ip',
      key: 'ip',
      render: (ip: string) => <Text code>{ip}</Text>,
    },
    {
      title: 'Port',
      dataIndex: 'port',
      key: 'port',
      width: 80,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 160,
      render: (status: string, record: DiscoveryResult) => {
        const tag = (
          <Tag color={resultStatusColors[status] || 'default'}>
            {status === 'success' && <CheckCircleOutlined />}
            {status === 'timeout' && <ClockCircleOutlined />}
            {status === 'connection_refused' && <CloseCircleOutlined />}
            {status === 'auth_failed' && <ExclamationCircleOutlined />}
            {status === 'error' && <ExclamationCircleOutlined />}
            {' '}{status.replace('_', ' ').toUpperCase()}
          </Tag>
        );
        // Show error message in tooltip for failed statuses
        if (record.error_msg && status !== 'success') {
          return <Tooltip title={record.error_msg}>{tag}</Tooltip>;
        }
        return tag;
      },
    },
    {
      title: 'Node Name',
      dataIndex: 'node_name',
      key: 'node_name',
      render: (name: string) => name || '-',
    },
    {
      title: 'Version',
      dataIndex: 'version',
      key: 'version',
      render: (version: string) => version || '-',
    },
    {
      title: 'Duration',
      dataIndex: 'duration_ms',
      key: 'duration_ms',
      width: 100,
      render: (ms: number) => `${ms}ms`,
    },
  ];

  // New Discovery tab content
  const newDiscoveryContent = (
    <>
      <Card title="Start Network Discovery" style={{ marginBottom: 24 }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleStartDiscovery}
          initialValues={{ port: 9001, timeout_seconds: 3, max_workers: 50 }}
        >
          <Form.Item
            name="cidr"
            label="CIDR Range"
            rules={[{ required: true, message: 'Please enter a CIDR range' }]}
            help={
              cidrValidation ? (
                cidrValidation.valid ? (
                  <Text type="success">Valid CIDR: {cidrValidation.count} IP addresses</Text>
                ) : (
                  <Text type="danger">{cidrValidation.error || 'Invalid CIDR'}</Text>
                )
              ) : undefined
            }
            validateStatus={cidrValidation ? (cidrValidation.valid ? 'success' : 'error') : undefined}
          >
            <Input
              placeholder="e.g., 192.168.1.0/24"
              onChange={(e) => validateCidr(e.target.value)}
              suffix={validatingCidr ? <Spin size="small" /> : <span />}
            />
          </Form.Item>

          <Space size="large" style={{ width: '100%' }}>
            <Form.Item
              name="port"
              label="Supervisor Port"
              rules={[{ required: true, message: 'Please enter port' }]}
              style={{ width: 150 }}
            >
              <InputNumber min={1} max={65535} style={{ width: '100%' }} />
            </Form.Item>

            <Form.Item
              name="username"
              label="Username"
              rules={[{ required: true, message: 'Please enter username' }]}
              style={{ width: 200 }}
            >
              <Input placeholder="admin" />
            </Form.Item>

            <Form.Item
              name="password"
              label="Password"
              rules={[{ required: true, message: 'Please enter password' }]}
              style={{ width: 200 }}
            >
              <Input.Password placeholder="Password" />
            </Form.Item>
          </Space>

          <Space size="large">
            <Form.Item name="timeout_seconds" label="Timeout (seconds)" style={{ width: 150 }}>
              <InputNumber min={1} max={30} style={{ width: '100%' }} />
            </Form.Item>

            <Form.Item name="max_workers" label="Max Workers" style={{ width: 150 }}>
              <InputNumber min={1} max={200} style={{ width: '100%' }} />
            </Form.Item>
          </Space>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              icon={<SearchOutlined />}
              disabled={!cidrValidation?.valid}
            >
              Start Discovery
            </Button>
          </Form.Item>
        </Form>
      </Card>

      {recentlyDiscovered.length > 0 && (
        <Card title="Recently Discovered Nodes" size="small">
          <Space direction="vertical" style={{ width: '100%' }}>
            {recentlyDiscovered.map((node, index) => (
              <Alert
                key={`${node.ip}-${index}`}
                type="success"
                message={
                  <Space>
                    <CheckCircleOutlined />
                    <Text strong>{node.node_name}</Text>
                    <Text type="secondary">({node.ip}:{node.port})</Text>
                    <Tag color="blue">v{node.version}</Tag>
                  </Space>
                }
                showIcon={false}
              />
            ))}
          </Space>
        </Card>
      )}
    </>
  );

  // History tab content
  const historyContent = (
    <Card
      title="Discovery Tasks"
      extra={
        <Button icon={<ReloadOutlined />} onClick={loadTasks} loading={tasksLoading}>
          Refresh
        </Button>
      }
    >
      <Table
        columns={taskColumns}
        dataSource={tasks}
        rowKey="id"
        loading={tasksLoading}
        pagination={{
          current: pagination.page,
          pageSize: pagination.limit,
          total: pagination.total,
          showSizeChanger: true,
          showTotal: (total) => `Total ${total} tasks`,
          onChange: (page, pageSize) => setPagination({ page, limit: pageSize, total: pagination.total }),
        }}
        locale={{
          emptyText: <Empty description="No discovery tasks yet" />,
        }}
      />
    </Card>
  );

  // Tab items (new API)
  const tabItems: TabsProps['items'] = [
    {
      key: 'new',
      label: 'New Discovery',
      children: newDiscoveryContent,
    },
    {
      key: 'history',
      label: (
        <Badge count={tasks.filter(t => t.status === 'running').length} offset={[10, 0]}>
          Discovery History
        </Badge>
      ),
      children: historyContent,
    },
  ];

  return (
    <div>
      <Title level={4} style={{ marginBottom: 24 }}>
        <RadarChartOutlined /> Node Discovery
      </Title>

      <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />

      {/* Task Detail Modal */}
      <Modal
        title={`Discovery Task #${activeTask?.id}`}
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={null}
        width={900}
      >
        {detailLoading ? (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Spin size="large" />
          </div>
        ) : activeTask ? (
          <div>
            <Card size="small" style={{ marginBottom: 16 }}>
              <Space size="large" wrap>
                <div>
                  <Text type="secondary">CIDR:</Text> <Text code>{activeTask.cidr}</Text>
                </div>
                <div>
                  <Text type="secondary">Port:</Text> <Text>{activeTask.port}</Text>
                </div>
                <div>
                  <Text type="secondary">Status:</Text>{' '}
                  <Tag color={statusColors[activeTask.status]}>{activeTask.status.toUpperCase()}</Tag>
                </div>
                <div>
                  <Text type="secondary">Created by:</Text> <Text>{activeTask.created_by}</Text>
                </div>
              </Space>
            </Card>

            <Card size="small" style={{ marginBottom: 16 }}>
              <Progress
                percent={activeTask.total_ips > 0 ? Math.round((activeTask.scanned_ips / activeTask.total_ips) * 100) : 0}
                status={activeTask.status === 'running' ? 'active' : undefined}
              />
              <Space size="large" style={{ marginTop: 8 }}>
                <Text>
                  <Text type="secondary">Scanned:</Text> {activeTask.scanned_ips}/{activeTask.total_ips}
                </Text>
                <Text type="success">
                  <CheckCircleOutlined /> Found: {activeTask.found_nodes}
                </Text>
                <Text type="danger">
                  <CloseCircleOutlined /> Failed: {activeTask.failed_ips}
                </Text>
              </Space>
            </Card>

            <Table
              columns={resultColumns}
              dataSource={activeResults}
              rowKey="id"
              size="small"
              pagination={{ pageSize: 10 }}
              locale={{
                emptyText: <Empty description="No results yet" />,
              }}
            />
          </div>
        ) : null}
      </Modal>
    </div>
  );
}
