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
  Alert,
} from 'antd';
import {
  SearchOutlined,
  ReloadOutlined,
  DownloadOutlined,
} from '@ant-design/icons';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';
import dayjs, { Dayjs } from 'dayjs';
import { activityLogsAPI } from '../../api/activityLogs';
import type { ActivityLog, ActivityLogsFilters, PaginationInfo } from '../../types';

const { RangePicker } = DatePicker;

const Logs: React.FC = () => {
  const [activityLogs, setActivityLogs] = useState<ActivityLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchText, setSearchText] = useState('');
  const [levelFilter, setLevelFilter] = useState<string>('');
  const [actionFilter, setActionFilter] = useState<string>('');
  const [resourceFilter, setResourceFilter] = useState<string>('');
  const [usernameFilter, setUsernameFilter] = useState<string>('');
  const [dateRange, setDateRange] = useState<[Dayjs | null, Dayjs | null] | null>(null);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [pagination, setPagination] = useState<PaginationInfo>({
    page: 1,
    page_size: 20,
    total: 0,
    total_pages: 0,
    has_next: false,
    has_prev: false,
  });

  useEffect(() => {
    loadLogs();
  }, []);

  // 自动刷新功能
  useEffect(() => {
    if (!autoRefresh || pagination.page !== 1) {
      return;
    }

    const intervalId = setInterval(() => {
      loadLogs();
    }, 30000); // 30 秒刷新一次

    return () => clearInterval(intervalId);
  }, [autoRefresh, pagination.page]);

  const buildFilters = (): ActivityLogsFilters => {
    const filters: ActivityLogsFilters = {
      page: pagination.page,
      page_size: pagination.page_size,
    };

    if (levelFilter) filters.level = levelFilter;
    if (actionFilter) filters.action = actionFilter;
    if (resourceFilter) filters.resource = resourceFilter;
    if (usernameFilter) filters.username = usernameFilter;
    
    if (dateRange && dateRange[0] && dateRange[1]) {
      filters.start_time = dateRange[0].toISOString();
      filters.end_time = dateRange[1].toISOString();
    }

    return filters;
  };

  const loadLogs = async (customFilters?: Partial<ActivityLogsFilters>) => {
    setLoading(true);
    setError(null);
    
    try {
      const filters = customFilters ? { ...buildFilters(), ...customFilters } : buildFilters();
      const response = await activityLogsAPI.getActivityLogs(filters);
      
      setActivityLogs(response.data.logs);
      setPagination(response.data.pagination);
    } catch (err) {
      console.error('Failed to load logs:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to load activity logs';
      setError(errorMessage);
      message.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = () => {
    setPagination(prev => ({ ...prev, page: 1 }));
    loadLogs({ page: 1 });
  };

  const handleExport = async () => {
    try {
      message.loading({ content: 'Exporting logs...', key: 'export' });
      const filters = buildFilters();
      const blob = await activityLogsAPI.exportLogs(filters);
      
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `activity-logs-${dayjs().format('YYYY-MM-DD-HHmmss')}.csv`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
      
      message.success({ content: 'Logs exported successfully', key: 'export' });
    } catch (err) {
      console.error('Failed to export logs:', err);
      message.error({ content: 'Failed to export logs', key: 'export' });
    }
  };

  const getLevelColor = (level: string) => {
    const upperLevel = level.toUpperCase();
    switch (upperLevel) {
      case 'ERROR': return 'red';
      case 'WARNING': return 'orange';
      case 'INFO': return 'blue';
      case 'DEBUG': return 'default';
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

  const handleTableChange = (newPagination: TablePaginationConfig) => {
    const newPage = newPagination.current || 1;
    const newPageSize = newPagination.pageSize || 20;
    
    setPagination(prev => ({
      ...prev,
      page: newPage,
      page_size: newPageSize,
    }));
    
    loadLogs({ page: newPage, page_size: newPageSize });
  };

  const activityLogColumns: ColumnsType<ActivityLog> = [
    {
      title: 'Timestamp',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (timestamp: string) => dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss'),
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
      title: 'User',
      dataIndex: 'username',
      key: 'username',
      width: 120,
      render: (username: string) => username || 'system',
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
      title: 'Resource',
      dataIndex: 'resource',
      key: 'resource',
      width: 120,
    },
    {
      title: 'Target',
      dataIndex: 'target',
      key: 'target',
      width: 150,
      ellipsis: true,
    },
    {
      title: 'Message',
      dataIndex: 'message',
      key: 'message',
      ellipsis: true,
    },
    {
      title: 'IP Address',
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 140,
    },
  ];

  // Client-side search filter for display purposes
  const displayedLogs = searchText
    ? activityLogs.filter(log =>
        log.message.toLowerCase().includes(searchText.toLowerCase()) ||
        log.username?.toLowerCase().includes(searchText.toLowerCase()) ||
        log.action.toLowerCase().includes(searchText.toLowerCase()) ||
        log.target?.toLowerCase().includes(searchText.toLowerCase())
      )
    : activityLogs;

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ margin: 0 }}>Activity Logs</h1>
        <Space>
          <Button
            type={autoRefresh ? 'default' : 'dashed'}
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            Auto Refresh: {autoRefresh ? 'ON' : 'OFF'}
          </Button>
          <Button
            icon={<DownloadOutlined />}
            onClick={handleExport}
            disabled={loading}
          >
            Export
          </Button>
          <Button
            type="primary"
            icon={<ReloadOutlined />}
            onClick={() => loadLogs()}
            loading={loading}
          >
            Refresh
          </Button>
        </Space>
      </div>

      {error && (
        <Alert
          message="Error"
          description={error}
          type="error"
          closable
          onClose={() => setError(null)}
          style={{ marginBottom: 16 }}
        />
      )}

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
            placeholder="Level"
            value={levelFilter || undefined}
            onChange={setLevelFilter}
            allowClear
            options={[
              { label: 'INFO', value: 'INFO' },
              { label: 'WARNING', value: 'WARNING' },
              { label: 'ERROR', value: 'ERROR' },
              { label: 'DEBUG', value: 'DEBUG' },
            ]}
          />
          <Select
            style={{ width: 150 }}
            placeholder="Action"
            value={actionFilter || undefined}
            onChange={setActionFilter}
            allowClear
            options={[
              { label: 'Start Process', value: 'start_process' },
              { label: 'Stop Process', value: 'stop_process' },
              { label: 'Restart Process', value: 'restart_process' },
              { label: 'Login', value: 'login' },
              { label: 'Logout', value: 'logout' },
            ]}
          />
          <Select
            style={{ width: 150 }}
            placeholder="Resource"
            value={resourceFilter || undefined}
            onChange={setResourceFilter}
            allowClear
            options={[
              { label: 'Process', value: 'process' },
              { label: 'Node', value: 'node' },
              { label: 'User', value: 'user' },
              { label: 'Auth', value: 'auth' },
              { label: 'System', value: 'system' },
            ]}
          />
          <Input
            placeholder="Username"
            style={{ width: 150 }}
            value={usernameFilter}
            onChange={(e) => setUsernameFilter(e.target.value)}
            allowClear
          />
          <RangePicker
            showTime
            format="YYYY-MM-DD HH:mm"
            placeholder={['Start Time', 'End Time']}
            value={dateRange}
            onChange={setDateRange}
          />
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={handleSearch}
            loading={loading}
          >
            Search
          </Button>
        </Space>
      </Card>

      {/* 日志表格 */}
      <Card>
        <Table
          columns={activityLogColumns}
          dataSource={displayedLogs}
          rowKey="id"
          loading={loading}
          pagination={{
            current: pagination.page,
            pageSize: pagination.page_size,
            total: pagination.total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} logs`,
            pageSizeOptions: ['10', '20', '50', '100'],
          }}
          onChange={handleTableChange}
          scroll={{ x: 1200 }}
          size="small"
        />
      </Card>
    </div>
  );
};

export default Logs;
