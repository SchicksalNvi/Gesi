import { useEffect, useState, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Table,
  Button,
  Space,
  Spin,
  Tag,
  message,
  Result,
  Radio,
} from 'antd';
import {
  ArrowLeftOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import type { RadioChangeEvent } from 'antd';
import { environmentsApi } from '@/api/environments';
import { useWebSocket } from '@/hooks/useWebSocket';
import { EnvironmentDetail as EnvironmentDetailType, NodeDetail } from '@/types';
import type { ColumnsType } from 'antd/es/table';

type StatusFilter = 'all' | 'online' | 'offline';

export default function EnvironmentDetail() {
  const { environmentName } = useParams<{ environmentName: string }>();
  const navigate = useNavigate();
  const [environment, setEnvironment] = useState<EnvironmentDetailType | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');

  useEffect(() => {
    if (environmentName) {
      loadEnvironmentDetail();
    }
  }, [environmentName]);

  const loadEnvironmentDetail = async () => {
    if (!environmentName) return;

    setLoading(true);
    setError(null);
    try {
      const response = await environmentsApi.getEnvironmentDetail(environmentName);
      if (response.environment) {
        setEnvironment(response.environment);
      } else {
        setError('Environment not found');
      }
    } catch (err: any) {
      console.error('Failed to load environment detail:', err);
      if (err.response?.status === 404) {
        setError('Environment not found');
      } else {
        message.error('Failed to load environment details. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  // WebSocket real-time updates
  useWebSocket({
    onMessage: (msg) => {
      if (msg.type === 'nodes_update' || msg.type === 'node_status_change') {
        loadEnvironmentDetail();
      }
    },
  });

  // Filter nodes based on status
  const filteredNodes = useMemo(() => {
    if (!environment) return [];
    
    switch (statusFilter) {
      case 'online':
        return environment.members.filter((node) => node.is_connected);
      case 'offline':
        return environment.members.filter((node) => !node.is_connected);
      default:
        return environment.members;
    }
  }, [environment, statusFilter]);

  const handleStatusFilterChange = (e: RadioChangeEvent) => {
    setStatusFilter(e.target.value);
  };

  const columns: ColumnsType<NodeDetail> = [
    {
      title: 'Node Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <a
          onClick={(e) => {
            e.preventDefault();
            navigate(`/nodes/${name}`);
          }}
          style={{ fontWeight: 500 }}
        >
          {name}
        </a>
      ),
    },
    {
      title: 'Host',
      dataIndex: 'host',
      key: 'host',
    },
    {
      title: 'Port',
      dataIndex: 'port',
      key: 'port',
    },
    {
      title: 'Status',
      dataIndex: 'is_connected',
      key: 'status',
      render: (isConnected: boolean) => (
        <Tag
          icon={isConnected ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
          color={isConnected ? 'success' : 'error'}
        >
          {isConnected ? 'Online' : 'Offline'}
        </Tag>
      ),
    },
    {
      title: 'Processes',
      dataIndex: 'processes',
      key: 'processes',
      render: (count: number) => count || 0,
    },
    {
      title: 'Last Ping',
      dataIndex: 'last_ping',
      key: 'last_ping',
      render: (lastPing: string) => {
        if (!lastPing) return '-';
        const date = new Date(lastPing);
        return date.toLocaleString();
      },
    },
  ];

  if (loading && !environment) {
    return (
      <div style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <Result
        status="404"
        title="Environment Not Found"
        subTitle={`The environment "${environmentName}" does not exist.`}
        extra={
          <Button type="primary" onClick={() => navigate('/environments')}>
            Back to Environments
          </Button>
        }
      />
    );
  }

  if (!environment) {
    return null;
  }

  return (
    <div>
      <div
        style={{
          marginBottom: 24,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <div>
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/environments')}
            >
              Back
            </Button>
            <h2 style={{ margin: 0, fontSize: 24 }}>
              Environment: {environment.name}
            </h2>
          </Space>
          <p style={{ color: '#666', marginTop: 8, marginLeft: 40 }}>
            {environment.members.length} {environment.members.length === 1 ? 'node' : 'nodes'} in this environment
          </p>
        </div>
        <Button
          type="primary"
          icon={<ReloadOutlined />}
          onClick={loadEnvironmentDetail}
          loading={loading}
        >
          Refresh
        </Button>
      </div>

      <Card>
        {environment.members.length === 0 ? (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <p style={{ color: '#999', fontSize: 16 }}>
              No nodes found in this environment
            </p>
          </div>
        ) : (
          <>
            <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Radio.Group value={statusFilter} onChange={handleStatusFilterChange}>
                <Radio.Button value="all">
                  All ({environment.members.length})
                </Radio.Button>
                <Radio.Button value="online">
                  <CheckCircleOutlined style={{ color: '#52c41a' }} /> Online ({environment.members.filter(n => n.is_connected).length})
                </Radio.Button>
                <Radio.Button value="offline">
                  <CloseCircleOutlined style={{ color: '#ff4d4f' }} /> Offline ({environment.members.filter(n => !n.is_connected).length})
                </Radio.Button>
              </Radio.Group>
            </div>
            <Table
              columns={columns}
              dataSource={filteredNodes}
              rowKey="name"
              pagination={false}
              loading={loading}
            />
          </>
        )}
      </Card>
    </div>
  );
}
