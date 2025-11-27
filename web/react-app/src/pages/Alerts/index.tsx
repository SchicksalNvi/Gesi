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
import { useWebSocket } from '@/hooks/useWebSocket';
import type { ColumnsType } from 'antd/es/table';

const { Search } = Input;

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
  const [statusFilter, setStatusFilter] = useState<string>('active'); // Default to active only

  // WebSocket for real-time updates
  useWebSocket((event) => {
    if (event.type === 'alert_created' || event.type === 'alert_updated' || event.type === 'alert_resolved') {
      loadAlerts();
    }
  });

  useEffect(() => {
    loadAlerts();
  }, [severityFilter, statusFilter, searchText]);

  const loadAlerts = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (severityFilter !== 'all') params.append('severity', severityFilter);
      if (statusFilter !== 'all') params.append('status', statusFilter);
      if (searchText) params.append('search', searchText);
      
      const response = await fetch(`/api/alerts?${params.toString()}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (!response.ok) {
        throw new Error('Failed to fetch alerts');
      }
      
      const data = await response.json();
      setAlerts(data.data || []);
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

  const handleAcknowledge = async (alertId: number) => {
    try {
      const response = await fetch(`/api/alerts/${alertId}/acknowledge`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
      });
      
      if (!response.ok) {
        throw new Error('Failed to acknowledge alert');
      }
      
      message.success('Alert acknowledged successfully');
      loadAlerts();
    } catch (error) {
      console.error('Failed to acknowledge alert:', error);
      message.error('Failed to acknowledge alert');
    }
  };

  const handleResolve = async (alertId: number) => {
    try {
      const response = await fetch(`/api/alerts/${alertId}/resolve`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
      });
      
      if (!response.ok) {
        throw new Error('Failed to resolve alert');
      }
      
      message.success('Alert resolved successfully');
      loadAlerts();
    } catch (error) {
      console.error('Failed to resolve alert:', error);
      message.error('Failed to resolve alert');
    }
  };

  // Filter alerts
  const filtered = alerts.filter(alert => {
    const matchesSearch = 
      alert.node_name.toLowerCase().includes(searchText.toLowerCase()) ||
      (alert.process_name && alert.process_name.toLowerCase().includes(searchText.toLowerCase())) ||
      alert.message.toLowerCase().includes(searchText.toLowerCase());
    
    const matchesSeverity = severityFilter === 'all' || alert.severity === severityFilter;
    const matchesStatus = statusFilter === 'all' || alert.status === statusFilter;
    
    return matchesSearch && matchesSeverity && matchesStatus;
  });

  // Aggregate alerts: for resolved alerts, only show the latest one per node+rule
  const filteredAlerts = statusFilter === 'resolved' || statusFilter === 'all'
    ? (() => {
        const grouped = new Map<string, Alert>();
        filtered.forEach(alert => {
          const key = `${alert.node_name}-${alert.alert_type || 'unknown'}`;
          const existing = grouped.get(key);
          if (!existing || new Date(alert.created_at) > new Date(existing.created_at)) {
            grouped.set(key, alert);
          }
        });
        return Array.from(grouped.values()).sort((a, b) => 
          new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
        );
      })()
    : filtered;

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
      dataIndex: 'start_time',
      key: 'start_time',
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
            style={{ width: 180 }}
            value={statusFilter}
            onChange={setStatusFilter}
            options={[
              { label: 'Active Only', value: 'active' },
              { label: 'Acknowledged', value: 'acknowledged' },
              { label: 'Resolved', value: 'resolved' },
              { label: 'All Status', value: 'all' },
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
