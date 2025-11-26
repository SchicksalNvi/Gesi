import { useEffect, useState } from 'react';
import {
  Table,
  Tag,
  Button,
  Space,
  Card,
  message,
  Input,
  Select,
  DatePicker,
  Badge,
} from 'antd';
import {
  ReloadOutlined,
  SearchOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  WarningOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';

const { Search } = Input;
const { RangePicker } = DatePicker;

interface Alert {
  id: number;
  node_name: string;
  process_name: string;
  alert_type: string;
  severity: 'info' | 'warning' | 'error' | 'critical';
  message: string;
  status: 'active' | 'acknowledged' | 'resolved';
  created_at: string;
  updated_at: string;
  acknowledged_by?: string;
  resolved_at?: string;
}

const Alerts: React.FC = () => {
  const navigate = useNavigate();
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchText, setSearchText] = useState('');
  const [severityFilter, setSeverityFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<string>('all');

  useEffect(() => {
    loadAlerts();
  }, []);

  const loadAlerts = async () => {
    setLoading(true);
    try {
      // 模拟数据 - 实际应该调用 API
      const mockAlerts: Alert[] = [
        {
          id: 1,
          node_name: 'prod-server-01',
          process_name: 'web-app',
          alert_type: 'process_stopped',
          severity: 'critical',
          message: 'Process web-app has stopped unexpectedly',
          status: 'active',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        {
          id: 2,
          node_name: 'prod-server-02',
          process_name: 'worker',
          alert_type: 'high_memory',
          severity: 'warning',
          message: 'Memory usage exceeded 80%',
          status: 'acknowledged',
          created_at: new Date(Date.now() - 3600000).toISOString(),
          updated_at: new Date().toISOString(),
          acknowledged_by: 'admin',
        },
        {
          id: 3,
          node_name: 'staging-server',
          process_name: 'api',
          alert_type: 'process_restart',
          severity: 'info',
          message: 'Process restarted successfully',
          status: 'resolved',
          created_at: new Date(Date.now() - 7200000).toISOString(),
          updated_at: new Date(Date.now() - 3600000).toISOString(),
          resolved_at: new Date(Date.now() - 3600000).toISOString(),
        },
      ];
      
      setAlerts(mockAlerts);
    } catch (error) {
      console.error('Failed to load alerts:', error);
      message.error('Failed to load alerts');
    } finally {
      setLoading(false);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'red';
      case 'error': return 'orange';
      case 'warning': return 'gold';
      case 'info': return 'blue';
      default: return 'default';
    }
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical': return <CloseCircleOutlined />;
      case 'error': return <CloseCircleOutlined />;
      case 'warning': return <WarningOutlined />;
      case 'info': return <InfoCircleOutlined />;
      default: return <InfoCircleOutlined />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'error';
      case 'acknowledged': return 'warning';
      case 'resolved': return 'success';
      default: return 'default';
    }
  };

  const handleAcknowledge = (alertId: number) => {
    message.success(`Alert ${alertId} acknowledged`);
    // 实际应该调用 API
    loadAlerts();
  };

  const handleResolve = (alertId: number) => {
    message.success(`Alert ${alertId} resolved`);
    // 实际应该调用 API
    loadAlerts();
  };

  const filteredAlerts = alerts.filter(alert => {
    const matchesSearch = 
      alert.node_name.toLowerCase().includes(searchText.toLowerCase()) ||
      alert.process_name.toLowerCase().includes(searchText.toLowerCase()) ||
      alert.message.toLowerCase().includes(searchText.toLowerCase());
    
    const matchesSeverity = severityFilter === 'all' || alert.severity === severityFilter;
    const matchesStatus = statusFilter === 'all' || alert.status === statusFilter;
    
    return matchesSearch && matchesSeverity && matchesStatus;
  });

  const columns: ColumnsType<Alert> = [
    {
      title: 'Severity',
      dataIndex: 'severity',
      key: 'severity',
      width: 120,
      render: (severity: string) => (
        <Tag icon={getSeverityIcon(severity)} color={getSeverityColor(severity)}>
          {severity.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: 'Node',
      dataIndex: 'node_name',
      key: 'node_name',
      width: 150,
    },
    {
      title: 'Process',
      dataIndex: 'process_name',
      key: 'process_name',
      width: 150,
    },
    {
      title: 'Type',
      dataIndex: 'alert_type',
      key: 'alert_type',
      width: 150,
      render: (type: string) => type.replace(/_/g, ' ').toUpperCase(),
    },
    {
      title: 'Message',
      dataIndex: 'message',
      key: 'message',
      ellipsis: true,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {status.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 200,
      render: (_, record: Alert) => (
        <Space>
          {record.status === 'active' && (
            <Button
              size="small"
              onClick={() => handleAcknowledge(record.id)}
            >
              Acknowledge
            </Button>
          )}
          {record.status !== 'resolved' && (
            <Button
              size="small"
              type="primary"
              onClick={() => handleResolve(record.id)}
            >
              Resolve
            </Button>
          )}
        </Space>
      ),
    },
  ];

  const activeCount = alerts.filter(a => a.status === 'active').length;
  const acknowledgedCount = alerts.filter(a => a.status === 'acknowledged').length;
  const resolvedCount = alerts.filter(a => a.status === 'resolved').length;

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <h1 style={{ margin: 0 }}>Alerts</h1>
          <Badge count={activeCount} style={{ backgroundColor: '#ff4d4f' }} />
        </Space>
        <Space>
          <Button
            onClick={() => navigate('/alerts/rules')}
          >
            Alert Rules
          </Button>
          <Button
            type="primary"
            icon={<ReloadOutlined />}
            onClick={loadAlerts}
            loading={loading}
          >
            Refresh
          </Button>
        </Space>
      </div>

      {/* 统计卡片 */}
      <Space style={{ marginBottom: 16, width: '100%' }} size="large">
        <Card size="small">
          <Space>
            <CloseCircleOutlined style={{ fontSize: 24, color: '#ff4d4f' }} />
            <div>
              <div style={{ fontSize: 12, color: '#666' }}>Active</div>
              <div style={{ fontSize: 20, fontWeight: 'bold' }}>{activeCount}</div>
            </div>
          </Space>
        </Card>
        <Card size="small">
          <Space>
            <WarningOutlined style={{ fontSize: 24, color: '#faad14' }} />
            <div>
              <div style={{ fontSize: 12, color: '#666' }}>Acknowledged</div>
              <div style={{ fontSize: 20, fontWeight: 'bold' }}>{acknowledgedCount}</div>
            </div>
          </Space>
        </Card>
        <Card size="small">
          <Space>
            <CheckCircleOutlined style={{ fontSize: 24, color: '#52c41a' }} />
            <div>
              <div style={{ fontSize: 12, color: '#666' }}>Resolved</div>
              <div style={{ fontSize: 20, fontWeight: 'bold' }}>{resolvedCount}</div>
            </div>
          </Space>
        </Card>
      </Space>

      {/* 筛选器 */}
      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
          <Search
            placeholder="Search alerts..."
            allowClear
            style={{ width: 300 }}
            onChange={(e) => setSearchText(e.target.value)}
            prefix={<SearchOutlined />}
          />
          <Select
            style={{ width: 150 }}
            value={severityFilter}
            onChange={setSeverityFilter}
            options={[
              { label: 'All Severities', value: 'all' },
              { label: 'Critical', value: 'critical' },
              { label: 'Error', value: 'error' },
              { label: 'Warning', value: 'warning' },
              { label: 'Info', value: 'info' },
            ]}
          />
          <Select
            style={{ width: 150 }}
            value={statusFilter}
            onChange={setStatusFilter}
            options={[
              { label: 'All Status', value: 'all' },
              { label: 'Active', value: 'active' },
              { label: 'Acknowledged', value: 'acknowledged' },
              { label: 'Resolved', value: 'resolved' },
            ]}
          />
        </Space>
      </Card>

      {/* 告警表格 */}
      <Card>
        <Table
          columns={columns}
          dataSource={filteredAlerts}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 20,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} alerts`,
          }}
          scroll={{ x: 1200 }}
        />
      </Card>
    </div>
  );
};

export default Alerts;
