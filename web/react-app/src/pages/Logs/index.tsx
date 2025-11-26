import { useEffect, useState } from 'react';
import {
  Card,
  Input,
  Select,
  Button,
  Space,
  Table,
  Tag,
  DatePicker,
  message,
  Tabs,
} from 'antd';
import {
  SearchOutlined,
  ReloadOutlined,
  DownloadOutlined,
  FilterOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;
const { TextArea } = Input;

interface LogEntry {
  id: number;
  timestamp: string;
  level: 'debug' | 'info' | 'warning' | 'error';
  source: string;
  message: string;
  node_name?: string;
  process_name?: string;
  user?: string;
}

interface ActivityLog {
  id: number;
  timestamp: string;
  user: string;
  action: string;
  resource_type: string;
  resource_name: string;
  details: string;
  ip_address: string;
}

const Logs: React.FC = () => {
  const [systemLogs, setSystemLogs] = useState<LogEntry[]>([]);
  const [activityLogs, setActivityLogs] = useState<ActivityLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchText, setSearchText] = useState('');
  const [levelFilter, setLevelFilter] = useState<string>('all');
  const [sourceFilter, setSourceFilter] = useState<string>('all');

  useEffect(() => {
    loadLogs();
  }, []);

  const loadLogs = async () => {
    setLoading(true);
    try {
      // 模拟系统日志数据
      const mockSystemLogs: LogEntry[] = [
        {
          id: 1,
          timestamp: new Date().toISOString(),
          level: 'info',
          source: 'system',
          message: 'System started successfully',
        },
        {
          id: 2,
          timestamp: new Date(Date.now() - 60000).toISOString(),
          level: 'warning',
          source: 'node',
          message: 'Node connection timeout, retrying...',
          node_name: 'prod-server-01',
        },
        {
          id: 3,
          timestamp: new Date(Date.now() - 120000).toISOString(),
          level: 'error',
          source: 'process',
          message: 'Process crashed with exit code 1',
          node_name: 'prod-server-02',
          process_name: 'web-app',
        },
        {
          id: 4,
          timestamp: new Date(Date.now() - 180000).toISOString(),
          level: 'info',
          source: 'auth',
          message: 'User logged in successfully',
          user: 'admin',
        },
      ];

      // 模拟活动日志数据
      const mockActivityLogs: ActivityLog[] = [
        {
          id: 1,
          timestamp: new Date().toISOString(),
          user: 'admin',
          action: 'start_process',
          resource_type: 'process',
          resource_name: 'web-app',
          details: 'Started process on prod-server-01',
          ip_address: '192.168.1.100',
        },
        {
          id: 2,
          timestamp: new Date(Date.now() - 60000).toISOString(),
          user: 'john_doe',
          action: 'stop_process',
          resource_type: 'process',
          resource_name: 'worker',
          details: 'Stopped process on prod-server-02',
          ip_address: '192.168.1.101',
        },
        {
          id: 3,
          timestamp: new Date(Date.now() - 120000).toISOString(),
          user: 'admin',
          action: 'create_user',
          resource_type: 'user',
          resource_name: 'jane_smith',
          details: 'Created new user account',
          ip_address: '192.168.1.100',
        },
      ];

      setSystemLogs(mockSystemLogs);
      setActivityLogs(mockActivityLogs);
    } catch (error) {
      console.error('Failed to load logs:', error);
      message.error('Failed to load logs');
    } finally {
      setLoading(false);
    }
  };

  const handleExport = () => {
    message.success('Logs exported successfully');
  };

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'error': return 'red';
      case 'warning': return 'orange';
      case 'info': return 'blue';
      case 'debug': return 'default';
      default: return 'default';
    }
  };

  const getActionColor = (action: string) => {
    if (action.includes('create')) return 'green';
    if (action.includes('delete')) return 'red';
    if (action.includes('update') || action.includes('edit')) return 'orange';
    if (action.includes('start')) return 'blue';
    if (action.includes('stop')) return 'volcano';
    return 'default';
  };

  const systemLogColumns: ColumnsType<LogEntry> = [
    {
      title: 'Timestamp',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (timestamp: string) => new Date(timestamp).toLocaleString(),
      sorter: (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime(),
    },
    {
      title: 'Level',
      dataIndex: 'level',
      key: 'level',
      width: 100,
      render: (level: string) => (
        <Tag color={getLevelColor(level)}>
          {level.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: 'Source',
      dataIndex: 'source',
      key: 'source',
      width: 120,
    },
    {
      title: 'Node',
      dataIndex: 'node_name',
      key: 'node_name',
      width: 150,
      render: (nodeName?: string) => nodeName || '-',
    },
    {
      title: 'Process',
      dataIndex: 'process_name',
      key: 'process_name',
      width: 150,
      render: (processName?: string) => processName || '-',
    },
    {
      title: 'Message',
      dataIndex: 'message',
      key: 'message',
      ellipsis: true,
    },
  ];

  const activityLogColumns: ColumnsType<ActivityLog> = [
    {
      title: 'Timestamp',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (timestamp: string) => new Date(timestamp).toLocaleString(),
      sorter: (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime(),
    },
    {
      title: 'User',
      dataIndex: 'user',
      key: 'user',
      width: 120,
    },
    {
      title: 'Action',
      dataIndex: 'action',
      key: 'action',
      width: 150,
      render: (action: string) => (
        <Tag color={getActionColor(action)}>
          {action.replace(/_/g, ' ').toUpperCase()}
        </Tag>
      ),
    },
    {
      title: 'Resource Type',
      dataIndex: 'resource_type',
      key: 'resource_type',
      width: 120,
    },
    {
      title: 'Resource Name',
      dataIndex: 'resource_name',
      key: 'resource_name',
      width: 150,
    },
    {
      title: 'Details',
      dataIndex: 'details',
      key: 'details',
      ellipsis: true,
    },
    {
      title: 'IP Address',
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 140,
    },
  ];

  const filteredSystemLogs = systemLogs.filter(log => {
    const matchesSearch = 
      log.message.toLowerCase().includes(searchText.toLowerCase()) ||
      log.source.toLowerCase().includes(searchText.toLowerCase()) ||
      log.node_name?.toLowerCase().includes(searchText.toLowerCase()) ||
      log.process_name?.toLowerCase().includes(searchText.toLowerCase());
    
    const matchesLevel = levelFilter === 'all' || log.level === levelFilter;
    const matchesSource = sourceFilter === 'all' || log.source === sourceFilter;
    
    return matchesSearch && matchesLevel && matchesSource;
  });

  const filteredActivityLogs = activityLogs.filter(log =>
    log.user.toLowerCase().includes(searchText.toLowerCase()) ||
    log.action.toLowerCase().includes(searchText.toLowerCase()) ||
    log.resource_name.toLowerCase().includes(searchText.toLowerCase()) ||
    log.details.toLowerCase().includes(searchText.toLowerCase())
  );

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ margin: 0 }}>Logs</h1>
        <Space>
          <Button
            icon={<DownloadOutlined />}
            onClick={handleExport}
          >
            Export
          </Button>
          <Button
            type="primary"
            icon={<ReloadOutlined />}
            onClick={loadLogs}
            loading={loading}
          >
            Refresh
          </Button>
        </Space>
      </div>

      {/* 筛选器 */}
      <Card style={{ marginBottom: 16 }}>
        <Space wrap style={{ width: '100%' }}>
          <Input
            placeholder="Search logs..."
            prefix={<SearchOutlined />}
            style={{ width: 300 }}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            allowClear
          />
          <Select
            style={{ width: 150 }}
            value={levelFilter}
            onChange={setLevelFilter}
            options={[
              { label: 'All Levels', value: 'all' },
              { label: 'Debug', value: 'debug' },
              { label: 'Info', value: 'info' },
              { label: 'Warning', value: 'warning' },
              { label: 'Error', value: 'error' },
            ]}
          />
          <Select
            style={{ width: 150 }}
            value={sourceFilter}
            onChange={setSourceFilter}
            options={[
              { label: 'All Sources', value: 'all' },
              { label: 'System', value: 'system' },
              { label: 'Node', value: 'node' },
              { label: 'Process', value: 'process' },
              { label: 'Auth', value: 'auth' },
            ]}
          />
          <RangePicker
            showTime
            format="YYYY-MM-DD HH:mm"
            placeholder={['Start Time', 'End Time']}
          />
        </Space>
      </Card>

      {/* 日志表格 */}
      <Card>
        <Tabs
          items={[
            {
              key: 'system',
              label: `System Logs (${filteredSystemLogs.length})`,
              children: (
                <Table
                  columns={systemLogColumns}
                  dataSource={filteredSystemLogs}
                  rowKey="id"
                  loading={loading}
                  pagination={{
                    pageSize: 20,
                    showSizeChanger: true,
                    showQuickJumper: true,
                    showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} logs`,
                  }}
                  scroll={{ x: 1000 }}
                  size="small"
                />
              ),
            },
            {
              key: 'activity',
              label: `Activity Logs (${filteredActivityLogs.length})`,
              children: (
                <Table
                  columns={activityLogColumns}
                  dataSource={filteredActivityLogs}
                  rowKey="id"
                  loading={loading}
                  pagination={{
                    pageSize: 20,
                    showSizeChanger: true,
                    showQuickJumper: true,
                    showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} logs`,
                  }}
                  scroll={{ x: 1000 }}
                  size="small"
                />
              ),
            },
          ]}
        />
      </Card>
    </div>
  );
};

export default Logs;
